package jsoni

import "fmt"

type invalidAny struct {
	baseAny
	err error
}

func newInvalidAny(path []interface{}) *invalidAny {
	return &invalidAny{baseAny{}, fmt.Errorf("%v not found", path)}
}

func (a *invalidAny) LastError() error     { return a.err }
func (a *invalidAny) ValueType() ValueType { return InvalidValue }
func (a *invalidAny) MustBeValid() Any     { panic(a.err) }
func (a *invalidAny) ToBool() bool         { return false }
func (a *invalidAny) ToInt() int           { return 0 }
func (a *invalidAny) ToInt32() int32       { return 0 }
func (a *invalidAny) ToInt64() int64       { return 0 }
func (a *invalidAny) ToUint() uint         { return 0 }
func (a *invalidAny) ToUint32() uint32     { return 0 }
func (a *invalidAny) ToUint64() uint64     { return 0 }
func (a *invalidAny) ToFloat32() float32   { return 0 }
func (a *invalidAny) ToFloat64() float64   { return 0 }
func (a *invalidAny) ToString() string     { return "" }
func (a *invalidAny) WriteTo(_ *Stream)    {}

func (a *invalidAny) Get(path ...interface{}) Any {
	if a.err == nil {
		return &invalidAny{baseAny{}, fmt.Errorf("get %v from invalid", path)}
	}
	return &invalidAny{baseAny{}, fmt.Errorf("%v, get %v from invalid", a.err, path)}
}

func (a *invalidAny) Parse() *Iterator          { return nil }
func (a *invalidAny) GetInterface() interface{} { return nil }
