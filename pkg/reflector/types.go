package reflector

import (
	"reflect"
	"time"
)

var (
	TimeType = reflect.TypeOf((*time.Time)(nil)).Elem()
	// ErrType defines the error's type
	// 参考 https://github.com/uber-go/dig/blob/master/types.go
	ErrType = reflect.TypeOf((*error)(nil)).Elem()
)
