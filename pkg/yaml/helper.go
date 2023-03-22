package yaml

import (
	"reflect"
	"strconv"

	"golang.org/x/xerrors"
)

// CastUint64 casts an uint64 to target typ.
func CastUint64(v uint64, typ reflect.Type) (reflect.Value, error) {
	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		vv := reflect.New(typ)
		vv.Elem().SetInt(int64(v))
		return vv.Elem(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		vv := reflect.New(typ)
		vv.Elem().SetUint(v)
		return vv.Elem(), nil
	case reflect.String:
		vv := reflect.New(typ)
		vv.Elem().SetString(strconv.FormatUint(v, 10))
		return vv.Elem(), nil
	}
	return reflect.Value{}, xerrors.New("failed to cast")
}
