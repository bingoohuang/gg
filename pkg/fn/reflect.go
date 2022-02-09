package fn

import (
	"reflect"
	"runtime"
)

// GetFuncName returns the func name of a func.
func GetFuncName(i interface{}) string {
	// github.com/bingoohuang/gg/pkg/fn.GetFuncName
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
