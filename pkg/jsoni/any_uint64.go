package jsoni

import (
	"context"
	"strconv"
)

type uint64Any struct {
	baseAny
	val uint64
}

func (a *uint64Any) LastError() error                          { return nil }
func (a *uint64Any) ValueType() ValueType                      { return NumberValue }
func (a *uint64Any) MustBeValid() Any                          { return a }
func (a *uint64Any) ToBool() bool                              { return a.val != 0 }
func (a *uint64Any) ToInt() int                                { return int(a.val) }
func (a *uint64Any) ToInt32() int32                            { return int32(a.val) }
func (a *uint64Any) ToInt64() int64                            { return int64(a.val) }
func (a *uint64Any) ToUint() uint                              { return uint(a.val) }
func (a *uint64Any) ToUint32() uint32                          { return uint32(a.val) }
func (a *uint64Any) ToUint64() uint64                          { return a.val }
func (a *uint64Any) ToFloat32() float32                        { return float32(a.val) }
func (a *uint64Any) ToFloat64() float64                        { return float64(a.val) }
func (a *uint64Any) ToString() string                          { return strconv.FormatUint(a.val, 10) }
func (a *uint64Any) WriteTo(_ context.Context, stream *Stream) { stream.WriteUint64(a.val) }
func (a *uint64Any) Parse() *Iterator                          { return nil }
func (a *uint64Any) GetInterface(context.Context) interface{}  { return a.val }
