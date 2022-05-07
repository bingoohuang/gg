package jsoni

import (
	"context"
	"encoding/base64"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/modern-go/reflect2"
)

const ptrSize = 32 << (^uintptr(0) >> 63)

func createEncoderOfNative(ctx *ctx, typ reflect2.Type) ValEncoder {
	if typ.Kind() == reflect.Slice && typ.(reflect2.SliceType).Elem().Kind() == reflect.Uint8 {
		return &base64Codec{sliceDecoder: decoderOfSlice(ctx, typ)}
	}
	typeName := typ.String()
	kind := typ.Kind()
	switch kind {
	case reflect.String:
		if typeName != "string" {
			return encoderOfType(ctx, PtrElem((*string)(nil)))
		}
		return &stringCodec{}
	case reflect.Int:
		if typeName != "int" {
			return encoderOfType(ctx, PtrElem((*int)(nil)))
		}
		if strconv.IntSize == 32 {
			return &int32Codec{}
		}
		return &int64Codec{}
	case reflect.Int8:
		if typeName != "int8" {
			return encoderOfType(ctx, PtrElem((*int8)(nil)))
		}
		return &int8Codec{}
	case reflect.Int16:
		if typeName != "int16" {
			return encoderOfType(ctx, PtrElem((*int16)(nil)))
		}
		return &int16Codec{}
	case reflect.Int32:
		if typeName != "int32" {
			return encoderOfType(ctx, PtrElem((*int32)(nil)))
		}
		return &int32Codec{}
	case reflect.Int64:
		if typeName != "int64" {
			return encoderOfType(ctx, PtrElem((*int64)(nil)))
		}
		if ctx.int64AsString {
			return &stringModeNumberEncoder{&int64Codec{}}
		}
		return &int64Codec{}
	case reflect.Uint:
		if typeName != "uint" {
			return encoderOfType(ctx, PtrElem((*uint)(nil)))
		}
		if strconv.IntSize == 32 {
			return &uint32Codec{}
		}
		return &uint64Codec{}
	case reflect.Uint8:
		if typeName != "uint8" {
			return encoderOfType(ctx, PtrElem((*uint8)(nil)))
		}
		return &uint8Codec{}
	case reflect.Uint16:
		if typeName != "uint16" {
			return encoderOfType(ctx, PtrElem((*uint16)(nil)))
		}
		return &uint16Codec{}
	case reflect.Uint32:
		if typeName != "uint32" {
			return encoderOfType(ctx, PtrElem((*uint32)(nil)))
		}
		return &uint32Codec{}
	case reflect.Uintptr:
		if typeName != "uintptr" {
			return encoderOfType(ctx, PtrElem((*uintptr)(nil)))
		}
		if ptrSize == 32 {
			return &uint32Codec{}
		}
		return &uint64Codec{}
	case reflect.Uint64:
		if typeName != "uint64" {
			return encoderOfType(ctx, PtrElem((*uint64)(nil)))
		}
		if ctx.int64AsString {
			return &stringModeNumberEncoder{&uint64Codec{}}
		}
		return &uint64Codec{}
	case reflect.Float32:
		if typeName != "float32" {
			return encoderOfType(ctx, PtrElem((*float32)(nil)))
		}
		return &float32Codec{}
	case reflect.Float64:
		if typeName != "float64" {
			return encoderOfType(ctx, PtrElem((*float64)(nil)))
		}
		return &float64Codec{}
	case reflect.Bool:
		if typeName != "bool" {
			return encoderOfType(ctx, PtrElem((*bool)(nil)))
		}
		return &boolCodec{}
	}
	return nil
}

func createDecoderOfNative(ctx *ctx, typ reflect2.Type) ValDecoder {
	if typ.Kind() == reflect.Slice && typ.(reflect2.SliceType).Elem().Kind() == reflect.Uint8 {
		return &base64Codec{sliceDecoder: decoderOfSlice(ctx, typ)}
	}
	typeName := typ.String()
	switch typ.Kind() {
	case reflect.String:
		if typeName != "string" {
			return decoderOfType(ctx, PtrElem((*string)(nil)))
		}
		return &stringCodec{}
	case reflect.Int:
		if typeName != "int" {
			return decoderOfType(ctx, PtrElem((*int)(nil)))
		}
		if strconv.IntSize == 32 {
			return &int32Codec{}
		}
		return &int64Codec{}
	case reflect.Int8:
		if typeName != "int8" {
			return decoderOfType(ctx, PtrElem((*int8)(nil)))
		}
		return &int8Codec{}
	case reflect.Int16:
		if typeName != "int16" {
			return decoderOfType(ctx, PtrElem((*int16)(nil)))
		}
		return &int16Codec{}
	case reflect.Int32:
		if typeName != "int32" {
			return decoderOfType(ctx, PtrElem((*int32)(nil)))
		}
		return &int32Codec{}
	case reflect.Int64:
		if typeName != "int64" {
			return decoderOfType(ctx, PtrElem((*int64)(nil)))
		}
		if ctx.int64AsString {
			return &stringModeNumberCompatibleDecoder{&int64Codec{}}
		}
		return &int64Codec{}
	case reflect.Uint:
		if typeName != "uint" {
			return decoderOfType(ctx, PtrElem((*uint)(nil)))
		}
		if strconv.IntSize == 32 {
			return &uint32Codec{}
		}
		return &uint64Codec{}
	case reflect.Uint8:
		if typeName != "uint8" {
			return decoderOfType(ctx, PtrElem((*uint8)(nil)))
		}
		return &uint8Codec{}
	case reflect.Uint16:
		if typeName != "uint16" {
			return decoderOfType(ctx, PtrElem((*uint16)(nil)))
		}
		return &uint16Codec{}
	case reflect.Uint32:
		if typeName != "uint32" {
			return decoderOfType(ctx, PtrElem((*uint32)(nil)))
		}
		return &uint32Codec{}
	case reflect.Uintptr:
		if typeName != "uintptr" {
			return decoderOfType(ctx, PtrElem((*uintptr)(nil)))
		}
		if ptrSize == 32 {
			return &uint32Codec{}
		}
		return &uint64Codec{}
	case reflect.Uint64:
		if typeName != "uint64" {
			return decoderOfType(ctx, PtrElem((*uint64)(nil)))
		}
		if ctx.int64AsString {
			return &stringModeNumberCompatibleDecoder{&int64Codec{}}
		}
		return &uint64Codec{}
	case reflect.Float32:
		if typeName != "float32" {
			return decoderOfType(ctx, PtrElem((*float32)(nil)))
		}
		return &float32Codec{}
	case reflect.Float64:
		if typeName != "float64" {
			return decoderOfType(ctx, PtrElem((*float64)(nil)))
		}
		return &float64Codec{}
	case reflect.Bool:
		if typeName != "bool" {
			return decoderOfType(ctx, PtrElem((*bool)(nil)))
		}
		return &boolCodec{}
	}
	return nil
}

type stringCodec struct{}

func (c *stringCodec) Decode(_ context.Context, ptr unsafe.Pointer, iter *Iterator) {
	*((*string)(ptr)) = iter.ReadString()
}

func (c *stringCodec) Encode(ctx context.Context, ptr unsafe.Pointer, stream *Stream) {
	s := *((*string)(ptr))

	if writeRawBytesIfClearQuotes(ctx, s, stream) {
		return
	}

	stream.WriteString(s)
}

func (c *stringCodec) IsEmpty(_ context.Context, p unsafe.Pointer, _ bool) bool {
	return *((*string)(p)) == ""
}

type int8Codec struct{}

func (c *int8Codec) Decode(_ context.Context, ptr unsafe.Pointer, iter *Iterator) {
	if !iter.ReadNil() {
		*((*int8)(ptr)) = iter.ReadInt8()
	}
}

func (c *int8Codec) Encode(_ context.Context, ptr unsafe.Pointer, stream *Stream) {
	stream.WriteInt8(*((*int8)(ptr)))
}

func (c *int8Codec) IsEmpty(_ context.Context, p unsafe.Pointer, _ bool) bool {
	return *((*int8)(p)) == 0
}

type int16Codec struct{}

func (c *int16Codec) Decode(_ context.Context, ptr unsafe.Pointer, iter *Iterator) {
	if !iter.ReadNil() {
		*((*int16)(ptr)) = iter.ReadInt16()
	}
}

func (c *int16Codec) Encode(_ context.Context, ptr unsafe.Pointer, stream *Stream) {
	stream.WriteInt16(*((*int16)(ptr)))
}

func (c *int16Codec) IsEmpty(_ context.Context, p unsafe.Pointer, _ bool) bool {
	return *((*int16)(p)) == 0
}

type int32Codec struct{}

func (c *int32Codec) Decode(_ context.Context, ptr unsafe.Pointer, iter *Iterator) {
	if !iter.ReadNil() {
		*((*int32)(ptr)) = iter.ReadInt32()
	}
}

func (c *int32Codec) Encode(_ context.Context, ptr unsafe.Pointer, stream *Stream) {
	stream.WriteInt32(*((*int32)(ptr)))
}

func (c *int32Codec) IsEmpty(_ context.Context, p unsafe.Pointer, _ bool) bool {
	return *((*int32)(p)) == 0
}

type int64Codec struct{}

func (c *int64Codec) Decode(_ context.Context, ptr unsafe.Pointer, iter *Iterator) {
	if !iter.ReadNil() {
		*((*int64)(ptr)) = iter.ReadInt64()
	}
}

func (c *int64Codec) Encode(_ context.Context, ptr unsafe.Pointer, stream *Stream) {
	stream.WriteInt64(*((*int64)(ptr)))
}

func (c *int64Codec) IsEmpty(_ context.Context, p unsafe.Pointer, _ bool) bool {
	return *((*int64)(p)) == 0
}

type uint8Codec struct{}

func (c *uint8Codec) Decode(_ context.Context, ptr unsafe.Pointer, iter *Iterator) {
	if !iter.ReadNil() {
		*((*uint8)(ptr)) = iter.ReadUint8()
	}
}

func (c *uint8Codec) Encode(_ context.Context, ptr unsafe.Pointer, stream *Stream) {
	stream.WriteUint8(*((*uint8)(ptr)))
}

func (c *uint8Codec) IsEmpty(_ context.Context, ptr unsafe.Pointer, _ bool) bool {
	return *((*uint8)(ptr)) == 0
}

type uint16Codec struct{}

func (c *uint16Codec) Decode(_ context.Context, ptr unsafe.Pointer, iter *Iterator) {
	if !iter.ReadNil() {
		*((*uint16)(ptr)) = iter.ReadUint16()
	}
}

func (c *uint16Codec) Encode(_ context.Context, ptr unsafe.Pointer, stream *Stream) {
	stream.WriteUint16(*((*uint16)(ptr)))
}

func (c *uint16Codec) IsEmpty(_ context.Context, p unsafe.Pointer, _ bool) bool {
	return *((*uint16)(p)) == 0
}

type uint32Codec struct{}

func (c *uint32Codec) Decode(_ context.Context, ptr unsafe.Pointer, iter *Iterator) {
	if !iter.ReadNil() {
		*((*uint32)(ptr)) = iter.ReadUint32()
	}
}

func (c *uint32Codec) Encode(_ context.Context, ptr unsafe.Pointer, stream *Stream) {
	stream.WriteUint32(*((*uint32)(ptr)))
}

func (c *uint32Codec) IsEmpty(_ context.Context, p unsafe.Pointer, _ bool) bool {
	return *((*uint32)(p)) == 0
}

type uint64Codec struct{}

func (c *uint64Codec) Decode(_ context.Context, ptr unsafe.Pointer, iter *Iterator) {
	if !iter.ReadNil() {
		*((*uint64)(ptr)) = iter.ReadUint64()
	}
}

func (c *uint64Codec) Encode(_ context.Context, ptr unsafe.Pointer, stream *Stream) {
	stream.WriteUint64(*((*uint64)(ptr)))
}

func (c *uint64Codec) IsEmpty(_ context.Context, p unsafe.Pointer, _ bool) bool {
	return *((*uint64)(p)) == 0
}

type float32Codec struct{}

func (c *float32Codec) Decode(_ context.Context, ptr unsafe.Pointer, iter *Iterator) {
	if !iter.ReadNil() {
		*((*float32)(ptr)) = iter.ReadFloat32()
	}
}

func (c *float32Codec) Encode(_ context.Context, ptr unsafe.Pointer, stream *Stream) {
	stream.WriteFloat32(*((*float32)(ptr)))
}

func (c *float32Codec) IsEmpty(_ context.Context, p unsafe.Pointer, _ bool) bool {
	return *((*float32)(p)) == 0
}

type float64Codec struct{}

func (c *float64Codec) Decode(_ context.Context, ptr unsafe.Pointer, iter *Iterator) {
	if !iter.ReadNil() {
		*((*float64)(ptr)) = iter.ReadFloat64()
	}
}

func (c *float64Codec) Encode(_ context.Context, ptr unsafe.Pointer, stream *Stream) {
	stream.WriteFloat64(*((*float64)(ptr)))
}

func (c *float64Codec) IsEmpty(_ context.Context, p unsafe.Pointer, _ bool) bool {
	return *((*float64)(p)) == 0
}

type boolCodec struct{}

func (c *boolCodec) Decode(_ context.Context, ptr unsafe.Pointer, iter *Iterator) {
	if !iter.ReadNil() {
		*((*bool)(ptr)) = iter.ReadBool()
	}
}

func (c *boolCodec) Encode(_ context.Context, ptr unsafe.Pointer, stream *Stream) {
	stream.WriteBool(*((*bool)(ptr)))
}

func (c *boolCodec) IsEmpty(_ context.Context, p unsafe.Pointer, _ bool) bool {
	return !(*((*bool)(p)))
}

type base64Codec struct {
	sliceType    *reflect2.UnsafeSliceType
	sliceDecoder ValDecoder
}

func (c *base64Codec) Decode(ctx context.Context, ptr unsafe.Pointer, iter *Iterator) {
	if iter.ReadNil() {
		c.sliceType.UnsafeSetNil(ptr)
		return
	}
	switch iter.WhatIsNext() {
	case StringValue:
		src := iter.ReadString()
		dst, err := base64.StdEncoding.DecodeString(src)
		if err != nil {
			iter.ReportError("decode base64", err.Error())
		} else {
			c.sliceType.UnsafeSet(ptr, unsafe.Pointer(&dst))
		}
	case ArrayValue:
		c.sliceDecoder.Decode(ctx, ptr, iter)
	default:
		iter.ReportError("base64Codec", "invalid input")
	}
}

func (c *base64Codec) Encode(_ context.Context, ptr unsafe.Pointer, stream *Stream) {
	if c.sliceType.UnsafeIsNil(ptr) {
		stream.WriteNil()
		return
	}
	src := *((*[]byte)(ptr))
	encoding := base64.StdEncoding
	stream.writeByte('"')
	if len(src) > 0 {
		size := encoding.EncodedLen(len(src))
		buf := make([]byte, size)
		encoding.Encode(buf, src)
		stream.buf = append(stream.buf, buf...)
	}
	stream.writeByte('"')
}

func (c *base64Codec) IsEmpty(_ context.Context, p unsafe.Pointer, _ bool) bool {
	return len(*((*[]byte)(p))) == 0
}
