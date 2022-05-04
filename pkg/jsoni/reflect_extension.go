package jsoni

import (
	"context"
	"fmt"
	"github.com/modern-go/reflect2"
	"reflect"
	"sort"
	"strings"
	"unicode"
	"unsafe"
)

type Extensions []Extension

func (es Extensions) UpdateStructDescriptor(structDescriptor *StructDescriptor) {
	for _, extension := range es {
		extension.UpdateStructDescriptor(structDescriptor)
	}
}
func (es Extensions) CreateMapKeyEncoder(typ reflect2.Type) ValEncoder {
	for _, extension := range es {
		if v := extension.CreateMapKeyEncoder(typ); v != nil {
			return v
		}
	}
	return nil
}

func (es Extensions) CreateMapKeyDecoder(typ reflect2.Type) ValDecoder {
	for _, extension := range es {
		if v := extension.CreateMapKeyDecoder(typ); v != nil {
			return v
		}
	}
	return nil
}

func (es Extensions) createEncoder(typ reflect2.Type) ValEncoder {
	for _, extension := range es {
		if e := extension.CreateEncoder(typ); e != nil {
			return e
		}
	}
	return nil
}

func (es Extensions) createDecoder(typ reflect2.Type) ValDecoder {
	for _, extension := range es {
		if d := extension.CreateDecoder(typ); d != nil {
			return d
		}
	}
	return nil
}

func (es Extensions) decorateEncoder(typ reflect2.Type, encoder ValEncoder) ValEncoder {
	for _, extension := range es {
		encoder = extension.DecorateEncoder(typ, encoder)
	}
	return encoder
}

func (es Extensions) decorateDecoder(typ reflect2.Type, decoder ValDecoder) ValDecoder {
	for _, extension := range es {
		decoder = extension.DecorateDecoder(typ, decoder)
	}
	return decoder
}

// StructDescriptor describe how should we encode/decode the struct
type StructDescriptor struct {
	Type   reflect2.Type
	Fields []*Binding
}

// GetField get one field from the descriptor by its name.
// Can not use map here to keep field orders.
func (structDescriptor *StructDescriptor) GetField(fieldName string) *Binding {
	for _, binding := range structDescriptor.Fields {
		if binding.Field.Name() == fieldName {
			return binding
		}
	}
	return nil
}

// Binding describe how should we encode/decode the struct field
type Binding struct {
	levels    []int
	Field     reflect2.StructField
	FromNames []string
	ToNames   []string
	Encoder   ValEncoder
	Decoder   ValDecoder
}

// Extension the one for all SPI. Customize encoding/decoding by specifying alternate encoder/decoder.
// Can also rename fields by UpdateStructDescriptor.

type Extension interface {
	UpdateStructDescriptor(structDescriptor *StructDescriptor)
	CreateMapKeyDecoder(typ reflect2.Type) ValDecoder
	CreateMapKeyEncoder(typ reflect2.Type) ValEncoder
	CreateDecoder(typ reflect2.Type) ValDecoder
	CreateEncoder(typ reflect2.Type) ValEncoder

	DecorateDecoder(typ reflect2.Type, decoder ValDecoder) ValDecoder
	DecorateEncoder(typ reflect2.Type, encoder ValEncoder) ValEncoder
}

// DummyExtension embed this type get dummy implementation for all methods of Extension
type DummyExtension struct{}

// UpdateStructDescriptor No-op
func (e *DummyExtension) UpdateStructDescriptor(*StructDescriptor) {}

// CreateMapKeyDecoder No-op
func (e *DummyExtension) CreateMapKeyDecoder(reflect2.Type) ValDecoder { return nil }

// CreateMapKeyEncoder No-op
func (e *DummyExtension) CreateMapKeyEncoder(reflect2.Type) ValEncoder { return nil }

// CreateDecoder No-op
func (e *DummyExtension) CreateDecoder(reflect2.Type) ValDecoder { return nil }

// CreateEncoder No-op
func (e *DummyExtension) CreateEncoder(reflect2.Type) ValEncoder { return nil }

// DecorateDecoder No-op
func (e *DummyExtension) DecorateDecoder(_ reflect2.Type, v ValDecoder) ValDecoder { return v }

// DecorateEncoder No-op
func (e *DummyExtension) DecorateEncoder(_ reflect2.Type, v ValEncoder) ValEncoder { return v }

type EncoderExtension map[reflect2.Type]ValEncoder

// UpdateStructDescriptor No-op
func (e EncoderExtension) UpdateStructDescriptor(*StructDescriptor) {}

// CreateDecoder No-op
func (e EncoderExtension) CreateDecoder(reflect2.Type) ValDecoder { return nil }

// CreateEncoder get encoder from map
func (e EncoderExtension) CreateEncoder(typ reflect2.Type) ValEncoder { return e[typ] }

// CreateMapKeyDecoder No-op
func (e EncoderExtension) CreateMapKeyDecoder(reflect2.Type) ValDecoder { return nil }

// CreateMapKeyEncoder No-op
func (e EncoderExtension) CreateMapKeyEncoder(reflect2.Type) ValEncoder { return nil }

// DecorateDecoder No-op
func (e EncoderExtension) DecorateDecoder(_ reflect2.Type, v ValDecoder) ValDecoder { return v }

// DecorateEncoder No-op
func (e EncoderExtension) DecorateEncoder(_ reflect2.Type, v ValEncoder) ValEncoder { return v }

type DecoderExtension map[reflect2.Type]ValDecoder

// UpdateStructDescriptor No-op
func (e DecoderExtension) UpdateStructDescriptor(*StructDescriptor) {}

// CreateMapKeyDecoder No-op
func (e DecoderExtension) CreateMapKeyDecoder(reflect2.Type) ValDecoder { return nil }

// CreateMapKeyEncoder No-op
func (e DecoderExtension) CreateMapKeyEncoder(reflect2.Type) ValEncoder { return nil }

// CreateDecoder get decoder from map
func (e DecoderExtension) CreateDecoder(typ reflect2.Type) ValDecoder { return e[typ] }

// CreateEncoder No-op
func (e DecoderExtension) CreateEncoder(reflect2.Type) ValEncoder { return nil }

// DecorateDecoder No-op
func (e DecoderExtension) DecorateDecoder(_ reflect2.Type, v ValDecoder) ValDecoder { return v }

// DecorateEncoder No-op
func (e DecoderExtension) DecorateEncoder(_ reflect2.Type, v ValEncoder) ValEncoder { return v }

type funcDecoder struct {
	fun DecoderFunc
}

func (f *funcDecoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *Iterator) {
	f.fun(ctx, ptr, iter)
}

type funcEncoder struct {
	fn        EncoderFunc
	isEmptyFn IsEmptyFn
}

func (e *funcEncoder) Encode(ctx context.Context, p unsafe.Pointer, stream *Stream) {
	e.fn(ctx, p, stream)
}
func (e *funcEncoder) IsEmpty(ctx context.Context, p unsafe.Pointer, checkZero bool) bool {
	return e.isEmptyFn != nil && e.isEmptyFn(ctx, p, checkZero)
}

type IsEmptyFn func(ctx context.Context, ptr unsafe.Pointer, checkZero bool) bool

// DecoderFunc the function form of TypeDecoder
type DecoderFunc func(ctx context.Context, ptr unsafe.Pointer, iter *Iterator)

// EncoderFunc the function form of TypeEncoder
type EncoderFunc func(ctx context.Context, ptr unsafe.Pointer, stream *Stream)

// RegisterTypeDecoderFunc register TypeDecoder for a type with function
func RegisterTypeDecoderFunc(typ string, fun DecoderFunc) {
	(ConfigDefault.(*frozenConfig)).RegisterTypeDecoderFunc(typ, fun)
}

// RegisterTypeDecoderFunc register TypeDecoder for a type with function
func (c *frozenConfig) RegisterTypeDecoderFunc(typ string, fun DecoderFunc) {
	c.typeDecoders[typ] = &funcDecoder{fun}
}

// RegisterTypeDecoder register TypeDecoder for a typ.
func RegisterTypeDecoder(typ string, decoder ValDecoder) {
	(ConfigDefault.(*frozenConfig)).RegisterTypeDecoder(typ, decoder)
}

// RegisterTypeDecoder register TypeDecoder for a typ.
func (c *frozenConfig) RegisterTypeDecoder(typ string, decoder ValDecoder) {
	c.typeDecoders[typ] = decoder
}

// RegisterFieldDecoderFunc register TypeDecoder for a struct field with function
func RegisterFieldDecoderFunc(typ string, field string, fun DecoderFunc) {
	RegisterFieldDecoder(typ, field, &funcDecoder{fun})
}

// RegisterFieldDecoder register TypeDecoder for a struct field
func RegisterFieldDecoder(typ string, field string, decoder ValDecoder) {
	(ConfigDefault.(*frozenConfig)).RegisterFieldDecoder(typ, field, decoder)
}

// RegisterFieldDecoder register TypeDecoder for a struct field
func (c *frozenConfig) RegisterFieldDecoder(typ string, field string, decoder ValDecoder) {
	c.fieldDecoders[fmt.Sprintf("%s/%s", typ, field)] = decoder
}

// RegisterTypeEncoderFunc register TypeEncoder for a type with encode/isEmpty function
func RegisterTypeEncoderFunc(typ string, fun EncoderFunc, isEmptyFunc IsEmptyFn) {
	(ConfigDefault.(*frozenConfig)).RegisterTypeEncoderFunc(typ, fun, isEmptyFunc)
}

// RegisterTypeEncoderFunc register TypeEncoder for a type with encode/isEmpty function
func (c *frozenConfig) RegisterTypeEncoderFunc(typ string, fun EncoderFunc, isEmptyFunc IsEmptyFn) {
	c.typeEncoders[typ] = &funcEncoder{fn: fun, isEmptyFn: isEmptyFunc}
}

// RegisterTypeEncoder register TypeEncoder for a type
func RegisterTypeEncoder(typ string, encoder ValEncoder) {
	(ConfigDefault.(*frozenConfig)).RegisterTypeEncoder(typ, encoder)
}

// RegisterTypeEncoder register TypeEncoder for a type
func (c *frozenConfig) RegisterTypeEncoder(typ string, encoder ValEncoder) {
	c.typeEncoders[typ] = encoder
}

// RegisterFieldEncoderFunc register TypeEncoder for a struct field with encode/isEmpty function
func RegisterFieldEncoderFunc(typ string, field string, fun EncoderFunc, isEmptyFunc IsEmptyFn) {
	RegisterFieldEncoder(typ, field, &funcEncoder{fn: fun, isEmptyFn: isEmptyFunc})
}

// RegisterFieldEncoder register TypeEncoder for a struct field
func RegisterFieldEncoder(typ string, field string, encoder ValEncoder) {
	(ConfigDefault.(*frozenConfig)).RegisterFieldEncoder(typ, field, encoder)
}

// RegisterFieldEncoder register TypeEncoder for a struct field
func (c *frozenConfig) RegisterFieldEncoder(typ string, field string, encoder ValEncoder) {
	c.fieldEncoders[fmt.Sprintf("%s/%s", typ, field)] = encoder
}

// RegisterExtension register extension
func RegisterExtension(extension Extension) {
	(ConfigDefault.(*frozenConfig)).RegisterExtension(extension)
}

func getTypeDecoderFromExtension(ctx *ctx, typ reflect2.Type) ValDecoder {
	d := _getTypeDecoderFromExtension(ctx, typ)
	if d == nil {
		return nil
	}

	d = ctx.frozenConfig.extensions.decorateDecoder(typ, d)
	d = ctx.decoderExtension.DecorateDecoder(typ, d)
	d = ctx.extensions.decorateDecoder(typ, d)

	return d
}
func _getTypeDecoderFromExtension(ctx *ctx, typ reflect2.Type) ValDecoder {
	if d := ctx.frozenConfig.extensions.createDecoder(typ); d != nil {
		return d
	}
	if d := ctx.decoderExtension.CreateDecoder(typ); d != nil {
		return d
	}
	if d := ctx.extensions.createDecoder(typ); d != nil {
		return d
	}
	if d := ctx.frozenConfig.typeDecoders[typ.String()]; d != nil {
		return d
	}
	if typ.Kind() == reflect.Ptr {
		ptrType := typ.(*reflect2.UnsafePtrType)
		if d := ctx.frozenConfig.typeDecoders[ptrType.Elem().String()]; d != nil {
			return &OptionalDecoder{ptrType.Elem(), d}
		}
	}
	return nil
}

func getTypeEncoderFromExtension(ctx *ctx, typ reflect2.Type) ValEncoder {
	e := _getTypeEncoderFromExtension(ctx, typ)
	if e == nil {
		return nil
	}

	e = ctx.frozenConfig.extensions.decorateEncoder(typ, e)
	e = ctx.encoderExtension.DecorateEncoder(typ, e)
	e = ctx.extensions.decorateEncoder(typ, e)

	return e
}

func _getTypeEncoderFromExtension(ctx *ctx, typ reflect2.Type) ValEncoder {
	if e := ctx.frozenConfig.extensions.createEncoder(typ); e != nil {
		return e
	}
	if e := ctx.encoderExtension.CreateEncoder(typ); e != nil {
		return e
	}
	if e := ctx.extensions.createEncoder(typ); e != nil {
		return e
	}
	if e := ctx.frozenConfig.typeEncoders[typ.String()]; e != nil {
		return e
	}
	if typ.Kind() == reflect.Ptr {
		if e := ctx.frozenConfig.typeEncoders[typ.(*reflect2.UnsafePtrType).Elem().String()]; e != nil {
			return &OptionalEncoder{e}
		}
	}
	return nil
}

func describeStruct(ctx *ctx, typ reflect2.Type) *StructDescriptor {
	var embeddedBindings []*Binding
	var bindings []*Binding
	structType := typ.(*reflect2.UnsafeStructType)
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		tag, hasTag := field.Tag().Lookup(ctx.getTagKey())
		if ctx.onlyTaggedField && !hasTag && !field.Anonymous() {
			continue
		}
		if tag == "-" || field.Name() == "_" {
			continue
		}
		tagParts := strings.Split(tag, ",")
		if field.Anonymous() && (tag == "" || tagParts[0] == "") {
			switch kind := field.Type().Kind(); kind {
			case reflect.Struct:
				structDescriptor := describeStruct(ctx, field.Type())
				for _, b := range structDescriptor.Fields {
					b.levels = append([]int{i}, b.levels...)
					fieldEncoder := b.Encoder.(*structFieldEncoder)
					omitempty := fieldEncoder.omitempty
					nilAsEmpty := fieldEncoder.nilAsEmpty
					b.Encoder = &structFieldEncoder{field: field, fieldEncoder: b.Encoder, omitempty: omitempty, nilAsEmpty: nilAsEmpty}
					b.Decoder = &structFieldDecoder{field: field, fieldDecoder: b.Decoder}
					embeddedBindings = append(embeddedBindings, b)
				}
				continue
			case reflect.Ptr:
				ptrType := field.Type().(*reflect2.UnsafePtrType)
				if ptrType.Elem().Kind() == reflect.Struct {
					structDescriptor := describeStruct(ctx, ptrType.Elem())
					for _, b := range structDescriptor.Fields {
						b.levels = append([]int{i}, b.levels...)
						fieldEncoder := b.Encoder.(*structFieldEncoder)
						omitempty := fieldEncoder.omitempty
						nilAsEmpty := fieldEncoder.nilAsEmpty
						b.Encoder = &dereferenceEncoder{ValueEncoder: b.Encoder}
						b.Encoder = &structFieldEncoder{field: field, fieldEncoder: b.Encoder, omitempty: omitempty, nilAsEmpty: nilAsEmpty}
						b.Decoder = &dereferenceDecoder{valueType: ptrType.Elem(), valueDecoder: b.Decoder}
						b.Decoder = &structFieldDecoder{field: field, fieldDecoder: b.Decoder}
						embeddedBindings = append(embeddedBindings, b)
					}
					continue
				}
			}
		}
		fieldNames := calcFieldNames(field.Name(), tagParts[0], tag)
		fieldCacheKey := fmt.Sprintf("%s/%s", typ.String(), field.Name())
		decoder := ctx.frozenConfig.fieldDecoders[fieldCacheKey]
		if decoder == nil {
			decoder = decoderOfType(ctx.append(field.Name()), field.Type())
		}
		encoder := ctx.frozenConfig.fieldEncoders[fieldCacheKey]
		if encoder == nil {
			encoder = encoderOfType(ctx.append(field.Name()), field.Type())
		}
		binding := &Binding{Field: field, FromNames: fieldNames, ToNames: fieldNames, Decoder: decoder, Encoder: encoder}
		binding.levels = []int{i}
		bindings = append(bindings, binding)
	}
	return createStructDescriptor(ctx, typ, bindings, embeddedBindings)
}
func createStructDescriptor(ctx *ctx, typ reflect2.Type, bindings []*Binding, embeddedBindings []*Binding) *StructDescriptor {
	structDescriptor := &StructDescriptor{Type: typ, Fields: bindings}
	ctx.frozenConfig.extensions.UpdateStructDescriptor(structDescriptor)
	ctx.encoderExtension.UpdateStructDescriptor(structDescriptor)
	ctx.decoderExtension.UpdateStructDescriptor(structDescriptor)
	ctx.extensions.UpdateStructDescriptor(structDescriptor)
	processTags(structDescriptor, ctx.frozenConfig)
	// merge normal & embedded bindings & sort with original order
	allBindings := sortableBindings(append(embeddedBindings, structDescriptor.Fields...))
	sort.Sort(allBindings)
	structDescriptor.Fields = allBindings
	return structDescriptor
}

type sortableBindings []*Binding

func (b sortableBindings) Len() int      { return len(b) }
func (b sortableBindings) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

func (b sortableBindings) Less(i, j int) bool {
	left := b[i].levels
	right := b[j].levels
	k := 0
	for {
		if left[k] < right[k] {
			return true
		} else if left[k] > right[k] {
			return false
		}
		k++
	}
}

func processTags(structDescriptor *StructDescriptor, cfg *frozenConfig) {
	for _, b := range structDescriptor.Fields {
		shouldOmitEmpty := cfg.omitEmptyStructField
		nilAsEmpty := cfg.nilAsEmpty
		tagParts := strings.Split(b.Field.Tag().Get(cfg.getTagKey()), ",")
		for _, tagPart := range tagParts[1:] {
			switch tagPart {
			case "nilasempty":
				nilAsEmpty = true
			case "omitempty":
				shouldOmitEmpty = true
			case "string":
				switch k := b.Field.Type().Kind(); k {
				case reflect.String:
					b.Decoder = &stringModeStringDecoder{b.Decoder, cfg}
					b.Encoder = &stringModeStringEncoder{b.Encoder, cfg}
				default:
					if (k == reflect.Int64 || k == reflect.Uint64) && cfg.int64AsString {
						// ignore
					} else {
						b.Decoder = &stringModeNumberDecoder{b.Decoder}
						b.Encoder = &stringModeNumberEncoder{b.Encoder}
					}
				}
			}
		}
		b.Decoder = &structFieldDecoder{field: b.Field, fieldDecoder: b.Decoder}
		b.Encoder = &structFieldEncoder{field: b.Field, fieldEncoder: b.Encoder, omitempty: shouldOmitEmpty, nilAsEmpty: nilAsEmpty}
	}
}

func calcFieldNames(originalFieldName string, tagProvidedFieldName string, wholeTag string) []string {
	// ignore?
	if wholeTag == "-" {
		return []string{}
	}
	// rename?
	var fieldNames []string
	if tagProvidedFieldName == "" {
		fieldNames = []string{originalFieldName}
	} else {
		fieldNames = []string{tagProvidedFieldName}
	}
	// private?
	isNotExported := unicode.IsLower(rune(originalFieldName[0])) || originalFieldName[0] == '_'
	if isNotExported {
		fieldNames = []string{}
	}
	return fieldNames
}
