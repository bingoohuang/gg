package jsoni

import (
	"fmt"
	"github.com/modern-go/reflect2"
	"io"
	"reflect"
	"sort"
	"unsafe"
)

func decoderOfMap(ctx *ctx, typ reflect2.Type) ValDecoder {
	mapType := typ.(*reflect2.UnsafeMapType)
	keyDecoder := decoderOfMapKey(ctx.append("[mapKey]"), mapType.Key())
	elemDecoder := decoderOfType(ctx.append("[mapElem]"), mapType.Elem())
	return &mapDecoder{
		mapType:     mapType,
		keyType:     mapType.Key(),
		elemType:    mapType.Elem(),
		keyDecoder:  keyDecoder,
		elemDecoder: elemDecoder,
	}
}

func encoderOfMap(ctx *ctx, typ reflect2.Type) ValEncoder {
	mapType := typ.(*reflect2.UnsafeMapType)
	if ctx.sortMapKeys {
		return &sortKeysMapEncoder{
			mapType:     mapType,
			keyEncoder:  encoderOfMapKey(ctx.append("[mapKey]"), mapType.Key()),
			elemEncoder: encoderOfType(ctx.append("[mapElem]"), mapType.Elem()),
		}
	}
	return &mapEncoder{
		mapType:     mapType,
		keyEncoder:  encoderOfMapKey(ctx.append("[mapKey]"), mapType.Key()),
		elemEncoder: encoderOfType(ctx.append("[mapElem]"), mapType.Elem()),
	}
}

func decoderOfMapKey(ctx *ctx, typ reflect2.Type) ValDecoder {
	decoder := ctx.decoderExtension.CreateMapKeyDecoder(typ)
	if decoder != nil {
		return decoder
	}
	for _, extension := range ctx.extraExtensions {
		decoder := extension.CreateMapKeyDecoder(typ)
		if decoder != nil {
			return decoder
		}
	}

	ptrType := reflect2.PtrTo(typ)
	if ptrType.Implements(unmarshalerType) {
		return &referenceDecoder{
			&unmarshalerDecoder{
				valType: ptrType,
			},
		}
	}
	if typ.Implements(unmarshalerType) {
		return &unmarshalerDecoder{
			valType: typ,
		}
	}
	if ptrType.Implements(textUnmarshalerType) {
		return &referenceDecoder{
			&textUnmarshalerDecoder{
				valType: ptrType,
			},
		}
	}
	if typ.Implements(textUnmarshalerType) {
		return &textUnmarshalerDecoder{
			valType: typ,
		}
	}

	switch typ.Kind() {
	case reflect.String:
		return decoderOfType(ctx, reflect2.DefaultTypeOfKind(reflect.String))
	case reflect.Bool,
		reflect.Uint8, reflect.Int8,
		reflect.Uint16, reflect.Int16,
		reflect.Uint32, reflect.Int32,
		reflect.Uint64, reflect.Int64,
		reflect.Uint, reflect.Int,
		reflect.Float32, reflect.Float64,
		reflect.Uintptr:
		typ = reflect2.DefaultTypeOfKind(typ.Kind())
		return &numericMapKeyDecoder{decoderOfType(ctx, typ)}
	default:
		return &lazyErrorDecoder{err: fmt.Errorf("unsupported map key type: %v", typ)}
	}
}

func encoderOfMapKey(ctx *ctx, typ reflect2.Type) ValEncoder {
	encoder := ctx.encoderExtension.CreateMapKeyEncoder(typ)
	if encoder != nil {
		return encoder
	}
	for _, extension := range ctx.extraExtensions {
		encoder := extension.CreateMapKeyEncoder(typ)
		if encoder != nil {
			return encoder
		}
	}

	if typ == textMarshalerType {
		return &directTextMarshalerEncoder{
			stringEncoder: ctx.EncoderOf(reflect2.TypeOf("")),
		}
	}
	if typ.Implements(textMarshalerType) {
		return &textMarshalerEncoder{
			valType:       typ,
			stringEncoder: ctx.EncoderOf(reflect2.TypeOf("")),
		}
	}

	switch typ.Kind() {
	case reflect.String:
		return encoderOfType(ctx, reflect2.DefaultTypeOfKind(reflect.String))
	case reflect.Bool,
		reflect.Uint8, reflect.Int8,
		reflect.Uint16, reflect.Int16,
		reflect.Uint32, reflect.Int32,
		reflect.Uint64, reflect.Int64,
		reflect.Uint, reflect.Int,
		reflect.Float32, reflect.Float64,
		reflect.Uintptr:
		typ = reflect2.DefaultTypeOfKind(typ.Kind())
		return &numericMapKeyEncoder{encoderOfType(ctx, typ)}
	default:
		if typ.Kind() == reflect.Interface {
			return &dynamicMapKeyEncoder{ctx, typ}
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

func (d *mapDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	mapType := d.mapType
	c := iter.nextToken()
	if c == 'n' {
		iter.skipThreeBytes('u', 'l', 'l')
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
	d.keyDecoder.Decode(key, iter)
	c = iter.nextToken()
	if c != ':' {
		iter.ReportError("ReadMapCB", "expect : after object field, but found "+string([]byte{c}))
		return
	}
	elem := d.elemType.UnsafeNew()
	d.elemDecoder.Decode(elem, iter)
	d.mapType.UnsafeSetIndex(ptr, key, elem)
	for c = iter.nextToken(); c == ','; c = iter.nextToken() {
		key := d.keyType.UnsafeNew()
		d.keyDecoder.Decode(key, iter)
		c = iter.nextToken()
		if c != ':' {
			iter.ReportError("ReadMapCB", "expect : after object field, but found "+string([]byte{c}))
			return
		}
		elem := d.elemType.UnsafeNew()
		d.elemDecoder.Decode(elem, iter)
		d.mapType.UnsafeSetIndex(ptr, key, elem)
	}
	if c != '}' {
		iter.ReportError("ReadMapCB", `expect }, but found `+string([]byte{c}))
	}
}

type numericMapKeyDecoder struct {
	decoder ValDecoder
}

func (d *numericMapKeyDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	c := iter.nextToken()
	if c != '"' {
		iter.ReportError("ReadMapCB", `expect ", but found `+string([]byte{c}))
		return
	}
	d.decoder.Decode(ptr, iter)
	c = iter.nextToken()
	if c != '"' {
		iter.ReportError("ReadMapCB", `expect ", but found `+string([]byte{c}))
		return
	}
}

type numericMapKeyEncoder struct {
	encoder ValEncoder
}

func (u *numericMapKeyEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	stream.writeByte('"')
	u.encoder.Encode(ptr, stream)
	stream.writeByte('"')
}

func (u *numericMapKeyEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return false
}

type dynamicMapKeyEncoder struct {
	ctx     *ctx
	valType reflect2.Type
}

func (e *dynamicMapKeyEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	obj := e.valType.UnsafeIndirect(ptr)
	encoderOfMapKey(e.ctx, reflect2.TypeOf(obj)).Encode(reflect2.PtrOf(obj), stream)
}

func (e *dynamicMapKeyEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	obj := e.valType.UnsafeIndirect(ptr)
	return encoderOfMapKey(e.ctx, reflect2.TypeOf(obj)).IsEmpty(reflect2.PtrOf(obj))
}

type mapEncoder struct {
	mapType     *reflect2.UnsafeMapType
	keyEncoder  ValEncoder
	elemEncoder ValEncoder
}

func (e *mapEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	if *(*unsafe.Pointer)(ptr) == nil {
		stream.WriteNil()
		return
	}
	stream.WriteObjectStart()
	iter := e.mapType.UnsafeIterate(ptr)
	for i := 0; iter.HasNext(); i++ {
		if i != 0 {
			stream.WriteMore()
		}
		key, elem := iter.UnsafeNext()
		e.keyEncoder.Encode(key, stream)
		if stream.indention > 0 {
			stream.writeTwoBytes(byte(':'), byte(' '))
		} else {
			stream.writeByte(':')
		}
		e.elemEncoder.Encode(elem, stream)
	}
	stream.WriteObjectEnd()
}

func (e *mapEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	iter := e.mapType.UnsafeIterate(ptr)
	return !iter.HasNext()
}

type sortKeysMapEncoder struct {
	mapType     *reflect2.UnsafeMapType
	keyEncoder  ValEncoder
	elemEncoder ValEncoder
}

func (e *sortKeysMapEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	if *(*unsafe.Pointer)(ptr) == nil {
		stream.WriteNil()
		return
	}
	stream.WriteObjectStart()
	mapIter := e.mapType.UnsafeIterate(ptr)
	subStream := stream.cfg.BorrowStream(nil)
	subStream.Attachment = stream.Attachment
	subIter := stream.cfg.BorrowIterator(nil)
	keyValues := encodedKeyValues{}
	for mapIter.HasNext() {
		key, elem := mapIter.UnsafeNext()
		subStreamIndex := subStream.Buffered()
		e.keyEncoder.Encode(key, subStream)
		if subStream.Error != nil && subStream.Error != io.EOF && stream.Error == nil {
			stream.Error = subStream.Error
		}
		encodedKey := subStream.Buffer()[subStreamIndex:]
		subIter.ResetBytes(encodedKey)
		decodedKey := subIter.ReadString()
		if stream.indention > 0 {
			subStream.writeTwoBytes(byte(':'), byte(' '))
		} else {
			subStream.writeByte(':')
		}
		e.elemEncoder.Encode(elem, subStream)
		keyValues = append(keyValues, encodedKV{
			key:      decodedKey,
			keyValue: subStream.Buffer()[subStreamIndex:],
		})
	}
	sort.Sort(keyValues)
	for i, keyValue := range keyValues {
		if i != 0 {
			stream.WriteMore()
		}
		stream.Write(keyValue.keyValue)
	}
	if subStream.Error != nil && stream.Error == nil {
		stream.Error = subStream.Error
	}
	stream.WriteObjectEnd()
	stream.cfg.ReturnStream(subStream)
	stream.cfg.ReturnIterator(subIter)
}

func (e *sortKeysMapEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	iter := e.mapType.UnsafeIterate(ptr)
	return !iter.HasNext()
}

type encodedKeyValues []encodedKV

type encodedKV struct {
	key      string
	keyValue []byte
}

func (r encodedKeyValues) Len() int           { return len(r) }
func (r encodedKeyValues) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r encodedKeyValues) Less(i, j int) bool { return r[i].key < r[j].key }
