package sqlkv

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bingoohuang/gg/pkg/sqx"
	"go.uber.org/multierr"
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
	DefaultGetSQL    = `select v from kv where k = :k and state = 1`
	DefaultInsertSQL = `insert into kv(k, v, state, created) values(:k, :v, 1, :time)`
	DefaultUpdateSQL = `update kv set v = :v, updated = :time, state = 1 where k = :k`
	DefaultDelSQL    = `update kv set state = 0  where k = :k`
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

	go cli.tickerRefresh()

	return cli, nil
}

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
	_, dbx, err := sqx.Open(c.DriverName, c.DataSourceName)
	if err != nil {
		return nil, err
	}

	defer func() { er = multierr.Append(er, dbx.Close()) }()

	kvs = make(map[string]string)
	sqx.SQL{Q: c.AllSQL}.QueryRaw(dbx, sqx.WithScanRow(func(_ []string, rows *sql.Rows, _ int) (bool, error) {
		rowValues, err := sqx.ScanRowValues(rows)
		if err != nil {
			return false, err
		}

		k := fmt.Sprintf("%v", rowValues[0])
		kvs[k] = fmt.Sprintf("%v", rowValues[1])
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
	m := map[string]string{
		"k":    k,
		"v":    v,
		"time": time.Now().Format(`2006-01-02 15:04:05.000`),
	}

	_, dbx, err := sqx.Open(c.DriverName, c.DataSourceName)
	if err != nil {
		return err
	}

	defer func() { er = multierr.Append(er, dbx.Close()) }()

	effectedRows, err := sqx.SQL{Q: c.UpdateSQL, Vars: sqx.Vars(m)}.Update(dbx)
	if err == nil && effectedRows > 0 {
		c.set(k, v)
		return nil
	}

	_, err = sqx.SQL{Q: c.InsertSQL, Vars: sqx.Vars(m)}.Update(dbx)
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

	_, dbx, err := sqx.Open(c.DriverName, c.DataSourceName)
	if err != nil {
		return false, "", err
	}

	defer func() { er = multierr.Append(er, dbx.Close()) }()

	m := map[string]string{"k": k}
	v, err = sqx.SQL{Q: c.GetSQL, Vars: sqx.Vars(m)}.QueryAsString(dbx)
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
	m := map[string]string{
		"k":    k,
		"time": time.Now().Format(`2006-01-02 15:04:05.000`),
	}
	_, db, err := sqx.Open(c.DriverName, c.DataSourceName)
	if err != nil {
		return err
	}

	defer func() { er = multierr.Append(er, db.Close()) }()

	if _, err := (sqx.SQL{Q: c.DelSQL, Vars: sqx.Vars(m)}).Update(db); err != nil {
		return err
	}

	return nil
}
