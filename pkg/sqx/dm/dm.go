package dm

import (
	"errors"
	"gitee.com/chunanyong/dm"
	"github.com/bingoohuang/gg/pkg/sqx"
	"io"
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

	if length == 0 {
		return "", nil
	}

	s, err := clob.ReadString(1, int(length))
	if err != nil && errors.Is(err, io.EOF) {
		return "", nil
	}
	return s, err
}

func init() {
	sqx.CustomDriverValueConverters[reflect.TypeOf((*dm.DmClob)(nil))] = sqx.CustomDriverValueConvertFn(ConvertDmClob)
}
