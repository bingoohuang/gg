package jsoni

import (
	"context"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/modern-go/reflect2"
)

// ValDecoder is an internal type registered to cache as needed.
// Don't confuse jsoni.ValDecoder with json.Decoder.
// For json.Decoder's adapter, refer to jsoni.AdapterDecoder(todo link).
//
// Reflection on type to create decoders, which is then cached
// Reflection on value is avoided as we can, as the reflect.Value itself will allocate, with following exceptions
// 1. create instance of new value, for example *int will need a int to be allocated
// 2. append to slice, if the existing cap is not enough, allocate will be done using Reflect.New
// 3. assignment to map, both key and value will be reflect.Value
// For a simple struct binding, it will be reflect.Value free and allocation free
type ValDecoder interface {
	Decode(ctx context.Context, ptr unsafe.Pointer, iter *Iterator)
}

// ValEncoder is an internal type registered to cache as needed.
// Don't confuse jsoni.ValEncoder with json.Encoder.
// For json.Encoder's adapter, refer to jsoni.AdapterEncoder(todo godoc link).
type ValEncoder interface {
	IsEmpty(ctx context.Context, ptr unsafe.Pointer, checkZero bool) bool
	Encode(ctx context.Context, ptr unsafe.Pointer, stream *Stream)
}

type checkIsEmpty interface {
	IsEmpty(ctx context.Context, ptr unsafe.Pointer, checkZero bool) bool
}

type ctx struct {
	*frozenConfig
	prefix   string
	encoders map[reflect2.Type]ValEncoder
	decoders map[reflect2.Type]ValDecoder
}

func (b *ctx) caseSensitive() bool {
	if b.frozenConfig == nil {
		// default is case-insensitive
		return false
	}
	return b.frozenConfig.caseSensitive
}

func (b *ctx) append(prefix string) *ctx {
	return &ctx{
		frozenConfig: b.frozenConfig,
		prefix:       b.prefix + " " + prefix,
		encoders:     b.encoders,
		decoders:     b.decoders,
	}
}

// ReadVal copy the underlying JSON into go interface, same as json.Unmarshal
func (iter *Iterator) ReadVal(ctx context.Context, obj interface{}) {
	depth := iter.depth
	cacheKey := reflect2.RTypeOf(obj)
	decoder := iter.cfg.getDecoderFromCache(cacheKey)
	if decoder == nil {
		typ := reflect2.TypeOf(obj)
		if typ == nil || typ.Kind() != reflect.Ptr {
			iter.ReportError("ReadVal", "can only unmarshal into pointer")
			return
		}
		decoder = iter.cfg.DecoderOf(typ)
	}
	ptr := reflect2.PtrOf(obj)
	if ptr == nil {
		iter.ReportError("ReadVal", "can not read into nil pointer")
		return
	}
	decoder.Decode(ctx, ptr, iter)
	if iter.depth != depth {
		iter.ReportError("ReadVal", "unexpected mismatched nesting")
		return
	}
}

// WriteVal copy the go interface into underlying JSON, same as json.Marshal
func (s *Stream) WriteVal(ctx context.Context, val interface{}) {
	if nil == val {
		s.WriteNil()
		return
	}
	cacheKey := reflect2.RTypeOf(val)
	encoder := s.cfg.getEncoderFromCache(cacheKey)
	if encoder == nil {
		typ := reflect2.TypeOf(val)
		encoder = s.cfg.EncoderOf(typ)
	}
	encoder.Encode(ctx, reflect2.PtrOf(val), s)
}

func (c *frozenConfig) DecoderOf(typ reflect2.Type) ValDecoder {
	cacheKey := typ.RType()
	if d := c.getDecoderFromCache(cacheKey); d != nil {
		return d
	}
	ct := &ctx{
		frozenConfig: c,
		decoders:     map[reflect2.Type]ValDecoder{},
		encoders:     map[reflect2.Type]ValEncoder{},
	}
	ptrType := typ.(*reflect2.UnsafePtrType)
	decoder := decoderOfType(ct, ptrType.Elem())
	c.addDecoderToCache(cacheKey, decoder)
	return decoder
}

func decoderOfType(ctx *ctx, typ reflect2.Type) ValDecoder {
	if d := getTypeDecoderFromExtension(ctx, typ); d != nil {
		return d
	}
	decoder := createDecoderOfType(ctx, typ)
	decoder = ctx.frozenConfig.extensions.decorateDecoder(typ, decoder)
	decoder = ctx.decoderExtension.DecorateDecoder(typ, decoder)
	decoder = ctx.extensions.decorateDecoder(typ, decoder)
	return decoder
}

func createDecoderOfType(ctx *ctx, typ reflect2.Type) ValDecoder {
	if d := ctx.decoders[typ]; d != nil {
		return d
	}
	placeholder := &placeholderDecoder{}
	ctx.decoders[typ] = placeholder
	decoder := _createDecoderOfType(ctx, typ)
	placeholder.decoder = decoder
	return decoder
}

func _createDecoderOfType(ctx *ctx, typ reflect2.Type) ValDecoder {
	if d := createDecoderOfJsonRawMessage(ctx, typ); d != nil {
		return d
	}
	if d := createDecoderOfJsonNumber(ctx, typ); d != nil {
		return d
	}
	if d := createDecoderOfMarshaler(ctx, typ); d != nil {
		return d
	}
	if d := createDecoderOfAny(ctx, typ); d != nil {
		return d
	}
	if d := createDecoderOfNative(ctx, typ); d != nil {
		return d
	}
	switch typ.Kind() {
	case reflect.Interface:
		if ifaceType, ok := typ.(*reflect2.UnsafeIFaceType); ok {
			return &ifaceDecoder{valType: ifaceType}
		}
		return &efaceDecoder{}
	case reflect.Struct:
		return decoderOfStruct(ctx, typ)
	case reflect.Array:
		return decoderOfArray(ctx, typ)
	case reflect.Slice:
		return decoderOfSlice(ctx, typ)
	case reflect.Map:
		return decoderOfMap(ctx, typ)
	case reflect.Ptr:
		return decoderOfOptional(ctx, typ)
	default:
		return &lazyErrorDecoder{err: fmt.Errorf("%s%s is unsupported type", ctx.prefix, typ.String())}
	}
}

func (c *frozenConfig) EncoderOf(typ reflect2.Type) ValEncoder {
	cacheKey := typ.RType()
	if encoder := c.getEncoderFromCache(cacheKey); encoder != nil {
		return encoder
	}
	encoder := encoderOfType(&ctx{
		frozenConfig: c,
		decoders:     map[reflect2.Type]ValDecoder{},
		encoders:     map[reflect2.Type]ValEncoder{},
	}, typ)
	if typ.LikePtr() {
		encoder = &onePtrEncoder{encoder}
	}
	c.addEncoderToCache(cacheKey, encoder)
	return encoder
}

type onePtrEncoder struct {
	encoder ValEncoder
}

func (e *onePtrEncoder) IsEmpty(ctx context.Context, p unsafe.Pointer, checkZero bool) bool {
	return e.encoder.IsEmpty(ctx, unsafe.Pointer(&p), checkZero)
}

func (e *onePtrEncoder) Encode(ctx context.Context, p unsafe.Pointer, s *Stream) {
	e.encoder.Encode(ctx, unsafe.Pointer(&p), s)
}

func encoderOfType(ctx *ctx, typ reflect2.Type) ValEncoder {
	if encoder := getTypeEncoderFromExtension(ctx, typ); encoder != nil {
		return encoder
	}
	encoder := createEncoderOfType(ctx, typ)
	encoder = ctx.frozenConfig.extensions.decorateEncoder(typ, encoder)
	encoder = ctx.encoderExtension.DecorateEncoder(typ, encoder)
	encoder = ctx.extensions.decorateEncoder(typ, encoder)
	return encoder
}

func createEncoderOfType(ctx *ctx, typ reflect2.Type) ValEncoder {
	encoder := ctx.encoders[typ]
	if encoder != nil {
		return encoder
	}
	placeholder := &placeholderEncoder{}
	ctx.encoders[typ] = placeholder
	encoder = _createEncoderOfType(ctx, typ)
	placeholder.encoder = encoder
	return encoder
}

func _createEncoderOfType(ctx *ctx, typ reflect2.Type) ValEncoder {
	if v := createEncoderOfJsonRawMessage(ctx, typ); v != nil {
		return v
	}
	if v := createEncoderOfJsonNumber(ctx, typ); v != nil {
		return v
	}
	if v := createEncoderOfMarshaler(ctx, typ); v != nil {
		return v
	}
	if v := createEncoderOfAny(ctx, typ); v != nil {
		return v
	}
	if v := createEncoderOfNative(ctx, typ); v != nil {
		return v
	}

	switch kind := typ.Kind(); kind {
	case reflect.Interface:
		return &dynamicEncoder{valType: typ}
	case reflect.Struct:
		return encoderOfStruct(ctx, typ)
	case reflect.Array:
		return encoderOfArray(ctx, typ)
	case reflect.Slice:
		return encoderOfSlice(ctx, typ)
	case reflect.Map:
		return encoderOfMap(ctx, typ)
	case reflect.Ptr:
		return encoderOfOptional(ctx, typ)
	default:
		return &lazyErrorEncoder{err: fmt.Errorf("%s%s is unsupported type", ctx.prefix, typ.String())}
	}
}

type lazyErrorDecoder struct {
	err error
}

func (d *lazyErrorDecoder) Decode(_ context.Context, _ unsafe.Pointer, iter *Iterator) {
	if iter.WhatIsNext() != NilValue {
		if iter.Error == nil {
			iter.Error = d.err
		}
	} else {
		iter.Skip()
	}
}

type lazyErrorEncoder struct {
	err error
}

func (e *lazyErrorEncoder) Encode(_ context.Context, ptr unsafe.Pointer, stream *Stream) {
	if ptr == nil {
		stream.WriteNil()
	} else if stream.Error == nil {
		stream.Error = e.err
	}
}

func (e *lazyErrorEncoder) IsEmpty(context.Context, unsafe.Pointer, bool) bool { return false }

type placeholderDecoder struct {
	decoder ValDecoder
}

func (d *placeholderDecoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *Iterator) {
	d.decoder.Decode(ctx, ptr, iter)
}

type placeholderEncoder struct {
	encoder ValEncoder
}

func (e *placeholderEncoder) Encode(ctx context.Context, p unsafe.Pointer, s *Stream) {
	e.encoder.Encode(ctx, p, s)
}

func (e *placeholderEncoder) IsEmpty(ctx context.Context, p unsafe.Pointer, checkZero bool) bool {
	return e.encoder.IsEmpty(ctx, p, checkZero)
}
