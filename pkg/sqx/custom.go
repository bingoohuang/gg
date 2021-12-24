package sqx

import (
	"reflect"
)

type CustomDriverValueConverter interface {
	Convert(value interface{}) (interface{}, error)
}

type CustomDriverValueConvertFn func(value interface{}) (interface{}, error)

func (fn CustomDriverValueConvertFn) Convert(value interface{}) (interface{}, error) {
	return fn(value)
}

var CustomDriverValueConverters = map[reflect.Type]CustomDriverValueConverter{}
