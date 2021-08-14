package jsoni

import (
	"encoding"
	"encoding/json"
	"unsafe"

	"github.com/modern-go/reflect2"
)

var marshalerType = PtrElem((*json.Marshaler)(nil))
var unmarshalerType = PtrElem((*json.Unmarshaler)(nil))
var textMarshalerType = PtrElem((*encoding.TextMarshaler)(nil))
var textUnmarshalerType = PtrElem((*encoding.TextUnmarshaler)(nil))

func createDecoderOfMarshaler(_ *ctx, typ reflect2.Type) ValDecoder {
	ptrType := reflect2.PtrTo(typ)
	if ptrType.Implements(unmarshalerType) {
		return &referenceDecoder{decoder: &unmarshalerDecoder{valType: ptrType}}
	}
	if ptrType.Implements(textUnmarshalerType) {
		return &referenceDecoder{decoder: &textUnmarshalerDecoder{valType: ptrType}}
	}
	return nil
}

func createEncoderOfMarshaler(ctx *ctx, typ reflect2.Type) ValEncoder {
	if typ == marshalerType {
		return &directMarshalerEncoder{checkIsEmpty: createCheckIsEmpty(ctx, typ)}
	}
	if typ.Implements(marshalerType) {
		return &marshalerEncoder{
			valType:      typ,
			checkIsEmpty: createCheckIsEmpty(ctx, typ),
		}
	}
	ptrType := reflect2.PtrTo(typ)
	if ctx.prefix != "" && ptrType.Implements(marshalerType) {
		encoder := &marshalerEncoder{
			valType:      ptrType,
			checkIsEmpty: createCheckIsEmpty(ctx, ptrType),
		}
		return &referenceEncoder{encoder}
	}
	if typ == textMarshalerType {
		return &directTextMarshalerEncoder{
			checkIsEmpty:  createCheckIsEmpty(ctx, typ),
			stringEncoder: ctx.EncoderOf(reflect2.TypeOf("")),
		}
	}
	if typ.Implements(textMarshalerType) {
		return &textMarshalerEncoder{
			valType:       typ,
			stringEncoder: ctx.EncoderOf(reflect2.TypeOf("")),
			checkIsEmpty:  createCheckIsEmpty(ctx, typ),
		}
	}
	// if prefix is empty, the type is the root type
	if ctx.prefix != "" && ptrType.Implements(textMarshalerType) {
		var encoder ValEncoder = &textMarshalerEncoder{
			valType:       ptrType,
			stringEncoder: ctx.EncoderOf(reflect2.TypeOf("")),
			checkIsEmpty:  createCheckIsEmpty(ctx, ptrType),
		}
		return &referenceEncoder{encoder}
	}
	return nil
}

type marshalerEncoder struct {
	checkIsEmpty checkIsEmpty
	valType      reflect2.Type
}

func (e *marshalerEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	obj := e.valType.UnsafeIndirect(ptr)
	if e.valType.IsNullable() && reflect2.IsNil(obj) {
		stream.WriteNil()
		return
	}
	marshaler := obj.(json.Marshaler)
	if bytes, err := marshaler.MarshalJSON(); err != nil {
		stream.Error = err
	} else {
		// html escape was already done by jsoniter
		// but the extra '\n' should be trimed
		l := len(bytes)
		if l > 0 && bytes[l-1] == '\n' {
			bytes = bytes[:l-1]
		}
		stream.Write(bytes)
	}
}

func (e *marshalerEncoder) IsEmpty(ptr unsafe.Pointer) bool { return e.checkIsEmpty.IsEmpty(ptr) }

type directMarshalerEncoder struct {
	checkIsEmpty checkIsEmpty
}

func (e *directMarshalerEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	marshaler := *(*json.Marshaler)(ptr)
	if marshaler == nil {
		stream.WriteNil()
		return
	}
	if bytes, err := marshaler.MarshalJSON(); err != nil {
		stream.Error = err
	} else {
		stream.Write(bytes)
	}
}

func (e *directMarshalerEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return e.checkIsEmpty.IsEmpty(ptr)
}

type textMarshalerEncoder struct {
	valType       reflect2.Type
	stringEncoder ValEncoder
	checkIsEmpty  checkIsEmpty
}

func (e *textMarshalerEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	obj := e.valType.UnsafeIndirect(ptr)
	if e.valType.IsNullable() && reflect2.IsNil(obj) {
		stream.WriteNil()
		return
	}
	marshaler := (obj).(encoding.TextMarshaler)
	if bytes, err := marshaler.MarshalText(); err != nil {
		stream.Error = err
	} else {
		str := string(bytes)
		e.stringEncoder.Encode(unsafe.Pointer(&str), stream)
	}
}

func (e *textMarshalerEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return e.checkIsEmpty.IsEmpty(ptr)
}

type directTextMarshalerEncoder struct {
	stringEncoder ValEncoder
	checkIsEmpty  checkIsEmpty
}

func (e *directTextMarshalerEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	marshaler := *(*encoding.TextMarshaler)(ptr)
	if marshaler == nil {
		stream.WriteNil()
		return
	}
	if bytes, err := marshaler.MarshalText(); err != nil {
		stream.Error = err
	} else {
		str := string(bytes)
		e.stringEncoder.Encode(unsafe.Pointer(&str), stream)
	}
}

func (e *directTextMarshalerEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return e.checkIsEmpty.IsEmpty(ptr)
}

type unmarshalerDecoder struct {
	valType reflect2.Type
}

func (d *unmarshalerDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	valType := d.valType
	obj := valType.UnsafeIndirect(ptr)
	unmarshaler := obj.(json.Unmarshaler)
	iter.nextToken()
	iter.unreadByte() // skip spaces
	bytes := iter.SkipAndReturnBytes()
	err := unmarshaler.UnmarshalJSON(bytes)
	if err != nil {
		iter.ReportError("unmarshalerDecoder", err.Error())
	}
}

type textUnmarshalerDecoder struct {
	valType reflect2.Type
}

func (d *textUnmarshalerDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	valType := d.valType
	obj := valType.UnsafeIndirect(ptr)
	if reflect2.IsNil(obj) {
		ptrType := valType.(*reflect2.UnsafePtrType)
		elemType := ptrType.Elem()
		elem := elemType.UnsafeNew()
		ptrType.UnsafeSet(ptr, unsafe.Pointer(&elem))
		obj = valType.UnsafeIndirect(ptr)
	}
	unmarshaler := (obj).(encoding.TextUnmarshaler)
	str := iter.ReadString()
	if err := unmarshaler.UnmarshalText([]byte(str)); err != nil {
		iter.ReportError("textUnmarshalerDecoder", err.Error())
	}
}
