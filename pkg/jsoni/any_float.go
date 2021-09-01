package jsoni

import (
	"strconv"
)

type floatAny struct {
	baseAny
	val float64
}

func (a *floatAny) Parse() *Iterator     { return nil }
func (a *floatAny) ValueType() ValueType { return NumberValue }
func (a *floatAny) MustBeValid() Any     { return a }
func (a *floatAny) LastError() error     { return nil }
func (a *floatAny) ToBool() bool         { return a.ToFloat64() != 0 }
func (a *floatAny) ToInt() int           { return int(a.val) }
func (a *floatAny) ToInt32() int32       { return int32(a.val) }
func (a *floatAny) ToInt64() int64       { return int64(a.val) }
func (a *floatAny) ToUint() uint {
	if a.val > 0 {
		return uint(a.val)
	}
	return 0
}

func (a *floatAny) ToUint32() uint32 {
	if a.val > 0 {
		return uint32(a.val)
	}
	return 0
}

func (a *floatAny) ToUint64() uint64 {
	if a.val > 0 {
		return uint64(a.val)
	}
	return 0
}

func (a *floatAny) ToFloat32() float32 { return float32(a.val) }
func (a *floatAny) ToFloat64() float64 { return a.val }

func (a *floatAny) ToString() string {
	return strconv.FormatFloat(a.val, 'E', -1, 64)
}

func (a *floatAny) WriteTo(stream *Stream)    { stream.WriteFloat64(a.val) }
func (a *floatAny) GetInterface() interface{} { return a.val }
