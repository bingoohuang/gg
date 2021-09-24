package jsoni

import (
	"context"
	"fmt"
	"github.com/modern-go/reflect2"
	"io"
	"reflect"
	"unsafe"
)

func encoderOfStruct(ctx *ctx, typ reflect2.Type) ValEncoder {
	type bindingTo struct {
		binding *Binding
		toName  string
		ignored bool
	}
	var orderedBindings []*bindingTo
	structDescriptor := describeStruct(ctx, typ)
	for _, binding := range structDescriptor.Fields {
		for _, toName := range binding.ToNames {
			b := &bindingTo{binding: binding, toName: toName}
			for _, old := range orderedBindings {
				if old.toName != toName {
					continue
				}
				old.ignored, b.ignored = resolveConflictBinding(ctx.frozenConfig, old.binding, b.binding)
			}
			orderedBindings = append(orderedBindings, b)
		}
	}
	if len(orderedBindings) == 0 {
		return &emptyStructEncoder{}
	}
	var finalOrderedFields []structFieldTo
	for _, b := range orderedBindings {
		if !b.ignored {
			finalOrderedFields = append(finalOrderedFields, structFieldTo{
				encoder: b.binding.Encoder.(*structFieldEncoder),
				toName:  b.toName,
			})
		}
	}
	return &structEncoder{typ: typ, fields: finalOrderedFields}
}

func createCheckIsEmpty(ctx *ctx, typ reflect2.Type) checkIsEmpty {
	if e := createEncoderOfNative(ctx, typ); e != nil {
		return e
	}

	switch kind := typ.Kind(); kind {
	case reflect.Interface:
		return &dynamicEncoder{typ}
	case reflect.Struct:
		return &structEncoder{typ: typ}
	case reflect.Array:
		return &arrayEncoder{}
	case reflect.Slice:
		return &sliceEncoder{}
	case reflect.Map:
		return encoderOfMap(ctx, typ)
	case reflect.Ptr:
		return &OptionalEncoder{}
	default:
		return &lazyErrorEncoder{err: fmt.Errorf("unsupported type: %v", typ)}
	}
}

func resolveConflictBinding(cfg *frozenConfig, old, new *Binding) (ignoreOld, ignoreNew bool) {
	newTagged := new.Field.Tag().Get(cfg.getTagKey()) != ""
	oldTagged := old.Field.Tag().Get(cfg.getTagKey()) != ""
	if newTagged {
		if oldTagged {
			if len(old.levels) > len(new.levels) {
				return true, false
			} else if len(new.levels) > len(old.levels) {
				return false, true
			} else {
				return true, true
			}
		} else {
			return true, false
		}
	} else {
		if oldTagged {
			return true, false
		}
		if len(old.levels) > len(new.levels) {
			return true, false
		} else if len(new.levels) > len(old.levels) {
			return false, true
		} else {
			return true, true
		}
	}
}

type structFieldEncoder struct {
	field        reflect2.StructField
	fieldEncoder ValEncoder
	omitempty    bool
}

func (e *structFieldEncoder) Encode(ctx context.Context, ptr unsafe.Pointer, stream *Stream) {
	fieldPtr := e.field.UnsafeGet(ptr)
	e.fieldEncoder.Encode(ctx, fieldPtr, stream)
	if stream.Error != nil && stream.Error != io.EOF {
		stream.Error = fmt.Errorf("%s: %s", e.field.Name(), stream.Error.Error())
	}
}

func (e *structFieldEncoder) IsEmpty(ctx context.Context, ptr unsafe.Pointer, checkZero bool) bool {
	fieldPtr := e.field.UnsafeGet(ptr)
	return e.fieldEncoder.IsEmpty(ctx, fieldPtr, checkZero)
}

func (e *structFieldEncoder) IsEmbeddedPtrNil(ptr unsafe.Pointer) bool {
	isEmbeddedPtrNil, converted := e.fieldEncoder.(IsEmbeddedPtrNil)
	if !converted {
		return false
	}
	fieldPtr := e.field.UnsafeGet(ptr)
	return isEmbeddedPtrNil.IsEmbeddedPtrNil(fieldPtr)
}

type IsEmbeddedPtrNil interface {
	IsEmbeddedPtrNil(ptr unsafe.Pointer) bool
}

type structEncoder struct {
	typ    reflect2.Type
	fields []structFieldTo
}

type structFieldTo struct {
	encoder *structFieldEncoder
	toName  string
}

func (e *structEncoder) Encode(ctx context.Context, ptr unsafe.Pointer, stream *Stream) {
	stream.WriteObjectStart()
	isNotFirst := false
	for _, field := range e.fields {
		if field.encoder.omitempty && field.encoder.IsEmpty(ctx, ptr, true) {
			continue
		}
		if field.encoder.IsEmbeddedPtrNil(ptr) {
			continue
		}
		if isNotFirst {
			stream.WriteMore()
		}
		stream.WriteObjectField(field.toName)
		field.encoder.Encode(ctx, ptr, stream)
		isNotFirst = true
	}
	stream.WriteObjectEnd()
	if stream.Error != nil && stream.Error != io.EOF {
		stream.Error = fmt.Errorf("%v.%s", e.typ, stream.Error.Error())
	}
}

func (e *structEncoder) IsEmpty(context.Context, unsafe.Pointer, bool) bool { return false }

type emptyStructEncoder struct{}

func (e *emptyStructEncoder) Encode(_ context.Context, _ unsafe.Pointer, stream *Stream) {
	stream.WriteEmptyObject()
}
func (e *emptyStructEncoder) IsEmpty(context.Context, unsafe.Pointer, bool) bool { return false }

type stringModeNumberEncoder struct {
	elemEncoder ValEncoder
}

func (e *stringModeNumberEncoder) Encode(ctx context.Context, ptr unsafe.Pointer, stream *Stream) {
	stream.writeByte('"')
	e.elemEncoder.Encode(ctx, ptr, stream)
	stream.writeByte('"')
}

func (e *stringModeNumberEncoder) IsEmpty(ctx context.Context, p unsafe.Pointer, checkZero bool) bool {
	return e.elemEncoder.IsEmpty(ctx, p, checkZero)
}

type stringModeStringEncoder struct {
	encoder ValEncoder
	cfg     *frozenConfig
}

func (e *stringModeStringEncoder) Encode(ctx context.Context, ptr unsafe.Pointer, stream *Stream) {
	temp := e.cfg.BorrowStream(nil)
	temp.Attachment = stream.Attachment
	defer e.cfg.ReturnStream(temp)
	e.encoder.Encode(ctx, ptr, temp)
	stream.WriteString(string(temp.Buffer()))
}

func (e *stringModeStringEncoder) IsEmpty(ctx context.Context, ptr unsafe.Pointer, checkZero bool) bool {
	return e.encoder.IsEmpty(ctx, ptr, checkZero)
}
