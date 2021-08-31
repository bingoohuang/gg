package sqx

import (
	"database/sql"
	"github.com/bingoohuang/gg/pkg/sqlparse/sqlparser"
	"go.uber.org/multierr"
	"log"
	"reflect"
)

type Sqx struct {
	*sql.DB
	sqlparser.DBType
	dbExec ExecFn
}

func LogSqlResultDesc(desc string, lastResult sql.Result) {
	lastInsertId, _ := lastResult.LastInsertId()
	rowsAffected, _ := lastResult.RowsAffected()
	if desc == "" {
		log.Printf("Result lastInsertId: %d, rowsAffected: %d", lastInsertId, rowsAffected)
	} else {
		log.Printf("%s result lastInsertId: %d, rowsAffected: %d", desc, lastInsertId, rowsAffected)
	}
}
func logQueryError(desc string, result sql.Result, err error) {
	if desc == "" {
		if err != nil {
			log.Printf("query error: %v", err)
		} else if result != nil {
			LogSqlResultDesc(desc, result)
		}
		return
	}

	if err != nil {
		log.Printf("[%s] query error: %v", desc, err)
	} else if result != nil {
		LogSqlResultDesc(desc, result)
	}
}

func logRows(desc string, rows int) {
	if desc == "" {
		log.Printf("query %d rows", rows)
	} else {
		log.Printf("[%s] query %d rows", desc, rows)
	}
}
func logQuery(desc, query string, args []interface{}) {
	if desc == "" {
		log.Printf("query [%s] with args: %v", query, args)
	} else {
		log.Printf("[%s] query [%s] with args: %v", desc, query, args)
	}
}

type Executable interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type ExecFn func(query string, args ...interface{}) (sql.Result, error)

func (f ExecFn) Exec(query string, args ...interface{}) (sql.Result, error) { return f(query, args...) }

func wrapExec(dbType sqlparser.DBType, convertOptions []sqlparser.ConvertOption, f ExecFn) ExecFn {
	return func(query string, args ...interface{}) (sql.Result, error) {
		qq, cr, err := dbType.Convert(query, convertOptions...)
		if err != nil {
			return nil, err
		}
		args = cr.PickArgs(args)

		logQuery(qq, "", args)
		result, err := f(qq, args...)
		logQueryError("", result, err)
		return result, err
	}
}

func NewSqx(db *sql.DB) *Sqx {
	dbType := sqlparser.ToDBType(DriverName(db))
	return &Sqx{DB: db, DBType: dbType, dbExec: wrapExec(dbType, nil, db.Exec)}
}

type QueryArgs struct {
	Desc    string
	Dest    interface{}
	Query   string
	Args    []interface{}
	Limit   int
	Options []sqlparser.ConvertOption
}

func (a *QueryArgs) GetQueryRows() int {
	if a.Dest == nil {
		return 0
	}

	v := reflect.ValueOf(a.Dest)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return v.Len()
	default:
		return 0
	}
}

func (s *Sqx) Query(arg *QueryArgs) error {
	options := arg.Options
	if arg.Limit > 0 {
		options = append([]sqlparser.ConvertOption{sqlparser.WithLimit(arg.Limit)}, options...)
	}
	qq, cr, err := s.DBType.Convert(arg.Query, options...)
	if err != nil {
		return err
	}

	args := cr.PickArgs(arg.Args)
	logQuery(arg.Desc, qq, args)

	err = NewSQL(qq, args...).Query(s.DB, arg.Dest)
	logQueryError(arg.Desc, nil, err)
	logRows(arg.Desc, arg.GetQueryRows())

	return err
}

func (s *Sqx) SelectDesc(desc string, dest interface{}, query string, args ...interface{}) error {
	return s.Query(&QueryArgs{Desc: desc, Dest: dest, Query: query, Args: args})
}

func (s *Sqx) Select(dest interface{}, query string, args ...interface{}) error {
	return s.Query(&QueryArgs{Dest: dest, Query: query, Args: args})
}

func (s *Sqx) GetDesc(desc string, dest interface{}, query string, args ...interface{}) error {
	return s.Query(&QueryArgs{Desc: desc, Dest: dest, Query: query, Args: args, Limit: 1})
}

func (s *Sqx) Get(dest interface{}, query string, args ...interface{}) error {
	return s.Query(&QueryArgs{Dest: dest, Query: query, Args: args, Limit: 1})
}

func (s *Sqx) Upsert(insertQuery, updateQuery string, args ...interface{}) (ur UpsertResult, err error) {
	return Upsert(s, insertQuery, updateQuery, args...)
}

func (s *Sqx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.dbExec(query, args...)
}

func (s *Sqx) Tx(f func(txExec ExecFn) error) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}

	txExec := wrapExec(s.DBType, nil, tx.Exec)
	if err := f(txExec); err != nil {
		err2 := tx.Rollback()
		return multierr.Append(err, err2)
	}

	return tx.Commit()
}

func Args(keys []string) []interface{} {
	args := make([]interface{}, len(keys))
	for i := 0; i < len(keys); i++ {
		args[i] = keys[i]
	}

	return args
}

type UpsertResult int

const (
	UpsertError UpsertResult = iota
	UpsertInserted
	UpsertUpdated
)

func Upsert(executable Executable, insertQuery, updateQuery string, args ...interface{}) (ur UpsertResult, err error) {
	_, err1 := executable.Exec(insertQuery, args...)
	if err1 == nil {
		return UpsertInserted, nil
	}

	_, err2 := executable.Exec(updateQuery, args...)
	if err2 == nil {
		return UpsertUpdated, nil
	}

	return UpsertError, multierr.Append(err1, err2)
}