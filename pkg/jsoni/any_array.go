package jsoni

import (
	"reflect"
	"unsafe"
)

type arrayLazyAny struct {
	baseAny
	cfg *frozenConfig
	buf []byte
	err error
}

func (a *arrayLazyAny) ValueType() ValueType { return ArrayValue }
func (a *arrayLazyAny) MustBeValid() Any     { return a }
func (a *arrayLazyAny) LastError() error     { return a.err }
func (a *arrayLazyAny) ToBool() bool {
	iter := a.cfg.BorrowIterator(a.buf)
	defer a.cfg.ReturnIterator(iter)
	return iter.ReadArray()
}

func (a *arrayLazyAny) ToInt() int {
	if a.ToBool() {
		return 1
	}
	return 0
}

func (a *arrayLazyAny) ToInt32() int32     { return int32(a.ToInt()) }
func (a *arrayLazyAny) ToInt64() int64     { return int64(a.ToInt()) }
func (a *arrayLazyAny) ToUint() uint       { return uint(a.ToInt()) }
func (a *arrayLazyAny) ToUint32() uint32   { return uint32(a.ToInt()) }
func (a *arrayLazyAny) ToUint64() uint64   { return uint64(a.ToInt()) }
func (a *arrayLazyAny) ToFloat32() float32 { return float32(a.ToInt()) }
func (a *arrayLazyAny) ToFloat64() float64 { return float64(a.ToInt()) }
func (a *arrayLazyAny) ToString() string   { return *(*string)(unsafe.Pointer(&a.buf)) }

func (a *arrayLazyAny) ToVal(val interface{}) {
	iter := a.cfg.BorrowIterator(a.buf)
	defer a.cfg.ReturnIterator(iter)
	iter.ReadVal(val)
}

func (a *arrayLazyAny) Get(path ...interface{}) Any {
	if len(path) == 0 {
		return a
	}
	switch firstPath := path[0].(type) {
	case int:
		iter := a.cfg.BorrowIterator(a.buf)
		defer a.cfg.ReturnIterator(iter)
		valueBytes := locateArrayElement(iter, firstPath)
		if valueBytes == nil {
			return newInvalidAny(path)
		}
		iter.ResetBytes(valueBytes)
		return locatePath(iter, path[1:])
	case int32:
		if '*' == firstPath {
			iter := a.cfg.BorrowIterator(a.buf)
			defer a.cfg.ReturnIterator(iter)
			arr := make([]Any, 0)
			iter.ReadArrayCB(func(iter *Iterator) bool {
				found := iter.readAny().Get(path[1:]...)
				if found.ValueType() != InvalidValue {
					arr = append(arr, found)
				}
				return true
			})
			return wrapArray(arr)
		}
		return newInvalidAny(path)
	default:
		return newInvalidAny(path)
	}
}

func (a *arrayLazyAny) Size() int {
	size := 0
	iter := a.cfg.BorrowIterator(a.buf)
	defer a.cfg.ReturnIterator(iter)
	iter.ReadArrayCB(func(iter *Iterator) bool {
		size++
		iter.Skip()
		return true
	})
	return size
}

func (a *arrayLazyAny) WriteTo(stream *Stream) { stream.Write(a.buf) }

func (a *arrayLazyAny) GetInterface() interface{} {
	iter := a.cfg.BorrowIterator(a.buf)
	defer a.cfg.ReturnIterator(iter)
	return iter.Read()
}

type arrayAny struct {
	baseAny
	val reflect.Value
}

func wrapArray(val interface{}) *arrayAny {
	return &arrayAny{baseAny{}, reflect.ValueOf(val)}
}

func (a *arrayAny) ValueType() ValueType { return ArrayValue }
func (a *arrayAny) MustBeValid() Any     { return a }
func (a *arrayAny) LastError() error     { return nil }
func (a *arrayAny) ToBool() bool         { return a.val.Len() != 0 }
func (a *arrayAny) ToInt() int {
	if a.val.Len() == 0 {
		return 0
	}
	return 1
}

func (a *arrayAny) ToInt32() int32     { return int32(a.ToInt()) }
func (a *arrayAny) ToInt64() int64     { return int64(a.ToInt()) }
func (a *arrayAny) ToUint() uint       { return uint(a.ToInt()) }
func (a *arrayAny) ToUint32() uint32   { return uint32(a.ToInt()) }
func (a *arrayAny) ToUint64() uint64   { return uint64(a.ToInt()) }
func (a *arrayAny) ToFloat32() float32 { return float32(a.ToInt()) }
func (a *arrayAny) ToFloat64() float64 { return float64(a.ToInt()) }

func (a *arrayAny) ToString() string {
	str, _ := MarshalToString(a.val.Interface())
	return str
}

func (a *arrayAny) Get(path ...interface{}) Any {
	if len(path) == 0 {
		return a
	}
	switch firstPath := path[0].(type) {
	case int:
		if firstPath < 0 || firstPath >= a.val.Len() {
			return newInvalidAny(path)
		}
		return Wrap(a.val.Index(firstPath).Interface())
	case int32:
		if '*' == firstPath {
			mappedAll := make([]Any, 0)
			for i := 0; i < a.val.Len(); i++ {
				mapped := Wrap(a.val.Index(i).Interface()).Get(path[1:]...)
				if mapped.ValueType() != InvalidValue {
					mappedAll = append(mappedAll, mapped)
				}
			}
			return wrapArray(mappedAll)
		}
		return newInvalidAny(path)
	default:
		return newInvalidAny(path)
	}
}

func (a *arrayAny) Size() int                 { return a.val.Len() }
func (a *arrayAny) WriteTo(stream *Stream)    { stream.WriteVal(a.val) }
func (a *arrayAny) GetInterface() interface{} { return a.val.Interface() }
