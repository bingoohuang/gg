package jsoni

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"sort"
	"unsafe"

	"github.com/modern-go/reflect2"
)

type MapEntryEncoder interface {
	EntryEncoder() (keyEncoder, elemEncoder ValEncoder)
}

type MapEntryDecoder interface {
	EntryDecoder() (keyDecoder, elemDecoder ValDecoder)
}

func (e *sortKeysMapEncoder) EntryEncoder() (k, v ValEncoder) { return e.keyEncoder, e.elemEncoder }
func (e *mapEncoder) EntryEncoder() (k, v ValEncoder)         { return e.keyEncoder, e.elemEncoder }
func (d *mapDecoder) EntryDecoder() (k, v ValDecoder)         { return d.keyDecoder, d.elemDecoder }

func decoderOfMap(ctx *ctx, typ reflect2.Type) ValDecoder {
	mapType := typ.(*reflect2.UnsafeMapType)
	return &mapDecoder{
		mapType:     mapType,
		keyType:     mapType.Key(),
		elemType:    mapType.Elem(),
		keyDecoder:  decoderOfMapKey(ctx.append("[mapKey]"), mapType.Key()),
		elemDecoder: decoderOfType(ctx.append("[mapElem]"), mapType.Elem()),
	}
}

func encoderOfMap(ctx *ctx, typ reflect2.Type) ValEncoder {
	mapType := typ.(*reflect2.UnsafeMapType)
	keyEncoder := encoderOfMapKey(ctx.append("[mapKey]"), mapType.Key())
	elemEncoder := encoderOfType(ctx.append("[mapElem]"), mapType.Elem())
	encoder := &mapEncoder{mapType: mapType, keyEncoder: keyEncoder, elemEncoder: elemEncoder, omitempty: ctx.omitEmptyMapKeys}
	if ctx.sortMapKeys {
		return &sortKeysMapEncoder{mapEncoder: encoder}
	}
	return encoder
}

func decoderOfMapKey(ctx *ctx, typ reflect2.Type) ValDecoder {
	if decoder := ctx.decoderExtension.CreateMapKeyDecoder(typ); decoder != nil {
		return decoder
	}
	if decoder := ctx.extensions.CreateMapKeyDecoder(typ); decoder != nil {
		return decoder
	}

	ptrType := reflect2.PtrTo(typ)
	if ptrType.Implements(unmarshalerType) {
		return &referenceDecoder{decoder: &unmarshalerDecoder{valType: ptrType}}
	}
	if ptrType.Implements(unmarshalerContextType) {
		return &referenceDecoder{decoder: &unmarshalerContextDecoder{valType: ptrType}}
	}
	if typ.Implements(unmarshalerType) {
		return &unmarshalerDecoder{valType: typ}
	}
	if typ.Implements(unmarshalerContextType) {
		return &unmarshalerContextDecoder{valType: typ}
	}
	if ptrType.Implements(textUnmarshalerType) {
		return &referenceDecoder{decoder: &textUnmarshalerDecoder{valType: ptrType}}
	}
	if typ.Implements(textUnmarshalerType) {
		return &textUnmarshalerDecoder{valType: typ}
	}

	switch typ.Kind() {
	case reflect.String:
		return decoderOfType(ctx, reflect2.DefaultTypeOfKind(reflect.String))
	case reflect.Bool,
		reflect.Uint8, reflect.Int8, reflect.Uint16, reflect.Int16, reflect.Uint32, reflect.Int32,
		reflect.Uint64, reflect.Int64, reflect.Uint, reflect.Int,
		reflect.Float32, reflect.Float64, reflect.Uintptr:
		typ = reflect2.DefaultTypeOfKind(typ.Kind())
		return &numericMapKeyDecoder{decoder: decoderOfType(ctx, typ)}
	default:
		return &lazyErrorDecoder{err: fmt.Errorf("unsupported map key type: %v", typ)}
	}
}

func encoderOfMapKey(ctx *ctx, typ reflect2.Type) ValEncoder {
	if encoder := ctx.encoderExtension.CreateMapKeyEncoder(typ); encoder != nil {
		return encoder
	}
	if encoder := ctx.extensions.CreateMapKeyEncoder(typ); encoder != nil {
		return encoder
	}
	if typ == textMarshalerType {
		return &directTextMarshalerEncoder{stringEncoder: ctx.EncoderOf(reflect2.TypeOf(""))}
	}
	if typ.Implements(textMarshalerType) {
		return &textMarshalerEncoder{valType: typ, stringEncoder: ctx.EncoderOf(reflect2.TypeOf(""))}
	}

	switch typ.Kind() {
	case reflect.String:
		return encoderOfType(ctx, reflect2.DefaultTypeOfKind(reflect.String))
	case reflect.Bool,
		reflect.Uint8, reflect.Int8, reflect.Uint16, reflect.Int16, reflect.Uint32, reflect.Int32,
		reflect.Uint64, reflect.Int64, reflect.Uint, reflect.Int,
		reflect.Float32, reflect.Float64, reflect.Uintptr:
		typ = reflect2.DefaultTypeOfKind(typ.Kind())
		return &numericMapKeyEncoder{encoder: encoderOfType(ctx, typ)}
	default:
		if typ.Kind() == reflect.Interface {
			return &dynamicMapKeyEncoder{ctx: ctx, valType: typ}
		}
		return &lazyErrorEncoder{err: fmt.Errorf("unsupported map key type: %v", typ)}
	}
}

type mapDecoder struct {
	mapType     *reflect2.UnsafeMapType
	keyType     reflect2.Type
	elemType    reflect2.Type
	keyDecoder  ValDecoder
	elemDecoder ValDecoder
}

func (d *mapDecoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *Iterator) {
	mapType := d.mapType
	c := iter.nextToken()
	if c == 'n' {
		iter.skip3Bytes('u', 'l', 'l')
		*(*unsafe.Pointer)(ptr) = nil
		mapType.UnsafeSet(ptr, mapType.UnsafeNew())
		return
	}
	if mapType.UnsafeIsNil(ptr) {
		mapType.UnsafeSet(ptr, mapType.UnsafeMakeMap(0))
	}
	if c != '{' {
		iter.ReportError("ReadMapCB", `expect { or n, but found `+string([]byte{c}))
		return
	}
	c = iter.nextToken()
	if c == '}' {
		return
	}
	iter.unreadByte()
	key := d.keyType.UnsafeNew()
	d.keyDecoder.Decode(ctx, key, iter)
	c = iter.nextToken()
	if c != ':' {
		iter.ReportError("ReadMapCB", "expect : after object field, but found "+string([]byte{c}))
		return
	}
	elem := d.elemType.UnsafeNew()
	d.elemDecoder.Decode(ctx, elem, iter)
	d.mapType.UnsafeSetIndex(ptr, key, elem)
	for c = iter.nextToken(); c == ','; c = iter.nextToken() {
		key := d.keyType.UnsafeNew()
		d.keyDecoder.Decode(ctx, key, iter)
		c = iter.nextToken()
		if c != ':' {
			iter.ReportError("ReadMapCB", "expect : after object field, but found "+string([]byte{c}))
			return
		}
		elem := d.elemType.UnsafeNew()
		d.elemDecoder.Decode(ctx, elem, iter)
		d.mapType.UnsafeSetIndex(ptr, key, elem)
	}
	if c != '}' {
		iter.ReportError("ReadMapCB", `expect }, but found `+string([]byte{c}))
	}
}

type numericMapKeyDecoder struct {
	decoder ValDecoder
}

func (d *numericMapKeyDecoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *Iterator) {
	c := iter.nextToken()
	if c != '"' {
		iter.ReportError("ReadMapCB", `expect ", but found `+string([]byte{c}))
		return
	}
	d.decoder.Decode(ctx, ptr, iter)
	c = iter.nextToken()
	if c != '"' {
		iter.ReportError("ReadMapCB", `expect ", but found `+string([]byte{c}))
		return
	}
}

type numericMapKeyEncoder struct {
	encoder ValEncoder
}

func (u *numericMapKeyEncoder) Encode(ctx context.Context, ptr unsafe.Pointer, stream *Stream) {
	stream.writeByte('"')
	u.encoder.Encode(ctx, ptr, stream)
	stream.writeByte('"')
}

func (u *numericMapKeyEncoder) IsEmpty(context.Context, unsafe.Pointer, bool) bool { return false }

type dynamicMapKeyEncoder struct {
	ctx     *ctx
	valType reflect2.Type
}

func (e *dynamicMapKeyEncoder) Encode(ctx context.Context, ptr unsafe.Pointer, stream *Stream) {
	obj := e.valType.UnsafeIndirect(ptr)
	encoderOfMapKey(e.ctx, reflect2.TypeOf(obj)).Encode(ctx, reflect2.PtrOf(obj), stream)
}

func (e *dynamicMapKeyEncoder) IsEmpty(ctx context.Context, ptr unsafe.Pointer, checkZero bool) bool {
	obj := e.valType.UnsafeIndirect(ptr)
	return encoderOfMapKey(e.ctx, reflect2.TypeOf(obj)).IsEmpty(ctx, reflect2.PtrOf(obj), checkZero)
}

type mapEncoder struct {
	mapType     *reflect2.UnsafeMapType
	keyEncoder  ValEncoder
	elemEncoder ValEncoder
	omitempty   bool
}

type sortKeysMapEncoder struct {
	*mapEncoder
}

func (e *mapEncoder) Encode(ctx context.Context, ptr unsafe.Pointer, stream *Stream) {
	if *(*unsafe.Pointer)(ptr) == nil {
		stream.WriteNil()
		return
	}
	stream.WriteObjectStart()
	iter := e.mapType.UnsafeIterate(ptr)
	for i := 0; iter.HasNext(); {
		if i > 0 {
			stream.WriteMore()
		}
		key, elem := iter.UnsafeNext()
		if e.omitempty && e.elemEncoder.IsEmpty(ctx, elem, true) {
			continue
		}

		e.keyEncoder.Encode(ctx, key, stream)
		if stream.indention > 0 {
			stream.write2Bytes(':', ' ')
		} else {
			stream.writeByte(':')
		}
		e.elemEncoder.Encode(ctx, elem, stream)
		i++
	}
	stream.WriteObjectEnd()
}

func (e *mapEncoder) IsEmpty(_ context.Context, ptr unsafe.Pointer, _ bool) bool {
	iter := e.mapType.UnsafeIterate(ptr)
	return !iter.HasNext()
}

func (e *sortKeysMapEncoder) Encode(ctx context.Context, ptr unsafe.Pointer, stream *Stream) {
	if *(*unsafe.Pointer)(ptr) == nil {
		stream.WriteNil()
		return
	}
	stream.WriteObjectStart()
	mapIter := e.mapType.UnsafeIterate(ptr)
	subStream := stream.cfg.BorrowStream(nil)
	subStream.Attachment = stream.Attachment
	subIter := stream.cfg.BorrowIterator(nil)
	keyValues := encodedKvs{}
	for mapIter.HasNext() {
		key, elem := mapIter.UnsafeNext()
		if e.omitempty && e.elemEncoder.IsEmpty(ctx, elem, true) {
			continue
		}

		subStreamIndex := subStream.Buffered()
		e.keyEncoder.Encode(ctx, key, subStream)
		if subStream.Error != nil && subStream.Error != io.EOF && stream.Error == nil {
			stream.Error = subStream.Error
		}
		encodedKey := subStream.Buffer()[subStreamIndex:]
		subIter.ResetBytes(encodedKey)
		decodedKey := subIter.ReadString()
		if stream.indention > 0 {
			subStream.write2Bytes(byte(':'), byte(' '))
		} else {
			subStream.writeByte(':')
		}
		e.elemEncoder.Encode(ctx, elem, subStream)
		keyValues = append(keyValues, encodedKv{
			key: decodedKey,
			val: subStream.Buffer()[subStreamIndex:],
		})
	}
	sort.Sort(keyValues)
	for i, keyValue := range keyValues {
		if i > 0 {
			stream.WriteMore()
		}
		_, _ = stream.Write(keyValue.val)
	}
	if subStream.Error != nil && stream.Error == nil {
		stream.Error = subStream.Error
	}
	stream.WriteObjectEnd()
	stream.cfg.ReturnStream(subStream)
	stream.cfg.ReturnIterator(subIter)
}

func (e *sortKeysMapEncoder) IsEmpty(_ context.Context, ptr unsafe.Pointer, _ bool) bool {
	iter := e.mapType.UnsafeIterate(ptr)
	return !iter.HasNext()
}

type encodedKvs []encodedKv

type encodedKv struct {
	key string
	val []byte
}

func (r encodedKvs) Len() int           { return len(r) }
func (r encodedKvs) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r encodedKvs) Less(i, j int) bool { return r[i].key < r[j].key }
