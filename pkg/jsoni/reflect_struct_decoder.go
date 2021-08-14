package jsoni

import (
	"fmt"
	"github.com/bingoohuang/gg/pkg/reflector"
	"io"
	"strings"
	"unsafe"

	"github.com/modern-go/reflect2"
)

func decoderOfStruct(ctx *ctx, typ reflect2.Type) ValDecoder {
	bindings := map[string]*Binding{}
	structDescriptor := describeStruct(ctx, typ)
	for _, binding := range structDescriptor.Fields {
		for _, fromName := range binding.FromNames {
			old := bindings[fromName]
			if old == nil {
				bindings[fromName] = binding
				continue
			}
			ignoreOld, ignoreNew := resolveConflictBinding(ctx.frozenConfig, old, binding)
			if ignoreOld {
				delete(bindings, fromName)
			}
			if !ignoreNew {
				bindings[fromName] = binding
			}
		}
	}
	fields := map[string]*structFieldDecoder{}
	for k, binding := range bindings {
		fields[k] = binding.Decoder.(*structFieldDecoder)
	}

	if !ctx.caseSensitive() {
		for k, b := range bindings {
			if _, found := fields[strings.ToLower(k)]; !found {
				fields[strings.ToLower(k)] = b.Decoder.(*structFieldDecoder)
			}
		}
	}

	return createStructDecoder(ctx, typ, fields)
}

func createStructDecoder(ctx *ctx, typ reflect2.Type, fields map[string]*structFieldDecoder) ValDecoder {
	if ctx.disallowUnknownFields {
		return &generalStructDecoder{typ: typ, fields: fields, disallowUnknownFields: true}
	}
	switch len(fields) {
	case 0:
		return &skipObjectDecoder{typ: typ}
	case 1:
		return createDecoder(1, ctx, typ, fields, &oneFieldStructDecoder{})
	case 2:
		return createDecoder(2, ctx, typ, fields, &twoFieldsStructDecoder{})
	case 3:
		return createDecoder(3, ctx, typ, fields, &threeFieldsStructDecoder{})
	case 4:
		return createDecoder(4, ctx, typ, fields, &fourFieldsStructDecoder{})
	case 5:
		return createDecoder(5, ctx, typ, fields, &fiveFieldsStructDecoder{})
	case 6:
		return createDecoder(6, ctx, typ, fields, &sixFieldsStructDecoder{})
	case 7:
		return createDecoder(7, ctx, typ, fields, &sevenFieldsStructDecoder{})
	case 8:
		return createDecoder(8, ctx, typ, fields, &eightFieldsStructDecoder{})
	case 9:
		return createDecoder(9, ctx, typ, fields, &nineFieldsStructDecoder{})
	case 10:
		return createDecoder(10, ctx, typ, fields, &tenFieldsStructDecoder{})
	}
	return &generalStructDecoder{typ: typ, fields: fields}
}

func createDecoder(n int, ctx *ctx, typ reflect2.Type, fields map[string]*structFieldDecoder, switcher switcher) ValDecoder {
	fieldNames := make([]int64, n)
	obj := reflector.New(switcher)
	knownHash := map[int64]struct{}{0: {}}
	i := 0

	for fieldName, fieldDecoder := range fields {
		fieldHash := calcHash(fieldName, ctx.caseSensitive())
		if _, known := knownHash[fieldHash]; known {
			return &generalStructDecoder{typ: typ, fields: fields}
		}
		knownHash[fieldHash] = struct{}{}
		fieldNames[i] = fieldHash
		obj.Field(fmt.Sprintf("FieldHash%d", i+1)).Set(fieldHash)
		obj.Field(fmt.Sprintf("FieldDecoder%d", i+1)).Set(fieldDecoder)
		i++
	}
	return &fieldsStructDecoder{typ: typ, switcher: switcher}
}

type generalStructDecoder struct {
	typ                   reflect2.Type
	fields                map[string]*structFieldDecoder
	disallowUnknownFields bool
}

func (d *generalStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	if !iter.incrementDepth() {
		return
	}
	var c byte
	for c = ','; c == ','; c = iter.nextToken() {
		d.decodeOneField(ptr, iter)
	}
	if iter.Error != nil && iter.Error != io.EOF && len(d.typ.Type1().Name()) != 0 {
		iter.Error = fmt.Errorf("%v.%s", d.typ, iter.Error.Error())
	}
	if c != '}' {
		iter.ReportError("struct Decode", `expect }, but found `+string([]byte{c}))
	}
	iter.decrementDepth()
}

func (d *generalStructDecoder) decodeOneField(ptr unsafe.Pointer, iter *Iterator) {
	var field string
	var fieldDecoder *structFieldDecoder
	if iter.cfg.objectFieldMustBeSimpleString {
		fieldBytes := iter.ReadStringAsSlice()
		field = *(*string)(unsafe.Pointer(&fieldBytes))
		fieldDecoder = d.fields[field]
		if fieldDecoder == nil && !iter.cfg.caseSensitive {
			fieldDecoder = d.fields[strings.ToLower(field)]
		}
	} else {
		field = iter.ReadString()
		fieldDecoder = d.fields[field]
		if fieldDecoder == nil && !iter.cfg.caseSensitive {
			fieldDecoder = d.fields[strings.ToLower(field)]
		}
	}
	if fieldDecoder == nil {
		if d.disallowUnknownFields {
			msg := "found unknown field: " + field
			iter.ReportError("ReadObject", msg)
		}
		if c := iter.nextToken(); c != ':' {
			iter.ReportError("ReadObject", "expect : after object field, but found "+string([]byte{c}))
		}
		iter.Skip()
		return
	}
	if c := iter.nextToken(); c != ':' {
		iter.ReportError("ReadObject", "expect : after object field, but found "+string([]byte{c}))
	}
	fieldDecoder.Decode(ptr, iter)
}

type skipObjectDecoder struct {
	typ reflect2.Type
}

func (d *skipObjectDecoder) Decode(_ unsafe.Pointer, iter *Iterator) {
	valueType := iter.WhatIsNext()
	if valueType != ObjectValue && valueType != NilValue {
		iter.ReportError("skipObjectDecoder", "expect object or null")
		return
	}
	iter.Skip()
}

type oneFieldStructDecoder struct {
	FieldHash1    int64
	FieldDecoder1 *structFieldDecoder
}

func (d *oneFieldStructDecoder) Switch(fieldHash int64) DecoderFn {
	switch fieldHash {
	case d.FieldHash1:
		return d.FieldDecoder1.Decode
	default:
		return nil
	}
}

type twoFieldsStructDecoder struct {
	FieldHash1, FieldHash2       int64
	FieldDecoder1, FieldDecoder2 *structFieldDecoder
}

func (d *twoFieldsStructDecoder) Switch(fieldHash int64) DecoderFn {
	switch fieldHash {
	case d.FieldHash1:
		return d.FieldDecoder1.Decode
	case d.FieldHash2:
		return d.FieldDecoder2.Decode
	default:
		return nil
	}
}

type switcher interface {
	Switch(fieldHash int64) DecoderFn
}

type fieldsStructDecoder struct {
	typ reflect2.Type
	switcher
}

func (d *fieldsStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	if !iter.incrementDepth() {
		return
	}
	for {
		fieldHash := iter.readFieldHash()
		if decoderFn := d.switcher.Switch(fieldHash); decoderFn != nil {
			decoderFn(ptr, iter)
		} else {
			iter.Skip()
		}

		if iter.isObjectEnd() {
			break
		}
	}
	if iter.Error != nil && iter.Error != io.EOF && len(d.typ.Type1().Name()) != 0 {
		iter.Error = fmt.Errorf("%v.%s", d.typ, iter.Error.Error())
	}
	iter.decrementDepth()
}

type DecoderFn func(ptr unsafe.Pointer, iter *Iterator)

type threeFieldsStructDecoder struct {
	FieldHash1, FieldHash2, FieldHash3          int64
	FieldDecoder1, FieldDecoder2, FieldDecoder3 *structFieldDecoder
}

func (d *threeFieldsStructDecoder) Switch(fieldHash int64) DecoderFn {
	switch fieldHash {
	case d.FieldHash1:
		return d.FieldDecoder1.Decode
	case d.FieldHash2:
		return d.FieldDecoder2.Decode
	case d.FieldHash3:
		return d.FieldDecoder3.Decode
	default:
		return nil
	}
}

type fourFieldsStructDecoder struct {
	FieldHash1, FieldHash2, FieldHash3, FieldHash4             int64
	FieldDecoder1, FieldDecoder2, FieldDecoder3, FieldDecoder4 *structFieldDecoder
}

func (d *fourFieldsStructDecoder) Switch(fieldHash int64) DecoderFn {
	switch fieldHash {
	case d.FieldHash1:
		return d.FieldDecoder1.Decode
	case d.FieldHash2:
		return d.FieldDecoder2.Decode
	case d.FieldHash3:
		return d.FieldDecoder3.Decode
	case d.FieldHash4:
		return d.FieldDecoder4.Decode
	default:
		return nil
	}
}

type fiveFieldsStructDecoder struct {
	FieldHash1, FieldHash2, FieldHash3, FieldHash4, FieldHash5                int64
	FieldDecoder1, FieldDecoder2, FieldDecoder3, FieldDecoder4, FieldDecoder5 *structFieldDecoder
}

func (d *fiveFieldsStructDecoder) Switch(fieldHash int64) DecoderFn {
	switch fieldHash {
	case d.FieldHash1:
		return d.FieldDecoder1.Decode
	case d.FieldHash2:
		return d.FieldDecoder2.Decode
	case d.FieldHash3:
		return d.FieldDecoder3.Decode
	case d.FieldHash4:
		return d.FieldDecoder4.Decode
	case d.FieldHash5:
		return d.FieldDecoder5.Decode
	default:
		return nil
	}
}

type sixFieldsStructDecoder struct {
	FieldHash1, FieldHash2, FieldHash3, FieldHash4, FieldHash5, FieldHash6                   int64
	FieldDecoder1, FieldDecoder2, FieldDecoder3, FieldDecoder4, FieldDecoder5, FieldDecoder6 *structFieldDecoder
}

func (d *sixFieldsStructDecoder) Switch(fieldHash int64) DecoderFn {
	switch fieldHash {
	case d.FieldHash1:
		return d.FieldDecoder1.Decode
	case d.FieldHash2:
		return d.FieldDecoder2.Decode
	case d.FieldHash3:
		return d.FieldDecoder3.Decode
	case d.FieldHash4:
		return d.FieldDecoder4.Decode
	case d.FieldHash5:
		return d.FieldDecoder5.Decode
	case d.FieldHash6:
		return d.FieldDecoder6.Decode
	default:
		return nil
	}
}

type sevenFieldsStructDecoder struct {
	FieldHash1, FieldHash2, FieldHash3, FieldHash4, FieldHash5, FieldHash6, FieldHash7                      int64
	FieldDecoder1, FieldDecoder2, FieldDecoder3, FieldDecoder4, FieldDecoder5, FieldDecoder6, FieldDecoder7 *structFieldDecoder
}

func (d *sevenFieldsStructDecoder) Switch(fieldHash int64) DecoderFn {
	switch fieldHash {
	case d.FieldHash1:
		return d.FieldDecoder1.Decode
	case d.FieldHash2:
		return d.FieldDecoder2.Decode
	case d.FieldHash3:
		return d.FieldDecoder3.Decode
	case d.FieldHash4:
		return d.FieldDecoder4.Decode
	case d.FieldHash5:
		return d.FieldDecoder5.Decode
	case d.FieldHash6:
		return d.FieldDecoder6.Decode
	case d.FieldHash7:
		return d.FieldDecoder7.Decode
	default:
		return nil
	}
}

type eightFieldsStructDecoder struct {
	FieldHash1, FieldHash2, FieldHash3, FieldHash4, FieldHash5, FieldHash6, FieldHash7, FieldHash8                         int64
	FieldDecoder1, FieldDecoder2, FieldDecoder3, FieldDecoder4, FieldDecoder5, FieldDecoder6, FieldDecoder7, FieldDecoder8 *structFieldDecoder
}

func (d *eightFieldsStructDecoder) Switch(fieldHash int64) DecoderFn {
	switch fieldHash {
	case d.FieldHash1:
		return d.FieldDecoder1.Decode
	case d.FieldHash2:
		return d.FieldDecoder2.Decode
	case d.FieldHash3:
		return d.FieldDecoder3.Decode
	case d.FieldHash4:
		return d.FieldDecoder4.Decode
	case d.FieldHash5:
		return d.FieldDecoder5.Decode
	case d.FieldHash6:
		return d.FieldDecoder6.Decode
	case d.FieldHash7:
		return d.FieldDecoder7.Decode
	case d.FieldHash8:
		return d.FieldDecoder8.Decode
	default:
		return nil
	}
}

type nineFieldsStructDecoder struct {
	FieldHash1, FieldHash2, FieldHash3, FieldHash4, FieldHash5, FieldHash6, FieldHash7, FieldHash8, FieldHash9                            int64
	FieldDecoder1, FieldDecoder2, FieldDecoder3, FieldDecoder4, FieldDecoder5, FieldDecoder6, FieldDecoder7, FieldDecoder8, FieldDecoder9 *structFieldDecoder
}

func (d *nineFieldsStructDecoder) Switch(fieldHash int64) DecoderFn {
	switch fieldHash {
	case d.FieldHash1:
		return d.FieldDecoder1.Decode
	case d.FieldHash2:
		return d.FieldDecoder2.Decode
	case d.FieldHash3:
		return d.FieldDecoder3.Decode
	case d.FieldHash4:
		return d.FieldDecoder4.Decode
	case d.FieldHash5:
		return d.FieldDecoder5.Decode
	case d.FieldHash6:
		return d.FieldDecoder6.Decode
	case d.FieldHash7:
		return d.FieldDecoder7.Decode
	case d.FieldHash8:
		return d.FieldDecoder8.Decode
	case d.FieldHash9:
		return d.FieldDecoder9.Decode
	default:
		return nil
	}
}

type tenFieldsStructDecoder struct {
	FieldHash1, FieldHash2, FieldHash3, FieldHash4, FieldHash5, FieldHash6, FieldHash7, FieldHash8, FieldHash9, FieldHash10                               int64
	FieldDecoder1, FieldDecoder2, FieldDecoder3, FieldDecoder4, FieldDecoder5, FieldDecoder6, FieldDecoder7, FieldDecoder8, FieldDecoder9, FieldDecoder10 *structFieldDecoder
}

func (d *tenFieldsStructDecoder) Switch(fieldHash int64) DecoderFn {
	switch fieldHash {
	case d.FieldHash1:
		return d.FieldDecoder1.Decode
	case d.FieldHash2:
		return d.FieldDecoder2.Decode
	case d.FieldHash3:
		return d.FieldDecoder3.Decode
	case d.FieldHash4:
		return d.FieldDecoder4.Decode
	case d.FieldHash5:
		return d.FieldDecoder5.Decode
	case d.FieldHash6:
		return d.FieldDecoder6.Decode
	case d.FieldHash7:
		return d.FieldDecoder7.Decode
	case d.FieldHash8:
		return d.FieldDecoder8.Decode
	case d.FieldHash9:
		return d.FieldDecoder9.Decode
	case d.FieldHash10:
		return d.FieldDecoder10.Decode
	default:
		return nil
	}
}

type structFieldDecoder struct {
	field        reflect2.StructField
	fieldDecoder ValDecoder
}

func (d *structFieldDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	fieldPtr := d.field.UnsafeGet(ptr)
	d.fieldDecoder.Decode(fieldPtr, iter)
	if iter.Error != nil && iter.Error != io.EOF {
		iter.Error = fmt.Errorf("%s: %s", d.field.Name(), iter.Error.Error())
	}
}

type stringModeStringDecoder struct {
	elemDecoder ValDecoder
	cfg         *frozenConfig
}

func (d *stringModeStringDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	d.elemDecoder.Decode(ptr, iter)
	str := *((*string)(ptr))
	tempIter := d.cfg.BorrowIterator([]byte(str))
	defer d.cfg.ReturnIterator(tempIter)
	*((*string)(ptr)) = tempIter.ReadString()
}

type stringModeNumberDecoder struct {
	decoder ValDecoder
}

func (d *stringModeNumberDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if iter.WhatIsNext() == NilValue {
		d.decoder.Decode(ptr, iter)
		return
	}

	if c := iter.nextToken(); c != '"' {
		iter.ReportError("stringModeNumberDecoder", `expect ", but found `+string([]byte{c}))
		return
	}
	d.decoder.Decode(ptr, iter)
	if iter.Error != nil {
		return
	}
	if c := iter.readByte(); c != '"' {
		iter.ReportError("stringModeNumberDecoder", `expect ", but found `+string([]byte{c}))
		return
	}
}

type stringModeNumberCompatibleDecoder struct {
	decoder ValDecoder
}

func (d *stringModeNumberCompatibleDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if iter.WhatIsNext() == NilValue {
		d.decoder.Decode(ptr, iter)
		return
	}

	isString := iter.nextToken() == '"'
	if !isString {
		iter.unreadByte()
	}
	d.decoder.Decode(ptr, iter)
	if iter.Error != nil {
		return
	}

	if isString {
		if c := iter.readByte(); c != '"' {
			iter.ReportError("stringModeNumberCompatibleDecoder", `expect ", but found `+string([]byte{c}))
			return
		}
	}
}
