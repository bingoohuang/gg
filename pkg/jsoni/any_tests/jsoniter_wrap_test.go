package any_tests

import (
	"context"
	"testing"

	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/stretchr/testify/require"
)

func Test_wrap_and_valuetype_everything(t *testing.T) {
	should := require.New(t)
	var i interface{}
	any := jsoni.Get([]byte("123"))
	// default of number type is float64
	i = float64(123)
	should.Equal(i, any.GetInterface(nil))

	any = jsoni.Wrap(int8(10))
	should.Equal(any.ValueType(), jsoni.NumberValue)
	should.Equal(any.LastError(), nil)
	//  get interface is not int8 interface
	// i = int8(10)
	// should.Equal(i, any.GetInterface())

	any = jsoni.Wrap(int16(10))
	should.Equal(any.ValueType(), jsoni.NumberValue)
	should.Equal(any.LastError(), nil)
	//i = int16(10)
	//should.Equal(i, any.GetInterface())

	any = jsoni.Wrap(int32(10))
	should.Equal(any.ValueType(), jsoni.NumberValue)
	should.Equal(any.LastError(), nil)
	i = int32(10)
	should.Equal(i, any.GetInterface(nil))
	any = jsoni.Wrap(int64(10))
	should.Equal(any.ValueType(), jsoni.NumberValue)
	should.Equal(any.LastError(), nil)
	i = int64(10)
	should.Equal(i, any.GetInterface(nil))

	any = jsoni.Wrap(uint(10))
	should.Equal(any.ValueType(), jsoni.NumberValue)
	should.Equal(any.LastError(), nil)
	// not equal
	//i = uint(10)
	//should.Equal(i, any.GetInterface())
	any = jsoni.Wrap(uint8(10))
	should.Equal(any.ValueType(), jsoni.NumberValue)
	should.Equal(any.LastError(), nil)
	// not equal
	// i = uint8(10)
	// should.Equal(i, any.GetInterface())
	any = jsoni.Wrap(uint16(10))
	should.Equal(any.ValueType(), jsoni.NumberValue)
	should.Equal(any.LastError(), nil)
	any = jsoni.Wrap(uint32(10))
	should.Equal(any.ValueType(), jsoni.NumberValue)
	should.Equal(any.LastError(), nil)
	i = uint32(10)
	should.Equal(i, any.GetInterface(nil))
	any = jsoni.Wrap(uint64(10))
	should.Equal(any.ValueType(), jsoni.NumberValue)
	should.Equal(any.LastError(), nil)
	i = uint64(10)
	should.Equal(i, any.GetInterface(nil))

	any = jsoni.Wrap(float32(10))
	should.Equal(any.ValueType(), jsoni.NumberValue)
	should.Equal(any.LastError(), nil)
	// not equal
	//i = float32(10)
	//should.Equal(i, any.GetInterface())
	any = jsoni.Wrap(float64(10))
	should.Equal(any.ValueType(), jsoni.NumberValue)
	should.Equal(any.LastError(), nil)
	i = float64(10)
	should.Equal(i, any.GetInterface(nil))

	any = jsoni.Wrap(true)
	should.Equal(any.ValueType(), jsoni.BoolValue)
	should.Equal(any.LastError(), nil)
	i = true
	should.Equal(i, any.GetInterface(nil))
	any = jsoni.Wrap(false)
	should.Equal(any.ValueType(), jsoni.BoolValue)
	should.Equal(any.LastError(), nil)
	i = false
	should.Equal(i, any.GetInterface(nil))

	any = jsoni.Wrap(nil)
	should.Equal(any.ValueType(), jsoni.NilValue)
	should.Equal(any.LastError(), nil)
	i = nil
	should.Equal(i, any.GetInterface(nil))

	stream := jsoni.NewStream(jsoni.ConfigDefault, nil, 32)
	any.WriteTo(context.Background(), stream)
	should.Equal("null", string(stream.Buffer()))
	should.Equal(any.LastError(), nil)

	any = jsoni.Wrap(struct{ age int }{age: 1})
	should.Equal(any.ValueType(), jsoni.ObjectValue)
	should.Equal(any.LastError(), nil)
	i = struct{ age int }{age: 1}
	should.Equal(i, any.GetInterface(nil))

	any = jsoni.Wrap(map[string]interface{}{"abc": 1})
	should.Equal(any.ValueType(), jsoni.ObjectValue)
	should.Equal(any.LastError(), nil)
	i = map[string]interface{}{"abc": 1}
	should.Equal(i, any.GetInterface(nil))

	any = jsoni.Wrap("abc")
	i = "abc"
	should.Equal(i, any.GetInterface(nil))
	should.Equal(nil, any.LastError())

}
