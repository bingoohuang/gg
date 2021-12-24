package sqlkv

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/bingoohuang/gg/pkg/sqx"
	"go.uber.org/multierr"
	"log"
	"sync"
	"text/template"
	"time"
)

type Config struct {
	DriverName     string
	DataSourceName string

	AllSQL    string
	GetSQL    string
	UpdateSQL string
	InsertSQL string
	DelSQL    string

	// RefreshInterval will Refresh the key values from the database in every Refresh interval.
	RefreshInterval time.Duration
}

// Client is a gokv.Store implementation for SQL databases.
type Client struct {
	Config

	cache     map[string]string
	cacheLock sync.Mutex

	allSql    *template.Template
	delSql    *template.Template
	getSql    *template.Template
	updateSQL *template.Template
	insertSQL *template.Template
}

func DefaultDuration(s, defaultValue time.Duration) time.Duration {
	if s == 0 {
		return defaultValue
	}
	return s
}
func Default(s, defaultValue string) string {
	if s == "" {
		return defaultValue
	}

	return s
}

const (
	DefaultAllSQL    = `select k,v from kv where state = 1`
	DefaultGetSQL    = `select v from kv where k = '{{.Key}}' and state = 1`
	DefaultInsertSQL = `insert into kv(k, v, state, created) values('{{.Key}}', '{{.Value}}', 1, '{{.Time}}') `
	DefaultUpdateSQL = `update kv set v = '{{.Value}}', updated = '{{.Time}}', state = 1 where k = '{{.Key}}'`
	DefaultDelSQL    = `update kv set state = 0  where k = '{{.Key}}'`
)

func NewClient(c Config) (*Client, error) {
	c.RefreshInterval = DefaultDuration(c.RefreshInterval, 60*time.Second)
	c.DriverName = Default(c.DriverName, "mysql")
	c.AllSQL = Default(c.AllSQL, DefaultAllSQL)
	c.GetSQL = Default(c.GetSQL, DefaultGetSQL)
	c.UpdateSQL = Default(c.UpdateSQL, DefaultUpdateSQL)
	c.InsertSQL = Default(c.InsertSQL, DefaultInsertSQL)
	c.DelSQL = Default(c.DelSQL, DefaultDelSQL)

	cli := &Client{
		Config: c,
		cache:  make(map[string]string),
	}

	var err error
	if cli.allSql, err = template.New("").Parse(c.AllSQL); err != nil {
		return nil, err
	}
	if cli.delSql, err = template.New("").Parse(c.DelSQL); err != nil {
		return nil, err
	}
	if cli.getSql, err = template.New("").Parse(c.GetSQL); err != nil {
		return nil, err
	}
	if cli.updateSQL, err = template.New("").Parse(c.UpdateSQL); err != nil {
		return nil, err
	}
	if cli.insertSQL, err = template.New("").Parse(c.InsertSQL); err != nil {
		return nil, err
	}

	go cli.tickerRefresh()

	return cli, nil
}

var (
	// ErrTooManyValues is the error to identify more than one values associated with a key.
	ErrTooManyValues = errors.New("more than one values associated with the key")
)

func (c *Client) tickerRefresh() {
	ticker := time.NewTicker(c.RefreshInterval)
	for range ticker.C {
		if _, err := c.All(); err != nil {
			log.Printf("W! refersh error %v", err)
		}
	}
}

// All list the keys in the store.
func (c *Client) All() (kvs map[string]string, er error) {
	var out bytes.Buffer
	if err := c.allSql.Execute(&out, map[string]string{}); err != nil {
		return nil, err
	}
	query := out.String()
	log.Printf("D! query: %s", query)

	dbx, err := sqx.OpenSqx(c.DriverName, c.DataSourceName)
	if err != nil {
		return nil, err
	}

	defer func() { er = multierr.Append(er, dbx.Close()) }()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	kvs = make(map[string]string)
	sqx.SQL{Q: query, Ctx: ctx}.QueryRaw(dbx, sqx.WithScanRow(func(cols []string, rows *sql.Rows, _ int) (bool, error) {
		rowValues, err := sqx.ScanRowValues(rows)
		if err != nil {
			return false, err
		}

		kvs[cols[0]] = fmt.Sprintf("%v", rowValues[1])
		return true, nil
	}))

	c.cacheLock.Lock()
	c.cache = kvs
	c.cacheLock.Unlock()

	return kvs, nil
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c *Client) Set(k, v string) (er error) {
	var out bytes.Buffer
	m := map[string]string{
		"Key":   k,
		"Value": v,
		"Time":  time.Now().Format(`2006-01-02 15:04:05.000`),
	}
	if err := c.updateSQL.Execute(&out, m); err != nil {
		return err
	}

	query := out.String()
	log.Printf("D! query: %s", query)

	dbx, err := sqx.OpenSqx(c.DriverName, c.DataSourceName)
	if err != nil {
		return err
	}

	defer func() { er = multierr.Append(er, dbx.Close()) }()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	effectedRows, err := sqx.SQL{Q: query, Ctx: ctx}.Update(dbx)
	if err == nil && effectedRows > 0 {
		c.set(k, v)
		return nil
	}

	out.Reset()
	if err := c.insertSQL.Execute(&out, m); err != nil {
		return err
	}

	query = out.String()
	log.Printf("D! query: %s", query)
	_, err = dbx.Exec(query)
	return err
}

func (c *Client) set(k, v string) {
	c.cacheLock.Lock()
	c.cache[k] = v
	c.cacheLock.Unlock()
}

// Get retrieves the stored value for the given key.
// You need to pass a pointer to the value, so in case of a struct
// the automatic unmarshalling can populate the fields of the object
// that v points to with the values of the retrieved object's values.
// If no value is found it returns (false, nil).
func (c *Client) Get(k string) (found bool, v string, er error) {
	c.cacheLock.Lock()
	if v, ok := c.cache[k]; ok {
		c.cacheLock.Unlock()
		return true, v, nil
	}
	c.cacheLock.Unlock()

	t, err := template.New("").Parse(c.GetSQL)
	if err != nil {
		return false, "", err
	}

	var out bytes.Buffer
	if err := t.Execute(&out, map[string]string{"Key": k}); err != nil {
		return false, "", err
	}

	query := out.String()
	log.Printf("D! query: %s", query)

	dbx, err := sqx.OpenSqx(c.DriverName, c.DataSourceName)
	if err != nil {
		return false, "", err
	}

	defer func() { er = multierr.Append(er, dbx.Close()) }()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	v, err = sqx.SQL{Q: query, Ctx: ctx}.QueryAsString(dbx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
		}
		return false, "", err
	}

	c.set(k, v)
	return true, v, nil
}

// Del deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (c *Client) Del(k string) (er error) {
	c.cacheLock.Lock()
	delete(c.cache, k)
	c.cacheLock.Unlock()

	defer func() {
		if err := c.del(k); err != nil {
			log.Printf("W! failed to del %v", err)
		}
	}()

	return nil

}
func (c *Client) del(k string) (er error) {
	var out bytes.Buffer
	if err := c.delSql.Execute(&out, map[string]string{
		"Key":  k,
		"Time": time.Now().Format(`2006-01-02 15:04:05.000`),
	}); err != nil {
		return err
	}

	query := out.String()
	log.Printf("D! query: %s", query)

	db, err := sql.Open(c.DriverName, c.DataSourceName)
	if err != nil {
		return err
	}

	defer func() { er = multierr.Append(er, db.Close()) }()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, query); err != nil {
		return err
	}

	return nil
}
