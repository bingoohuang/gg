package jsoni

import (
	"context"
	"strconv"
)

type int64Any struct {
	baseAny
	val int64
}

func (a *int64Any) LastError() error                          { return nil }
func (a *int64Any) ValueType() ValueType                      { return NumberValue }
func (a *int64Any) MustBeValid() Any                          { return a }
func (a *int64Any) ToBool() bool                              { return a.val != 0 }
func (a *int64Any) ToInt() int                                { return int(a.val) }
func (a *int64Any) ToInt32() int32                            { return int32(a.val) }
func (a *int64Any) ToInt64() int64                            { return a.val }
func (a *int64Any) ToUint() uint                              { return uint(a.val) }
func (a *int64Any) ToUint32() uint32                          { return uint32(a.val) }
func (a *int64Any) ToUint64() uint64                          { return uint64(a.val) }
func (a *int64Any) ToFloat32() float32                        { return float32(a.val) }
func (a *int64Any) ToFloat64() float64                        { return float64(a.val) }
func (a *int64Any) ToString() string                          { return strconv.FormatInt(a.val, 10) }
func (a *int64Any) WriteTo(_ context.Context, stream *Stream) { stream.WriteInt64(a.val) }
func (a *int64Any) Parse() *Iterator                          { return nil }
func (a *int64Any) GetInterface(context.Context) interface{}  { return a.val }
