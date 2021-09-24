package jsoni

import (
	"context"
	"github.com/modern-go/reflect2"
	"unsafe"
)

func decoderOfOptional(ctx *ctx, typ reflect2.Type) ValDecoder {
	ptrType := typ.(*reflect2.UnsafePtrType)
	elemType := ptrType.Elem()
	decoder := decoderOfType(ctx, elemType)
	return &OptionalDecoder{elemType, decoder}
}

func encoderOfOptional(ctx *ctx, typ reflect2.Type) ValEncoder {
	ptrType := typ.(*reflect2.UnsafePtrType)
	elemType := ptrType.Elem()
	elemEncoder := encoderOfType(ctx, elemType)
	encoder := &OptionalEncoder{elemEncoder}
	return encoder
}

type OptionalDecoder struct {
	ValueType    reflect2.Type
	ValueDecoder ValDecoder
}

func (d *OptionalDecoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *Iterator) {
	if iter.ReadNil() {
		*((*unsafe.Pointer)(ptr)) = nil
	} else {
		if *((*unsafe.Pointer)(ptr)) == nil {
			//pointer to null, we have to allocate memory to hold the value
			newPtr := d.ValueType.UnsafeNew()
			d.ValueDecoder.Decode(ctx, newPtr, iter)
			*((*unsafe.Pointer)(ptr)) = newPtr
		} else {
			//reuse existing instance
			d.ValueDecoder.Decode(ctx, *((*unsafe.Pointer)(ptr)), iter)
		}
	}
}

type dereferenceDecoder struct {
	// only to deference a pointer
	valueType    reflect2.Type
	valueDecoder ValDecoder
}

func (d *dereferenceDecoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *Iterator) {
	if *((*unsafe.Pointer)(ptr)) == nil {
		//pointer to null, we have to allocate memory to hold the value
		newPtr := d.valueType.UnsafeNew()
		d.valueDecoder.Decode(ctx, newPtr, iter)
		*((*unsafe.Pointer)(ptr)) = newPtr
	} else {
		//reuse existing instance
		d.valueDecoder.Decode(ctx, *((*unsafe.Pointer)(ptr)), iter)
	}
}

type OptionalEncoder struct {
	ValueEncoder ValEncoder
}

func (e *OptionalEncoder) Encode(ctx context.Context, ptr unsafe.Pointer, stream *Stream) {
	if *((*unsafe.Pointer)(ptr)) == nil {
		stream.WriteNil()
	} else {
		e.ValueEncoder.Encode(ctx, *((*unsafe.Pointer)(ptr)), stream)
	}
}

func (e *OptionalEncoder) IsEmpty(_ context.Context, ptr unsafe.Pointer, _ bool) bool {
	return *((*unsafe.Pointer)(ptr)) == nil
}

type dereferenceEncoder struct {
	ValueEncoder ValEncoder
}

func (e *dereferenceEncoder) Encode(ctx context.Context, ptr unsafe.Pointer, stream *Stream) {
	if *((*unsafe.Pointer)(ptr)) == nil {
		stream.WriteNil()
	} else {
		e.ValueEncoder.Encode(ctx, *((*unsafe.Pointer)(ptr)), stream)
	}
}

func (e *dereferenceEncoder) IsEmpty(ctx context.Context, ptr unsafe.Pointer, checkZero bool) bool {
	if dePtr := *((*unsafe.Pointer)(ptr)); dePtr != nil {
		return e.ValueEncoder.IsEmpty(ctx, dePtr, checkZero)

	}
	return true
}

func (e *dereferenceEncoder) IsEmbeddedPtrNil(ptr unsafe.Pointer) bool {
	deReferenced := *((*unsafe.Pointer)(ptr))
	if deReferenced == nil {
		return true
	}
	isEmbeddedPtrNil, converted := e.ValueEncoder.(IsEmbeddedPtrNil)
	if !converted {
		return false
	}
	fieldPtr := deReferenced
	return isEmbeddedPtrNil.IsEmbeddedPtrNil(fieldPtr)
}

type referenceEncoder struct {
	encoder ValEncoder
}

func (e *referenceEncoder) Encode(ctx context.Context, p unsafe.Pointer, s *Stream) {
	e.encoder.Encode(ctx, unsafe.Pointer(&p), s)
}

func (e *referenceEncoder) IsEmpty(ctx context.Context, p unsafe.Pointer, checkZero bool) bool {
	return e.encoder.IsEmpty(ctx, unsafe.Pointer(&p), checkZero)
}

type referenceDecoder struct {
	decoder ValDecoder
}

func (d *referenceDecoder) Decode(ctx context.Context, p unsafe.Pointer, i *Iterator) {
	d.decoder.Decode(ctx, unsafe.Pointer(&p), i)
}
