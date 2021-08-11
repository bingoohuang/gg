package sqx

import (
	"database/sql"
	"github.com/bingoohuang/gg/pkg/reflector"
	"reflect"
)

// 参考 https://github.com/uber-go/dig/blob/master/types.go
// nolint:gochecknoglobals
var (
	_sqlScannerType = reflect.TypeOf((*sql.Scanner)(nil)).Elem()
)

// ImplSQLScanner tells t whether it implements sql.Scanner interface.
func ImplSQLScanner(t reflect.Type) bool { return reflector.ImplType(t, _sqlScannerType) }
