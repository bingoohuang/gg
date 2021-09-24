package test

import (
	"context"
	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/modern-go/reflect2"
	"github.com/stretchr/testify/require"
	"reflect"
	"strconv"
	"testing"
	"unsafe"
)

type TestObject1 struct {
	Field1 string
}

type testExtension struct {
	jsoni.DummyExtension
}

func (extension *testExtension) UpdateStructDescriptor(structDescriptor *jsoni.StructDescriptor) {
	if structDescriptor.Type.String() != "test.TestObject1" {
		return
	}
	binding := structDescriptor.GetField("Field1")
	binding.Encoder = &funcEncoder{fun: func(_ context.Context, ptr unsafe.Pointer, stream *jsoni.Stream) {
		str := *((*string)(ptr))
		val, _ := strconv.Atoi(str)
		stream.WriteInt(val)
	}}
	binding.Decoder = &funcDecoder{func(_ context.Context, ptr unsafe.Pointer, iter *jsoni.Iterator) {
		*((*string)(ptr)) = strconv.Itoa(iter.ReadInt())
	}}
	binding.ToNames = []string{"field-1"}
	binding.FromNames = []string{"field-1"}
}

func Test_customize_field_by_extension(t *testing.T) {
	should := require.New(t)
	cfg := jsoni.Config{}.Froze()
	cfg.RegisterExtension(&testExtension{})
	obj := TestObject1{}
	ctx := context.Background()
	err := cfg.UnmarshalFromString(ctx, `{"field-1": 100}`, &obj)
	should.Nil(err)
	should.Equal("100", obj.Field1)
	str, err := cfg.MarshalToString(ctx, obj)
	should.Nil(err)
	should.Equal(`{"field-1":100}`, str)
}

func Test_customize_map_key_encoder(t *testing.T) {
	should := require.New(t)
	cfg := jsoni.Config{}.Froze()
	cfg.RegisterExtension(&testMapKeyExtension{})
	m := map[int]int{1: 2}
	ctx := context.Background()
	output, err := cfg.MarshalToString(ctx, m)
	should.NoError(err)
	should.Equal(`{"2":2}`, output)
	m = map[int]int{}
	should.NoError(cfg.UnmarshalFromString(ctx, output, &m))
	should.Equal(map[int]int{1: 2}, m)
}

type testMapKeyExtension struct {
	jsoni.DummyExtension
}

func (extension *testMapKeyExtension) CreateMapKeyEncoder(typ reflect2.Type) jsoni.ValEncoder {
	if typ.Kind() == reflect.Int {
		return &funcEncoder{
			fun: func(_ context.Context, ptr unsafe.Pointer, stream *jsoni.Stream) {
				stream.WriteRaw(`"`)
				stream.WriteInt(*(*int)(ptr) + 1)
				stream.WriteRaw(`"`)
			},
		}
	}
	return nil
}

func (extension *testMapKeyExtension) CreateMapKeyDecoder(typ reflect2.Type) jsoni.ValDecoder {
	if typ.Kind() == reflect.Int {
		return &funcDecoder{
			fun: func(_ context.Context, ptr unsafe.Pointer, iter *jsoni.Iterator) {
				i, err := strconv.Atoi(iter.ReadString())
				if err != nil {
					iter.ReportError("read map key", err.Error())
					return
				}
				i--
				*(*int)(ptr) = i
			},
		}
	}
	return nil
}

type funcDecoder struct {
	fun jsoni.DecoderFunc
}

func (decoder *funcDecoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *jsoni.Iterator) {
	decoder.fun(ctx, ptr, iter)
}

type funcEncoder struct {
	fun         jsoni.EncoderFunc
	isEmptyFunc func(ptr unsafe.Pointer) bool
}

func (encoder *funcEncoder) Encode(ctx context.Context, ptr unsafe.Pointer, stream *jsoni.Stream) {
	encoder.fun(ctx, ptr, stream)
}

func (encoder *funcEncoder) IsEmpty(_ context.Context, ptr unsafe.Pointer, _ bool) bool {
	if encoder.isEmptyFunc == nil {
		return false
	}
	return encoder.isEmptyFunc(ptr)
}
