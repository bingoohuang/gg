package sqx

import (
	"database/sql"
	"github.com/bingoohuang/gg/pkg/reflector"
	"log"
	"reflect"
)

// DaoLogger is the interface for dao logging.
type DaoLogger interface {
	// LogError logs the error
	LogError(err error)
	// LogStart logs the sql before the sql execution
	LogStart(id, sql string, vars interface{})
}

var (
	_daoLoggerType = reflect.TypeOf((*DaoLogger)(nil)).Elem()
	_dbGetterType  = reflect.TypeOf((*DBGetter)(nil)).Elem()
)

// DaoLog implements the interface for dao logging with standard log.
type DaoLog struct{}

// LogError logs the error.
func (d *DaoLog) LogError(err error) {
	log.Printf("E! error occurred %v", err)
}

// LogStart logs the sql before the sql execution.
func (d *DaoLog) LogStart(id, sql string, vars interface{}) {
	log.Printf("exec %s [%s] with %v", id, sql, vars)
}

// getDBFn is the function type to get a sql.DBGetter.
type getDBFn func() *sql.DB

// GetDB returns a sql.DBGetter.
func (f getDBFn) GetDB() *sql.DB { return f() }

func createDBGetter(v reflect.Value, option *CreateDaoOpt) {
	if option.DBGetter != nil {
		return
	}

	if fv := findTypedField(v, _dbGetterType); fv.IsValid() {
		option.DBGetter = fv.Interface().(DBGetter)
		return
	}

	option.DBGetter = getDBFn(func() *sql.DB { return DB })
}

func createLogger(v reflect.Value, option *CreateDaoOpt) {
	if option.Logger != nil {
		return
	}

	if fv := findTypedField(v, _daoLoggerType); fv.IsValid() {
		option.Logger = fv.Interface().(DaoLogger)
		return
	}

	option.Logger = &DaoLog{}
}

func findTypedField(v reflect.Value, t reflect.Type) reflect.Value {
	for i := 0; i < v.NumField(); i++ {
		f := v.Type().Field(i)

		if f.PkgPath != "" /* not exportable? */ {
			continue
		}

		fv := v.Field(i)
		if reflector.ImplType(f.Type, t) && !fv.IsNil() {
			return fv
		}
	}

	return reflect.Value{}
}
