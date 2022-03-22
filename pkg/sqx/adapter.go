package sqx

import (
	"context"
	"database/sql"
	"github.com/bingoohuang/gg/pkg/sqlparse/sqlparser"
	"go.uber.org/multierr"
	"log"
	"strings"
)

type DBTypeAware interface {
	GetDBType() sqlparser.DBType
}

func (s Sqx) GetDBType() sqlparser.DBType { return s.DBType }

type Sqx struct {
	DB      SqxDB
	DBType  sqlparser.DBType
	CloseFn func() error
}

func LogSqlResultDesc(desc string, lastResult sql.Result) {
	lastInsertId, _ := lastResult.LastInsertId()
	rowsAffected, _ := lastResult.RowsAffected()
	log.Printf("%sresult lastInsertId: %d, rowsAffected: %d", quoteDesc(desc), lastInsertId, rowsAffected)
}

func quoteDesc(desc string) string {
	if desc != "" && !strings.HasPrefix(desc, "[") {
		desc = "[" + desc + "] "
	}
	return desc
}

func logQueryError(nolog bool, desc string, result sql.Result, err error) {
	if nolog {
		return
	}

	if err != nil {
		log.Printf("%squery error: %v", quoteDesc(desc), err)
	} else if result != nil {
		LogSqlResultDesc(desc, result)
	}
}

func logRows(desc string, rows int) {
	log.Printf("%squery %d rows", quoteDesc(desc), rows)
}

func logQuery(desc, query string, args []interface{}) {
	log.Printf("%squery [%s] with args: %v", quoteDesc(desc), query, args)
}

type Executable interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type ExecFn func(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

func (f ExecFn) Exec(query string, args ...interface{}) (sql.Result, error) {
	return f(context.Background(), query, args...)
}
func (f ExecFn) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return f(ctx, query, args...)
}

type Queryable interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type QueryFn func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)

func (f QueryFn) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return f(context.Background(), query, args...)
}
func (f QueryFn) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return f(ctx, query, args...)
}

func (s *Sqx) Close() error {
	if s.CloseFn != nil {
		return s.CloseFn()
	}

	return nil
}
func (s *Sqx) DoQuery(arg *QueryArgs) error {
	return arg.DoQuery(s.typedDB())
}

func (s *Sqx) DoExec(arg *QueryArgs) (int64, error) {
	return arg.DoExec(s.typedDB())
}

func (s *Sqx) DoExecRaw(arg *QueryArgs) (sql.Result, error) {
	return arg.DoExecRaw(s.typedDB())
}

func (a *QueryArgs) DoExecRaw(db SqxDB) (sql.Result, error) {
	ns := NewSQL(a.Query, a.Args...)
	ns.Ctx = a.Ctx
	return ns.WithConvertOptions(a.Options).UpdateRaw(db)
}

func (a *QueryArgs) DoExec(db SqxDB) (int64, error) {
	ns := NewSQL(a.Query, a.Args...)
	ns.Ctx = a.Ctx
	return ns.WithConvertOptions(a.Options).Update(db)
}

func (a *QueryArgs) DoQuery(db SqxDB) error {
	ns := NewSQL(a.Query, a.Args...)
	ns.Ctx = a.Ctx
	return ns.WithConvertOptions(a.Options).Query(db, a.Dest)
}

func NewSqx(db *sql.DB) *Sqx {
	return &Sqx{DB: db, DBType: sqlparser.ToDBType(DriverName(db.Driver())), CloseFn: db.Close}
}

type QueryArgs struct {
	Desc    string
	Dest    interface{}
	Query   string
	Args    []interface{}
	Limit   int
	Options []sqlparser.ConvertOption
	Ctx     context.Context
}

func (s *Sqx) SelectDesc(desc string, dest interface{}, query string, args ...interface{}) error {
	return s.DoQuery(&QueryArgs{Desc: desc, Dest: dest, Query: query, Args: args})
}

func (s *Sqx) Select(dest interface{}, query string, args ...interface{}) error {
	return s.DoQuery(&QueryArgs{Dest: dest, Query: query, Args: args})
}

func (s *Sqx) GetDesc(desc string, dest interface{}, query string, args ...interface{}) error {
	return s.DoQuery(&QueryArgs{Desc: desc, Dest: dest, Query: query, Args: args, Limit: 1})
}

func (s *Sqx) Get(dest interface{}, query string, args ...interface{}) error {
	return s.DoQuery(&QueryArgs{Dest: dest, Query: query, Args: args, Limit: 1})
}

func (s *Sqx) Upsert(insertQuery, updateQuery string, args ...interface{}) (ur UpsertResult, err error) {
	return Upsert(s, insertQuery, updateQuery, args...)
}

func (s *Sqx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return s.DoExecRaw(&QueryArgs{Ctx: ctx, Query: query, Args: args, Limit: 1})
}

func (s *Sqx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.ExecContext(context.Background(), query, args...)
}

func (s *Sqx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return s.QueryContext(context.Background(), query, args...)
}

type ContextKey int

const (
	AdaptedKey ContextKey = iota + 1
)

func (s *Sqx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	s2 := SQL{Ctx: ctx, Q: query, Vars: args}

	if adapted, ok := ctx.Value(AdaptedKey).(bool); !ok || !adapted {
		if err := s2.adaptQuery(s); err != nil {
			return nil, err
		}
	}

	rows, err := s.DB.QueryContext(ctx, s2.Q, s2.Vars...)
	if err != nil {
		logQueryError(false, "", nil, err)
	}
	return rows, err
}

type BeginTx interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

func (s *Sqx) Tx(f func(sqx *Sqx) error) error {
	return s.TxContext(context.Background(), f)
}

func (s *Sqx) TxContext(ctx context.Context, f func(sqx *Sqx) error) error {
	btx, ok := s.DB.(BeginTx)
	if !ok {
		panic("can't begin transaction")
	}

	tx, err := btx.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := f(&Sqx{DB: tx, DBType: s.DBType}); err != nil {
		return multierr.Append(err, tx.Rollback())
	}

	return tx.Commit()
}

type DBRaw struct {
	SqxDB
	DBType sqlparser.DBType
}

func (t DBRaw) GetDBType() sqlparser.DBType { return t.DBType }

func (s Sqx) typedDB() SqxDB {
	return &DBRaw{
		SqxDB:  s.DB,
		DBType: s.DBType,
	}
}

func VarsStr(keys ...string) []interface{} {
	args := make([]interface{}, len(keys))
	for i := 0; i < len(keys); i++ {
		args[i] = keys[i]
	}

	return args
}

func Vars(keys ...interface{}) []interface{} {
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
	return UpsertContext(context.Background(), executable, insertQuery, updateQuery, args...)
}

func UpsertContext(ctx context.Context, executable Executable, insertQuery, updateQuery string, args ...interface{}) (ur UpsertResult, err error) {
	_, err1 := executable.ExecContext(ctx, insertQuery, args...)
	if err1 == nil {
		return UpsertInserted, nil
	}

	_, err2 := executable.ExecContext(ctx, updateQuery, args...)
	if err2 == nil {
		return UpsertUpdated, nil
	}

	return UpsertError, multierr.Append(err1, err2)
}
