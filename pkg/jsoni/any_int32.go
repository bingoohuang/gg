package jsoni

import (
	"strconv"
)

type int32Any struct {
	baseAny
	val int32
}

func (a *int32Any) LastError() error          { return nil }
func (a *int32Any) ValueType() ValueType      { return NumberValue }
func (a *int32Any) MustBeValid() Any          { return a }
func (a *int32Any) ToBool() bool              { return a.val != 0 }
func (a *int32Any) ToInt() int                { return int(a.val) }
func (a *int32Any) ToInt32() int32            { return a.val }
func (a *int32Any) ToInt64() int64            { return int64(a.val) }
func (a *int32Any) ToUint() uint              { return uint(a.val) }
func (a *int32Any) ToUint32() uint32          { return uint32(a.val) }
func (a *int32Any) ToUint64() uint64          { return uint64(a.val) }
func (a *int32Any) ToFloat32() float32        { return float32(a.val) }
func (a *int32Any) ToFloat64() float64        { return float64(a.val) }
func (a *int32Any) ToString() string          { return strconv.FormatInt(int64(a.val), 10) }
func (a *int32Any) WriteTo(stream *Stream)    { stream.WriteInt32(a.val) }
func (a *int32Any) Parse() *Iterator          { return nil }
func (a *int32Any) GetInterface() interface{} { return a.val }
