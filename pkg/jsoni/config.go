package jsoni

import (
	"context"
	"encoding/json"
	"github.com/bingoohuang/gg/pkg/ss"
	"io"
	"reflect"
	"sync"
	"unsafe"

	"github.com/modern-go/concurrent"
	"github.com/modern-go/reflect2"
)

// Config customize how the API should behave.
// The API is created from Config by Froze.
type Config struct {
	IndentionStep                 int
	MarshalFloatWith6Digits       bool
	EscapeHTML                    bool
	SortMapKeys                   bool
	OmitEmptyStructField          bool // omit empty struct field.
	OmitEmptyMapKeys              bool // omit keys whose value is empty.
	UseNumber                     bool
	DisallowUnknownFields         bool
	TagKey                        string
	OnlyTaggedField               bool
	ValidateJsonRawMessage        bool
	ObjectFieldMustBeSimpleString bool
	CaseSensitive                 bool
	Int64AsString                 bool
	NilAsEmpty                    bool
}

// API the public interface of this package.
// Primary Marshal and Unmarshal.
type API interface {
	IteratorPool
	StreamPool
	MarshalToString(ctx context.Context, v interface{}) (string, error)
	Marshal(ctx context.Context, v interface{}) ([]byte, error)
	MarshalIndent(ctx context.Context, v interface{}, prefix, indent string) ([]byte, error)
	UnmarshalFromString(ctx context.Context, str string, v interface{}) error
	Unmarshal(ctx context.Context, data []byte, v interface{}) error
	Get(data []byte, path ...interface{}) Any
	NewEncoder(writer io.Writer) *Encoder
	NewDecoder(reader io.Reader) *Decoder
	Valid(ctx context.Context, data []byte) bool
	RegisterExtension(extension Extension)
	DecoderOf(typ reflect2.Type) ValDecoder
	EncoderOf(typ reflect2.Type) ValEncoder

	RegisterTypeEncoder(typ string, encoder ValEncoder)
	RegisterTypeDecoder(typ string, decoder ValDecoder)
	RegisterTypeEncoderFunc(typ string, fun EncoderFunc, isEmptyFunc IsEmptyFn)
	RegisterTypeDecoderFunc(typ string, fun DecoderFunc)
	RegisterFieldEncoder(typ string, field string, encoder ValEncoder)
	RegisterFieldDecoder(typ string, field string, decoder ValDecoder)
}

// ConfigDefault the default API
var ConfigDefault = Config{
	EscapeHTML: true,
}.Froze()

// ConfigCompatibleWithStandardLibrary tries to be 100% compatible with standard library behavior
var ConfigCompatibleWithStandardLibrary = Config{
	EscapeHTML:             true,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
}.Froze()

// ConfigFastest marshals float with only 6 digits precision
var ConfigFastest = Config{
	EscapeHTML:                    false,
	MarshalFloatWith6Digits:       true, // will lose precession
	ObjectFieldMustBeSimpleString: true, // do not unescape object field
}.Froze()

type frozenConfig struct {
	configBeforeFrozen Config

	indentionStep        int
	sortMapKeys          bool
	omitEmptyStructField bool // omit empty struct field
	omitEmptyMapKeys     bool // omit empty keys whose value is empty

	objectFieldMustBeSimpleString bool

	onlyTaggedField       bool
	disallowUnknownFields bool

	caseSensitive bool
	int64AsString bool
	nilAsEmpty    bool
	clearQuotes   bool // clear Valid JSONObject/JSONArray string without quotes

	decoderCache     *concurrent.Map
	encoderCache     *concurrent.Map
	encoderExtension Extension
	decoderExtension Extension
	extensions       Extensions
	streamPool       *sync.Pool
	iteratorPool     *sync.Pool

	typeDecoders  map[string]ValDecoder
	fieldDecoders map[string]ValDecoder
	typeEncoders  map[string]ValEncoder
	fieldEncoders map[string]ValEncoder
}

func (c *frozenConfig) initCache() {
	c.decoderCache = concurrent.NewMap()
	c.encoderCache = concurrent.NewMap()
}

func (c *frozenConfig) addDecoderToCache(cacheKey uintptr, decoder ValDecoder) {
	c.decoderCache.Store(cacheKey, decoder)
}

func (c *frozenConfig) addEncoderToCache(cacheKey uintptr, encoder ValEncoder) {
	c.encoderCache.Store(cacheKey, encoder)
}

func (c *frozenConfig) getDecoderFromCache(cacheKey uintptr) ValDecoder {
	if decoder, ok := c.decoderCache.Load(cacheKey); ok {
		return decoder.(ValDecoder)
	}
	return nil
}

func (c *frozenConfig) getEncoderFromCache(cacheKey uintptr) ValEncoder {
	encoder, found := c.encoderCache.Load(cacheKey)
	if found {
		return encoder.(ValEncoder)
	}
	return nil
}

var cfgCache = concurrent.NewMap()

func getFrozenConfigFromCache(cfg Config) *frozenConfig {
	if obj, found := cfgCache.Load(cfg); found {
		return obj.(*frozenConfig)
	}
	return nil
}

func addFrozenConfigToCache(cfg Config, frozenConfig *frozenConfig) {
	cfgCache.Store(cfg, frozenConfig)
}

// Froze forge API from config
func (c Config) Froze() API {
	api := &frozenConfig{
		sortMapKeys:                   c.SortMapKeys,
		omitEmptyStructField:          c.OmitEmptyStructField,
		omitEmptyMapKeys:              c.OmitEmptyMapKeys,
		indentionStep:                 c.IndentionStep,
		objectFieldMustBeSimpleString: c.ObjectFieldMustBeSimpleString,
		onlyTaggedField:               c.OnlyTaggedField,
		disallowUnknownFields:         c.DisallowUnknownFields,
		caseSensitive:                 c.CaseSensitive,
		int64AsString:                 c.Int64AsString,
		nilAsEmpty:                    c.NilAsEmpty,

		typeDecoders:  map[string]ValDecoder{},
		fieldDecoders: map[string]ValDecoder{},
		typeEncoders:  map[string]ValEncoder{},
		fieldEncoders: map[string]ValEncoder{},
	}
	api.streamPool = &sync.Pool{New: func() interface{} { return NewStream(api, nil, 512) }}
	api.iteratorPool = &sync.Pool{New: func() interface{} { return NewIterator(api) }}
	api.initCache()
	encoderExtension := EncoderExtension{}
	decoderExtension := DecoderExtension{}
	if c.MarshalFloatWith6Digits {
		api.marshalFloatWith6Digits(encoderExtension)
	}
	if c.EscapeHTML {
		api.escapeHTML(encoderExtension)
	}
	if c.UseNumber {
		api.useNumber(decoderExtension)
	}
	if c.ValidateJsonRawMessage {
		api.validateJsonRawMessage(encoderExtension)
	}
	api.encoderExtension = encoderExtension
	api.decoderExtension = decoderExtension
	api.configBeforeFrozen = c
	return api
}

func (c Config) frozeWithCacheReuse(extraExtensions []Extension) *frozenConfig {
	api := getFrozenConfigFromCache(c)
	if api != nil {
		return api
	}
	api = c.Froze().(*frozenConfig)
	for _, extension := range extraExtensions {
		api.RegisterExtension(extension)
	}
	addFrozenConfigToCache(c, api)
	return api
}

func (c *frozenConfig) validateJsonRawMessage(extension EncoderExtension) {
	encoder := &funcEncoder{fn: func(ctx context.Context, ptr unsafe.Pointer, stream *Stream) {
		rawMessage := *(*json.RawMessage)(ptr)
		iter := c.BorrowIterator(rawMessage)
		defer c.ReturnIterator(iter)
		iter.Read(ctx)
		if iter.Error != nil && iter.Error != io.EOF {
			stream.WriteRaw("null")
		} else {
			stream.WriteRaw(string(rawMessage))
		}
	}, isEmptyFn: func(ctx context.Context, ptr unsafe.Pointer, checkZero bool) bool {
		return len(*((*json.RawMessage)(ptr))) == 0
	}}
	extension[PtrElem((*json.RawMessage)(nil))] = encoder
	extension[PtrElem((*RawMessage)(nil))] = encoder
}

func PtrElem(obj interface{}) reflect2.Type { return reflect2.TypeOfPtr(obj).Elem() }

func (c *frozenConfig) useNumber(extension DecoderExtension) {
	extension[PtrElem((*interface{})(nil))] = &funcDecoder{fun: func(ctx context.Context, ptr unsafe.Pointer, iter *Iterator) {
		exitingValue := *((*interface{})(ptr))
		if exitingValue != nil && reflect.TypeOf(exitingValue).Kind() == reflect.Ptr {
			iter.ReadVal(ctx, exitingValue)
			return
		}
		if iter.WhatIsNext() == NumberValue {
			*((*interface{})(ptr)) = json.Number(iter.readNumberAsString())
		} else {
			*((*interface{})(ptr)) = iter.Read(ctx)
		}
	}}
}
func (c *frozenConfig) getTagKey() string { return ss.Or(c.configBeforeFrozen.TagKey, "json") }

func (c *frozenConfig) RegisterExtension(extension Extension) {
	c.extensions = append(c.extensions, extension)
}

type lossyFloat32Encoder struct{}

func (e *lossyFloat32Encoder) Encode(_ context.Context, ptr unsafe.Pointer, stream *Stream) {
	stream.WriteFloat32Lossy(*((*float32)(ptr)))
}

func (e *lossyFloat32Encoder) IsEmpty(_ context.Context, ptr unsafe.Pointer, _ bool) bool {
	return *((*float32)(ptr)) == 0
}

type lossyFloat64Encoder struct{}

func (e *lossyFloat64Encoder) Encode(_ context.Context, ptr unsafe.Pointer, stream *Stream) {
	stream.WriteFloat64Lossy(*((*float64)(ptr)))
}

func (e *lossyFloat64Encoder) IsEmpty(_ context.Context, ptr unsafe.Pointer, _ bool) bool {
	return *((*float64)(ptr)) == 0
}

// EnableLossyFloatMarshalling keeps 10**(-6) precision
// for float variables for better performance.
func (c *frozenConfig) marshalFloatWith6Digits(extension EncoderExtension) {
	// for better performance
	extension[PtrElem((*float32)(nil))] = &lossyFloat32Encoder{}
	extension[PtrElem((*float64)(nil))] = &lossyFloat64Encoder{}
}

type htmlEscapedStringEncoder struct{}

func (e *htmlEscapedStringEncoder) Encode(ctx context.Context, ptr unsafe.Pointer, stream *Stream) {
	s := *((*string)(ptr))

	if writeRawBytesIfClearQuotes(ctx, s, stream) {
		return
	}

	stream.WriteStringWithHTMLEscaped(s)
}

func (e *htmlEscapedStringEncoder) IsEmpty(_ context.Context, ptr unsafe.Pointer, _ bool) bool {
	return *((*string)(ptr)) == ""
}

func (c *frozenConfig) escapeHTML(encoderExtension EncoderExtension) {
	encoderExtension[PtrElem((*string)(nil))] = &htmlEscapedStringEncoder{}
}

func (c *frozenConfig) MarshalToString(parent context.Context, v interface{}) (string, error) {
	stream := c.BorrowStream(nil)
	defer c.ReturnStream(stream)
	stream.WriteVal(createCfgContext(parent, c), v)
	if stream.Error != nil {
		return "", stream.Error
	}
	return string(stream.Buffer()), nil
}

func (c *frozenConfig) Marshal(parent context.Context, v interface{}) ([]byte, error) {
	stream := c.BorrowStream(nil)
	defer c.ReturnStream(stream)
	stream.WriteVal(createCfgContext(parent, c), v)
	if stream.Error != nil {
		return nil, stream.Error
	}
	result := stream.Buffer()
	copied := make([]byte, len(result))
	copy(copied, result)
	return copied, nil
}

func createCfgContext(parent context.Context, c *frozenConfig) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithValue(parent, ContextCfg, c)
}

func (c *frozenConfig) MarshalIndent(parent context.Context, v interface{}, prefix, indent string) ([]byte, error) {
	if prefix != "" {
		panic("prefix is not supported")
	}
	for _, r := range indent {
		if r != ' ' {
			panic("indent can only be space")
		}
	}
	newCfg := c.configBeforeFrozen
	newCfg.IndentionStep = len(indent)
	return newCfg.frozeWithCacheReuse(c.extensions).Marshal(createCfgContext(parent, c), v)
}

func (c *frozenConfig) UnmarshalFromString(parent context.Context, str string, v interface{}) error {
	data := []byte(str)
	iter := c.BorrowIterator(data)
	defer c.ReturnIterator(iter)
	iter.ReadVal(createCfgContext(parent, c), v)
	if t := iter.nextToken(); t == 0 {
		if iter.Error == io.EOF {
			return nil
		}
		return iter.Error
	}
	iter.ReportError("Unmarshal", "there are bytes left after unmarshal")
	return iter.Error
}

func (c *frozenConfig) Get(data []byte, path ...interface{}) Any {
	iter := c.BorrowIterator(data)
	defer c.ReturnIterator(iter)
	return locatePath(iter, path)
}

func (c *frozenConfig) Unmarshal(parent context.Context, data []byte, v interface{}) error {
	iter := c.BorrowIterator(data)
	defer c.ReturnIterator(iter)
	iter.ReadVal(createCfgContext(parent, c), v)
	if t := iter.nextToken(); t == 0 {
		if iter.Error == io.EOF {
			return nil
		}
		return iter.Error
	}
	iter.ReportError("Unmarshal", "there are bytes left after unmarshal")
	return iter.Error
}

func (c *frozenConfig) NewEncoder(writer io.Writer) *Encoder {
	return &Encoder{stream: NewStream(c, writer, 512)}
}

func (c *frozenConfig) NewDecoder(reader io.Reader) *Decoder {
	return &Decoder{iter: Parse(c, reader, 512)}
}

func (c *frozenConfig) Valid(_ context.Context, data []byte) bool {
	iter := c.BorrowIterator(data)
	defer c.ReturnIterator(iter)
	iter.Skip()
	return iter.Error == nil
}
