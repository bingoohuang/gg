package jsoni

import (
	"context"
	"encoding/json"
	"github.com/modern-go/reflect2"
	"strconv"
	"unsafe"
)

type Number string

// String returns the literal text of the number.
func (n Number) String() string { return string(n) }

// Float64 returns the number as a float64.
func (n Number) Float64() (float64, error) {
	return strconv.ParseFloat(string(n), 64)
}

// Int64 returns the number as an int64.
func (n Number) Int64() (int64, error) {
	return strconv.ParseInt(string(n), 10, 64)
}

func CastJsonNumber(val interface{}) (string, bool) {
	switch typedVal := val.(type) {
	case json.Number:
		return string(typedVal), true
	case Number:
		return string(typedVal), true
	}
	return "", false
}

var jsonNumberType = PtrElem((*json.Number)(nil))
var jsoniNumberType = PtrElem((*Number)(nil))

func createDecoderOfJsonNumber(_ *ctx, typ reflect2.Type) ValDecoder {
	if typ.AssignableTo(jsonNumberType) {
		return &jsonNumberCodec{}
	}
	if typ.AssignableTo(jsoniNumberType) {
		return &jsoniNumberCodec{}
	}
	return nil
}

func createEncoderOfJsonNumber(_ *ctx, typ reflect2.Type) ValEncoder {
	if typ.AssignableTo(jsonNumberType) {
		return &jsonNumberCodec{}
	}
	if typ.AssignableTo(jsoniNumberType) {
		return &jsoniNumberCodec{}
	}
	return nil
}

type jsonNumberCodec struct{}

func (c *jsonNumberCodec) Decode(_ context.Context, ptr unsafe.Pointer, iter *Iterator) {
	switch iter.WhatIsNext() {
	case StringValue:
		*((*json.Number)(ptr)) = json.Number(iter.ReadString())
	case NilValue:
		iter.skip4Bytes('n', 'u', 'l', 'l')
		*((*json.Number)(ptr)) = ""
	default:
		*((*json.Number)(ptr)) = json.Number(iter.readNumberAsString())
	}
}

func (c *jsonNumberCodec) Encode(_ context.Context, ptr unsafe.Pointer, stream *Stream) {
	if n := *((*json.Number)(ptr)); len(n) == 0 {
		stream.writeByte('0')
	} else {
		stream.WriteRaw(string(n))
	}
}

func (c *jsonNumberCodec) IsEmpty(_ context.Context, p unsafe.Pointer) bool {
	return len(*((*json.Number)(p))) == 0
}

type jsoniNumberCodec struct{}

func (c *jsoniNumberCodec) Decode(_ context.Context, ptr unsafe.Pointer, iter *Iterator) {
	switch r := (*Number)(ptr); iter.WhatIsNext() {
	case StringValue:
		*r = Number(iter.ReadString())
	case NilValue:
		iter.skip4Bytes('n', 'u', 'l', 'l')
		*r = ""
	default:
		*r = Number(iter.readNumberAsString())
	}
}

func (c *jsoniNumberCodec) Encode(_ context.Context, ptr unsafe.Pointer, stream *Stream) {
	if n := *((*Number)(ptr)); len(n) == 0 {
		stream.writeByte('0')
	} else {
		stream.WriteRaw(string(n))
	}
}

func (c *jsoniNumberCodec) IsEmpty(_ context.Context, p unsafe.Pointer) bool {
	return len(*((*Number)(p))) == 0
}
