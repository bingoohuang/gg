package misc_tests

import (
	"context"
	"encoding/json"
	"math"
	"testing"

	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/stretchr/testify/require"
)

func Test_read_big_float(t *testing.T) {
	should := require.New(t)
	iter := jsoni.ParseString(jsoni.ConfigDefault, `12.3`)
	val := iter.ReadBigFloat()
	val64, _ := val.Float64()
	should.Equal(12.3, val64)
}

func Test_read_big_int(t *testing.T) {
	should := require.New(t)
	iter := jsoni.ParseString(jsoni.ConfigDefault, `92233720368547758079223372036854775807`)
	val := iter.ReadBigInt()
	should.NotNil(val)
	should.Equal(`92233720368547758079223372036854775807`, val.String())
}

func Test_read_float_as_interface(t *testing.T) {
	should := require.New(t)
	iter := jsoni.ParseString(jsoni.ConfigDefault, `12.3`)

	should.Equal(12.3, iter.Read(context.Background()))
}

func Test_wrap_float(t *testing.T) {
	should := require.New(t)
	str, err := jsoni.MarshalToString(jsoni.WrapFloat64(12.3))
	should.Nil(err)
	should.Equal("12.3", str)
}

func Test_read_float64_cursor(t *testing.T) {
	should := require.New(t)
	ctx := context.Background()
	iter := jsoni.ParseString(jsoni.ConfigDefault, "[1.23456789\n,2,3]")
	should.True(iter.ReadArray())
	should.Equal(1.23456789, iter.Read(ctx))
	should.True(iter.ReadArray())
	should.Equal(float64(2), iter.Read(ctx))
}

func Test_read_float_scientific(t *testing.T) {
	should := require.New(t)
	var obj interface{}
	should.NoError(jsoni.UnmarshalFromString(`1e1`, &obj))
	should.Equal(float64(10), obj)
	should.NoError(json.Unmarshal([]byte(`1e1`), &obj))
	should.Equal(float64(10), obj)
	should.NoError(jsoni.UnmarshalFromString(`1.0e1`, &obj))
	should.Equal(float64(10), obj)
	should.NoError(json.Unmarshal([]byte(`1.0e1`), &obj))
	should.Equal(float64(10), obj)
}

func Test_lossy_float_marshal(t *testing.T) {
	should := require.New(t)
	ctx := context.Background()
	api := jsoni.Config{MarshalFloatWith6Digits: true}.Froze()
	output, err := api.MarshalToString(ctx, float64(0.1234567))
	should.Nil(err)
	should.Equal("0.123457", output)
	output, err = api.MarshalToString(ctx, float32(0.1234567))
	should.Nil(err)
	should.Equal("0.123457", output)
}

func Test_read_number(t *testing.T) {
	should := require.New(t)
	iter := jsoni.ParseString(jsoni.ConfigDefault, `92233720368547758079223372036854775807`)
	val := iter.ReadNumber()
	should.Equal(`92233720368547758079223372036854775807`, string(val))
}

func Test_encode_inf(t *testing.T) {
	should := require.New(t)
	_, err := json.Marshal(math.Inf(1))
	should.Error(err)
	_, err = jsoni.Marshal(float32(math.Inf(1)))
	should.Error(err)
	_, err = jsoni.Marshal(math.Inf(-1))
	should.Error(err)
}

func Test_encode_nan(t *testing.T) {
	should := require.New(t)
	_, err := json.Marshal(math.NaN())
	should.Error(err)
	_, err = jsoni.Marshal(float32(math.NaN()))
	should.Error(err)
	_, err = jsoni.Marshal(math.NaN())
	should.Error(err)
}

func Benchmark_jsoniter_float(b *testing.B) {
	b.ReportAllocs()
	input := []byte(`1.1123,`)
	iter := jsoni.NewIterator(jsoni.ConfigDefault)
	for n := 0; n < b.N; n++ {
		iter.ResetBytes(input)
		iter.ReadFloat64()
	}
}

func Benchmark_json_float(b *testing.B) {
	for n := 0; n < b.N; n++ {
		result := float64(0)
		json.Unmarshal([]byte(`1.1`), &result)
	}
}
