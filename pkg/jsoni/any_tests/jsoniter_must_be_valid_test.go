package any_tests

import (
	"testing"

	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/stretchr/testify/require"
)

// if must be valid is useless, just drop this test
func Test_must_be_valid(t *testing.T) {
	should := require.New(t)
	any := jsoni.Get([]byte("123"))
	should.Equal(any.MustBeValid().ToInt(), 123)

	any = jsoni.Wrap(int8(10))
	should.Equal(any.MustBeValid().ToInt(), 10)

	any = jsoni.Wrap(int16(10))
	should.Equal(any.MustBeValid().ToInt(), 10)

	any = jsoni.Wrap(int32(10))
	should.Equal(any.MustBeValid().ToInt(), 10)

	any = jsoni.Wrap(int64(10))
	should.Equal(any.MustBeValid().ToInt(), 10)

	any = jsoni.Wrap(uint(10))
	should.Equal(any.MustBeValid().ToInt(), 10)

	any = jsoni.Wrap(uint8(10))
	should.Equal(any.MustBeValid().ToInt(), 10)

	any = jsoni.Wrap(uint16(10))
	should.Equal(any.MustBeValid().ToInt(), 10)

	any = jsoni.Wrap(uint32(10))
	should.Equal(any.MustBeValid().ToInt(), 10)

	any = jsoni.Wrap(uint64(10))
	should.Equal(any.MustBeValid().ToInt(), 10)

	any = jsoni.Wrap(float32(10))
	should.Equal(any.MustBeValid().ToFloat64(), float64(10))

	any = jsoni.Wrap(float64(10))
	should.Equal(any.MustBeValid().ToFloat64(), float64(10))

	any = jsoni.Wrap(true)
	should.Equal(any.MustBeValid().ToFloat64(), float64(1))

	any = jsoni.Wrap(false)
	should.Equal(any.MustBeValid().ToFloat64(), float64(0))

	any = jsoni.Wrap(nil)
	should.Equal(any.MustBeValid().ToFloat64(), float64(0))

	any = jsoni.Wrap(struct{ age int }{age: 1})
	should.Equal(any.MustBeValid().ToFloat64(), float64(0))

	any = jsoni.Wrap(map[string]interface{}{"abc": 1})
	should.Equal(any.MustBeValid().ToFloat64(), float64(0))

	any = jsoni.Wrap("abc")
	should.Equal(any.MustBeValid().ToFloat64(), float64(0))

	any = jsoni.Wrap([]int{})
	should.Equal(any.MustBeValid().ToFloat64(), float64(0))

	any = jsoni.Wrap([]int{1, 2})
	should.Equal(any.MustBeValid().ToFloat64(), float64(1))
}
