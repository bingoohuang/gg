package jsoni

import (
	"io"
	"unsafe"
)

type numberLazyAny struct {
	baseAny
	cfg *frozenConfig
	buf []byte
	err error
}

func (a *numberLazyAny) ValueType() ValueType { return NumberValue }
func (a *numberLazyAny) MustBeValid() Any     { return a }
func (a *numberLazyAny) LastError() error     { return a.err }
func (a *numberLazyAny) ToBool() bool         { return a.ToFloat64() != 0 }

func (a *numberLazyAny) iterFunc(f func(*Iterator)) {
	i := a.cfg.BorrowIterator(a.buf)
	f(i)
	if i.Error != nil && i.Error != io.EOF {
		a.err = i.Error
	}

	a.cfg.ReturnIterator(i)
}

func (a *numberLazyAny) ToInt() (val int) {
	a.iterFunc(func(iter *Iterator) { val = iter.ReadInt() })
	return
}

func (a *numberLazyAny) ToInt32() (val int32) {
	a.iterFunc(func(i *Iterator) { val = i.ReadInt32() })
	return
}

func (a *numberLazyAny) ToInt64() (val int64) {
	a.iterFunc(func(i *Iterator) { val = i.ReadInt64() })
	return
}

func (a *numberLazyAny) ToUint() (val uint) {
	a.iterFunc(func(i *Iterator) { val = i.ReadUint() })
	return
}

func (a *numberLazyAny) ToUint32() (val uint32) {
	a.iterFunc(func(i *Iterator) { val = i.ReadUint32() })
	return
}

func (a *numberLazyAny) ToUint64() (val uint64) {
	a.iterFunc(func(i *Iterator) { val = i.ReadUint64() })
	return
}

func (a *numberLazyAny) ToFloat32() (val float32) {
	a.iterFunc(func(i *Iterator) { val = i.ReadFloat32() })
	return
}

func (a *numberLazyAny) ToFloat64() (val float64) {
	a.iterFunc(func(i *Iterator) { val = i.ReadFloat64() })
	return
}

func (a *numberLazyAny) ToString() string       { return *(*string)(unsafe.Pointer(&a.buf)) }
func (a *numberLazyAny) WriteTo(stream *Stream) { stream.Write(a.buf) }

func (a *numberLazyAny) GetInterface() (val interface{}) {
	a.iterFunc(func(i *Iterator) { val = i.Read() })
	return
}
