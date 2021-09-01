package jsoni

type nilAny struct{ baseAny }

func (a *nilAny) LastError() error          { return nil }
func (a *nilAny) ValueType() ValueType      { return NilValue }
func (a *nilAny) MustBeValid() Any          { return a }
func (a *nilAny) ToBool() bool              { return false }
func (a *nilAny) ToInt() int                { return 0 }
func (a *nilAny) ToInt32() int32            { return 0 }
func (a *nilAny) ToInt64() int64            { return 0 }
func (a *nilAny) ToUint() uint              { return 0 }
func (a *nilAny) ToUint32() uint32          { return 0 }
func (a *nilAny) ToUint64() uint64          { return 0 }
func (a *nilAny) ToFloat32() float32        { return 0 }
func (a *nilAny) ToFloat64() float64        { return 0 }
func (a *nilAny) ToString() string          { return "" }
func (a *nilAny) WriteTo(stream *Stream)    { stream.WriteNil() }
func (a *nilAny) Parse() *Iterator          { return nil }
func (a *nilAny) GetInterface() interface{} { return nil }
