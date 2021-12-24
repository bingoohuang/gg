package dm

import (
	"errors"
	"gitee.com/chunanyong/dm"
	"github.com/bingoohuang/gg/pkg/sqx"
	"reflect"
)

func ConvertDmClob(value interface{}) (interface{}, error) {
	clob, ok := value.(*dm.DmClob)
	if !ok {
		return value, errors.New("conversion to *dm.DmClob type failed")
	}

	length, err := clob.GetLength()
	if err != nil {
		return clob, err
	}

	s, err := clob.ReadString(1, int(length))
	return s, err
}

func init() {
	sqx.CustomDriverValueConverters[reflect.TypeOf((*dm.DmClob)(nil))] = sqx.CustomDriverValueConvertFn(ConvertDmClob)
}
