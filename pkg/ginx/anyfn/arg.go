package anyfn

import (
	"fmt"
	"reflect"

	"github.com/bingoohuang/gg/pkg/ginx/cast"
)

type ArgIn struct {
	Index int
	Type  reflect.Type
	Kind  reflect.Kind
	Ptr   bool
}

func parseArgIns(ft reflect.Type) []ArgIn {
	numIn := ft.NumIn()
	argIns := make([]ArgIn, numIn)

	for i := 0; i < numIn; i++ {
		argIns[i] = parseArgs(ft, i)
	}

	return argIns
}

func parseArgs(ft reflect.Type, argIndex int) ArgIn {
	argType := ft.In(argIndex)
	ptr := argType.Kind() == reflect.Ptr

	if ptr {
		argType = argType.Elem()
	}

	return ArgIn{Index: argIndex, Type: argType, Kind: argType.Kind(), Ptr: ptr}
}

func (arg ArgIn) convertValue(s string) (reflect.Value, error) {
	v, err := cast.To(s, arg.Type)
	if err != nil {
		return reflect.Value{}, &AdapterError{
			Err:     err,
			Context: fmt.Sprintf("To %s to %v", s, arg.Type),
		}
	}

	return ConvertPtr(arg.Ptr, v), nil
}
