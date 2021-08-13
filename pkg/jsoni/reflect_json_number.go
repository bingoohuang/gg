package jsoni

import (
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
var jsoniterNumberType = PtrElem((*Number)(nil))

func createDecoderOfJsonNumber(_ *ctx, typ reflect2.Type) ValDecoder {
	if typ.AssignableTo(jsonNumberType) {
		return &jsonNumberCodec{}
	}
	if typ.AssignableTo(jsoniterNumberType) {
		return &jsoniterNumberCodec{}
	}
	return nil
}

func createEncoderOfJsonNumber(_ *ctx, typ reflect2.Type) ValEncoder {
	if typ.AssignableTo(jsonNumberType) {
		return &jsonNumberCodec{}
	}
	if typ.AssignableTo(jsoniterNumberType) {
		return &jsoniterNumberCodec{}
	}
	return nil
}

type jsonNumberCodec struct{}

func (c *jsonNumberCodec) Decode(ptr unsafe.Pointer, iter *Iterator) {
	switch iter.WhatIsNext() {
	case StringValue:
		*((*json.Number)(ptr)) = json.Number(iter.ReadString())
	case NilValue:
		iter.skipFourBytes('n', 'u', 'l', 'l')
		*((*json.Number)(ptr)) = ""
	default:
		*((*json.Number)(ptr)) = json.Number([]byte(iter.readNumberAsString()))
	}
}

func (c *jsonNumberCodec) Encode(ptr unsafe.Pointer, stream *Stream) {
	number := *((*json.Number)(ptr))
	if len(number) == 0 {
		stream.writeByte('0')
	} else {
		stream.WriteRaw(string(number))
	}
}

func (c *jsonNumberCodec) IsEmpty(ptr unsafe.Pointer) bool {
	return len(*((*json.Number)(ptr))) == 0
}

type jsoniterNumberCodec struct{}

func (c *jsoniterNumberCodec) Decode(ptr unsafe.Pointer, iter *Iterator) {
	switch iter.WhatIsNext() {
	case StringValue:
		*((*Number)(ptr)) = Number(iter.ReadString())
	case NilValue:
		iter.skipFourBytes('n', 'u', 'l', 'l')
		*((*Number)(ptr)) = ""
	default:
		*((*Number)(ptr)) = Number([]byte(iter.readNumberAsString()))
	}
}

func (c *jsoniterNumberCodec) Encode(ptr unsafe.Pointer, stream *Stream) {
	number := *((*Number)(ptr))
	if len(number) == 0 {
		stream.writeByte('0')
	} else {
		stream.WriteRaw(string(number))
	}
}

func (c *jsoniterNumberCodec) IsEmpty(ptr unsafe.Pointer) bool {
	return len(*((*Number)(ptr))) == 0
}
