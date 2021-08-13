package jsoni

import (
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

func (d *OptionalDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if iter.ReadNil() {
		*((*unsafe.Pointer)(ptr)) = nil
	} else {
		if *((*unsafe.Pointer)(ptr)) == nil {
			//pointer to null, we have to allocate memory to hold the value
			newPtr := d.ValueType.UnsafeNew()
			d.ValueDecoder.Decode(newPtr, iter)
			*((*unsafe.Pointer)(ptr)) = newPtr
		} else {
			//reuse existing instance
			d.ValueDecoder.Decode(*((*unsafe.Pointer)(ptr)), iter)
		}
	}
}

type dereferenceDecoder struct {
	// only to deference a pointer
	valueType    reflect2.Type
	valueDecoder ValDecoder
}

func (d *dereferenceDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if *((*unsafe.Pointer)(ptr)) == nil {
		//pointer to null, we have to allocate memory to hold the value
		newPtr := d.valueType.UnsafeNew()
		d.valueDecoder.Decode(newPtr, iter)
		*((*unsafe.Pointer)(ptr)) = newPtr
	} else {
		//reuse existing instance
		d.valueDecoder.Decode(*((*unsafe.Pointer)(ptr)), iter)
	}
}

type OptionalEncoder struct {
	ValueEncoder ValEncoder
}

func (e *OptionalEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	if *((*unsafe.Pointer)(ptr)) == nil {
		stream.WriteNil()
	} else {
		e.ValueEncoder.Encode(*((*unsafe.Pointer)(ptr)), stream)
	}
}

func (e *OptionalEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*unsafe.Pointer)(ptr)) == nil
}

type dereferenceEncoder struct {
	ValueEncoder ValEncoder
}

func (e *dereferenceEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	if *((*unsafe.Pointer)(ptr)) == nil {
		stream.WriteNil()
	} else {
		e.ValueEncoder.Encode(*((*unsafe.Pointer)(ptr)), stream)
	}
}

func (e *dereferenceEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	dePtr := *((*unsafe.Pointer)(ptr))
	if dePtr == nil {
		return true
	}
	return e.ValueEncoder.IsEmpty(dePtr)
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
	fieldPtr := unsafe.Pointer(deReferenced)
	return isEmbeddedPtrNil.IsEmbeddedPtrNil(fieldPtr)
}

type referenceEncoder struct {
	encoder ValEncoder
}

func (e *referenceEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	e.encoder.Encode(unsafe.Pointer(&ptr), stream)
}

func (e *referenceEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return e.encoder.IsEmpty(unsafe.Pointer(&ptr))
}

type referenceDecoder struct {
	decoder ValDecoder
}

func (d *referenceDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	d.decoder.Decode(unsafe.Pointer(&ptr), iter)
}
