package jsoni

import "context"

type trueAny struct{ baseAny }

func (a *trueAny) LastError() error                          { return nil }
func (a *trueAny) ToBool() bool                              { return true }
func (a *trueAny) ToInt() int                                { return 1 }
func (a *trueAny) ToInt32() int32                            { return 1 }
func (a *trueAny) ToInt64() int64                            { return 1 }
func (a *trueAny) ToUint() uint                              { return 1 }
func (a *trueAny) ToUint32() uint32                          { return 1 }
func (a *trueAny) ToUint64() uint64                          { return 1 }
func (a *trueAny) ToFloat32() float32                        { return 1 }
func (a *trueAny) ToFloat64() float64                        { return 1 }
func (a *trueAny) ToString() string                          { return "true" }
func (a *trueAny) WriteTo(_ context.Context, stream *Stream) { stream.WriteTrue() }
func (a *trueAny) Parse() *Iterator                          { return nil }
func (a *trueAny) GetInterface(context.Context) interface{}  { return true }
func (a *trueAny) ValueType() ValueType                      { return BoolValue }
func (a *trueAny) MustBeValid() Any                          { return a }

type falseAny struct{ baseAny }

func (a *falseAny) LastError() error                          { return nil }
func (a *falseAny) ToBool() bool                              { return false }
func (a *falseAny) ToInt() int                                { return 0 }
func (a *falseAny) ToInt32() int32                            { return 0 }
func (a *falseAny) ToInt64() int64                            { return 0 }
func (a *falseAny) ToUint() uint                              { return 0 }
func (a *falseAny) ToUint32() uint32                          { return 0 }
func (a *falseAny) ToUint64() uint64                          { return 0 }
func (a *falseAny) ToFloat32() float32                        { return 0 }
func (a *falseAny) ToFloat64() float64                        { return 0 }
func (a *falseAny) ToString() string                          { return "false" }
func (a *falseAny) WriteTo(_ context.Context, stream *Stream) { stream.WriteFalse() }
func (a *falseAny) Parse() *Iterator                          { return nil }
func (a *falseAny) GetInterface(context.Context) interface{}  { return false }
func (a *falseAny) ValueType() ValueType                      { return BoolValue }
func (a *falseAny) MustBeValid() Any                          { return a }
