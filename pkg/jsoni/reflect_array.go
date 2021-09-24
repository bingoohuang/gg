package jsoni

import (
	"context"
	"fmt"
	"github.com/modern-go/reflect2"
	"io"
	"unsafe"
)

func decoderOfArray(ctx *ctx, typ reflect2.Type) ValDecoder {
	arrayType := typ.(*reflect2.UnsafeArrayType)
	decoder := decoderOfType(ctx.append("[arrayElem]"), arrayType.Elem())
	return &arrayDecoder{arrayType, decoder}
}

func encoderOfArray(ctx *ctx, typ reflect2.Type) ValEncoder {
	arrayType := typ.(*reflect2.UnsafeArrayType)
	if arrayType.Len() == 0 {
		return emptyArrayEncoder{}
	}
	encoder := encoderOfType(ctx.append("[arrayElem]"), arrayType.Elem())
	return &arrayEncoder{arrayType, encoder}
}

type emptyArrayEncoder struct{}

func (e emptyArrayEncoder) Encode(_ context.Context, _ unsafe.Pointer, stream *Stream) {
	stream.WriteEmptyArray()
}
func (e emptyArrayEncoder) IsEmpty(context.Context, unsafe.Pointer, bool) bool { return true }

type arrayEncoder struct {
	arrayType   *reflect2.UnsafeArrayType
	elemEncoder ValEncoder
}

func (e *arrayEncoder) Encode(ctx context.Context, ptr unsafe.Pointer, stream *Stream) {
	stream.WriteArrayStart()
	elemPtr := ptr
	e.elemEncoder.Encode(ctx, elemPtr, stream)
	for i := 1; i < e.arrayType.Len(); i++ {
		stream.WriteMore()
		elemPtr = e.arrayType.UnsafeGetIndex(ptr, i)
		e.elemEncoder.Encode(ctx, elemPtr, stream)
	}
	stream.WriteArrayEnd()
	if stream.Error != nil && stream.Error != io.EOF {
		stream.Error = fmt.Errorf("%v: %s", e.arrayType, stream.Error.Error())
	}
}

func (e *arrayEncoder) IsEmpty(context.Context, unsafe.Pointer, bool) bool { return false }

type arrayDecoder struct {
	arrayType   *reflect2.UnsafeArrayType
	elemDecoder ValDecoder
}

func (d *arrayDecoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *Iterator) {
	d.doDecode(ctx, ptr, iter)
	if iter.Error != nil && iter.Error != io.EOF {
		iter.Error = fmt.Errorf("%v: %s", d.arrayType, iter.Error.Error())
	}
}

func (d *arrayDecoder) doDecode(ctx context.Context, ptr unsafe.Pointer, iter *Iterator) {
	c := iter.nextToken()
	arrayType := d.arrayType
	if c == 'n' {
		iter.skip3Bytes('u', 'l', 'l')
		return
	}
	if c != '[' {
		iter.ReportError("decode array", "expect [ or n, but found "+string([]byte{c}))
		return
	}
	c = iter.nextToken()
	if c == ']' {
		return
	}
	iter.unreadByte()
	elemPtr := arrayType.UnsafeGetIndex(ptr, 0)
	d.elemDecoder.Decode(ctx, elemPtr, iter)
	length := 1
	for c = iter.nextToken(); c == ','; c = iter.nextToken() {
		if length >= arrayType.Len() {
			iter.Skip()
			continue
		}
		idx := length
		length += 1
		elemPtr = arrayType.UnsafeGetIndex(ptr, idx)
		d.elemDecoder.Decode(ctx, elemPtr, iter)
	}
	if c != ']' {
		iter.ReportError("decode array", "expect ], but found "+string([]byte{c}))
		return
	}
}
