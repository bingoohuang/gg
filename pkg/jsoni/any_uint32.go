package jsoni

import (
	"strconv"
)

type uint32Any struct {
	baseAny
	val uint32
}

func (a *uint32Any) LastError() error          { return nil }
func (a *uint32Any) ValueType() ValueType      { return NumberValue }
func (a *uint32Any) MustBeValid() Any          { return a }
func (a *uint32Any) ToBool() bool              { return a.val != 0 }
func (a *uint32Any) ToInt() int                { return int(a.val) }
func (a *uint32Any) ToInt32() int32            { return int32(a.val) }
func (a *uint32Any) ToInt64() int64            { return int64(a.val) }
func (a *uint32Any) ToUint() uint              { return uint(a.val) }
func (a *uint32Any) ToUint32() uint32          { return a.val }
func (a *uint32Any) ToUint64() uint64          { return uint64(a.val) }
func (a *uint32Any) ToFloat32() float32        { return float32(a.val) }
func (a *uint32Any) ToFloat64() float64        { return float64(a.val) }
func (a *uint32Any) ToString() string          { return strconv.FormatInt(int64(a.val), 10) }
func (a *uint32Any) WriteTo(stream *Stream)    { stream.WriteUint32(a.val) }
func (a *uint32Any) Parse() *Iterator          { return nil }
func (a *uint32Any) GetInterface() interface{} { return a.val }
