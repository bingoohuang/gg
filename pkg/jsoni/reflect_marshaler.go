package jsoni

import (
	"context"
	"encoding"
	"encoding/json"
	"unsafe"

	"github.com/modern-go/reflect2"
)

var (
	marshalerType       = PtrElem((*json.Marshaler)(nil))
	unmarshalerType     = PtrElem((*json.Unmarshaler)(nil))
	textMarshalerType   = PtrElem((*encoding.TextMarshaler)(nil))
	textUnmarshalerType = PtrElem((*encoding.TextUnmarshaler)(nil))

	marshalerContextType   = PtrElem((*MarshalerContext)(nil))
	unmarshalerContextType = PtrElem((*UnmarshalerContext)(nil))
)

func createDecoderOfMarshaler(_ *ctx, typ reflect2.Type) ValDecoder {
	ptrType := reflect2.PtrTo(typ)
	if ptrType.Implements(unmarshalerType) {
		return &referenceDecoder{decoder: &unmarshalerDecoder{valType: ptrType}}
	}
	if ptrType.Implements(unmarshalerContextType) {
		return &referenceDecoder{decoder: &unmarshalerContextDecoder{valType: ptrType}}
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
		return &marshalerEncoder{valueType: typ, checkIsEmpty: createCheckIsEmpty(ctx, typ)}
	}
	if typ == marshalerContextType {
		return &directMarshalerContextEncoder{checkIsEmpty: createCheckIsEmpty(ctx, typ)}
	}
	if typ.Implements(marshalerContextType) {
		return &marshalerContextEncoder{valueType: typ, checkIsEmpty: createCheckIsEmpty(ctx, typ)}
	}

	ptrType := reflect2.PtrTo(typ)
	if ptrType.Implements(marshalerType) {
		encoder := &marshalerEncoder{valueType: ptrType, checkIsEmpty: createCheckIsEmpty(ctx, ptrType)}
		return &referenceEncoder{encoder: encoder}
	}
	if ptrType.Implements(marshalerContextType) {
		encoder := &marshalerContextEncoder{valueType: ptrType, checkIsEmpty: createCheckIsEmpty(ctx, ptrType)}
		return &referenceEncoder{encoder: encoder}
	}

	if typ == textMarshalerType {
		return &directTextMarshalerEncoder{checkIsEmpty: createCheckIsEmpty(ctx, typ), stringEncoder: ctx.EncoderOf(reflect2.TypeOf(""))}
	}
	if typ.Implements(textMarshalerType) {
		return &textMarshalerEncoder{valType: typ, stringEncoder: ctx.EncoderOf(reflect2.TypeOf("")), checkIsEmpty: createCheckIsEmpty(ctx, typ)}
	}
	// if prefix is empty, the type is the root type
	if ctx.prefix != "" && ptrType.Implements(textMarshalerType) {
		encoder := &textMarshalerEncoder{valType: ptrType, stringEncoder: ctx.EncoderOf(reflect2.TypeOf("")), checkIsEmpty: createCheckIsEmpty(ctx, ptrType)}
		return &referenceEncoder{encoder: encoder}
	}
	return nil
}

type marshalerEncoder struct {
	checkIsEmpty checkIsEmpty
	valueType    reflect2.Type
}

func (e *marshalerEncoder) Encode(_ context.Context, ptr unsafe.Pointer, stream *Stream) {
	obj := e.valueType.UnsafeIndirect(ptr)
	if e.valueType.IsNullable() && reflect2.IsNil(obj) {
		stream.WriteNil()
		return
	}

	bytes, err := obj.(json.Marshaler).MarshalJSON()
	if err != nil {
		stream.Error = err
		return
	}
	// html escape was already done by jsoniter but the extra '\n' should be trimmed
	if l := len(bytes); l > 0 && bytes[l-1] == '\n' {
		bytes = bytes[:l-1]
	}
	stream.Write(bytes)
}

func (e *marshalerEncoder) IsEmpty(ctx context.Context, ptr unsafe.Pointer) bool {
	return e.checkIsEmpty.IsEmpty(ctx, ptr)
}

type marshalerContextEncoder struct {
	checkIsEmpty checkIsEmpty
	valueType    reflect2.Type
}

func (e *marshalerContextEncoder) Encode(ctx context.Context, ptr unsafe.Pointer, stream *Stream) {
	obj := e.valueType.UnsafeIndirect(ptr)
	if e.valueType.IsNullable() && reflect2.IsNil(obj) {
		stream.WriteNil()
		return
	}

	bytes, err := obj.(MarshalerContext).MarshalJSONContext(ctx)
	if err != nil {
		stream.Error = err
		return
	}
	// html escape was already done by jsoniter but the extra '\n' should be trimed
	if l := len(bytes); l > 0 && bytes[l-1] == '\n' {
		bytes = bytes[:l-1]
	}
	stream.Write(bytes)
}

func (e *marshalerContextEncoder) IsEmpty(ctx context.Context, ptr unsafe.Pointer) bool {
	return e.checkIsEmpty.IsEmpty(ctx, ptr)
}

type directMarshalerEncoder struct {
	checkIsEmpty checkIsEmpty
}

func (e *directMarshalerEncoder) Encode(_ context.Context, ptr unsafe.Pointer, stream *Stream) {
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

func (e *directMarshalerEncoder) IsEmpty(ctx context.Context, ptr unsafe.Pointer) bool {
	return e.checkIsEmpty.IsEmpty(ctx, ptr)
}

type directMarshalerContextEncoder struct {
	checkIsEmpty checkIsEmpty
}

func (e *directMarshalerContextEncoder) Encode(ctx context.Context, ptr unsafe.Pointer, stream *Stream) {
	marshaler := *(*MarshalerContext)(ptr)
	if marshaler == nil {
		stream.WriteNil()
		return
	}
	if bytes, err := marshaler.MarshalJSONContext(ctx); err != nil {
		stream.Error = err
	} else {
		stream.Write(bytes)
	}
}

func (e *directMarshalerContextEncoder) IsEmpty(ctx context.Context, ptr unsafe.Pointer) bool {
	return e.checkIsEmpty.IsEmpty(ctx, ptr)
}

type textMarshalerEncoder struct {
	valType       reflect2.Type
	stringEncoder ValEncoder
	checkIsEmpty  checkIsEmpty
}

func (e *textMarshalerEncoder) Encode(ctx context.Context, ptr unsafe.Pointer, stream *Stream) {
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
		e.stringEncoder.Encode(ctx, unsafe.Pointer(&str), stream)
	}
}

func (e *textMarshalerEncoder) IsEmpty(ctx context.Context, ptr unsafe.Pointer) bool {
	return e.checkIsEmpty.IsEmpty(ctx, ptr)
}

type directTextMarshalerEncoder struct {
	stringEncoder ValEncoder
	checkIsEmpty  checkIsEmpty
}

func (e *directTextMarshalerEncoder) Encode(ctx context.Context, ptr unsafe.Pointer, stream *Stream) {
	marshaler := *(*encoding.TextMarshaler)(ptr)
	if marshaler == nil {
		stream.WriteNil()
		return
	}
	if bytes, err := marshaler.MarshalText(); err != nil {
		stream.Error = err
	} else {
		str := string(bytes)
		e.stringEncoder.Encode(ctx, unsafe.Pointer(&str), stream)
	}
}

func (e *directTextMarshalerEncoder) IsEmpty(ctx context.Context, p unsafe.Pointer) bool {
	return e.checkIsEmpty.IsEmpty(ctx, p)
}

type unmarshalerDecoder struct{ valType reflect2.Type }

func (d *unmarshalerDecoder) Decode(_ context.Context, ptr unsafe.Pointer, iter *Iterator) {
	obj := d.valType.UnsafeIndirect(ptr)
	iter.nextToken()
	iter.unreadByte() // skip spaces
	bytes := iter.SkipAndReturnBytes()
	if err := obj.(json.Unmarshaler).UnmarshalJSON(bytes); err != nil {
		iter.ReportError("unmarshalerDecoder", err.Error())
	}
}

type unmarshalerContextDecoder struct{ valType reflect2.Type }

func (d *unmarshalerContextDecoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *Iterator) {
	obj := d.valType.UnsafeIndirect(ptr)
	iter.nextToken()
	iter.unreadByte() // skip spaces
	bytes := iter.SkipAndReturnBytes()
	if err := obj.(UnmarshalerContext).UnmarshalJSONContext(ctx, bytes); err != nil {
		iter.ReportError("unmarshalerDecoder", err.Error())
	}
}

type textUnmarshalerDecoder struct {
	valType reflect2.Type
}

func (d *textUnmarshalerDecoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *Iterator) {
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
