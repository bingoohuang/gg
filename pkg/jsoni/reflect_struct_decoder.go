package jsoni

import (
	"context"
	"fmt"
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

	if l := len(fields); l == 0 {
		return &skipObjectDecoder{}
	} else if l <= 10 {
		return createDecoder(ctx, typ, fields)
	} else {
		return &generalStructDecoder{typ: typ, fields: fields}
	}
}

func createDecoder(ctx *ctx, typ reflect2.Type, fields map[string]*structFieldDecoder) ValDecoder {
	knownHash := map[int64]struct{}{0: {}}
	hashDecoders := make(map[int64]*structFieldDecoder)
	for fieldName, fieldDecoder := range fields {
		fieldHash := calcHash(fieldName, ctx.caseSensitive())
		if _, known := knownHash[fieldHash]; known {
			return &generalStructDecoder{typ: typ, fields: fields}
		}
		knownHash[fieldHash] = struct{}{}
		hashDecoders[fieldHash] = fieldDecoder
	}
	return &fieldsStructDecoder{typ: typ, HashDecoders: hashDecoders}
}

type generalStructDecoder struct {
	typ                   reflect2.Type
	fields                map[string]*structFieldDecoder
	disallowUnknownFields bool
}

func (d *generalStructDecoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() || !iter.incrementDepth() {
		return
	}
	var c byte
	for c = ','; c == ','; c = iter.nextToken() {
		d.decodeOneField(ctx, ptr, iter)
	}
	if iter.Error != nil && iter.Error != io.EOF && len(d.typ.Type1().Name()) != 0 {
		iter.Error = fmt.Errorf("%v.%s", d.typ, iter.Error.Error())
	}
	if c != '}' {
		iter.ReportError("struct Decode", `expect }, but found `+string([]byte{c}))
	}
	iter.decrementDepth()
}

func (d *generalStructDecoder) decodeOneField(ctx context.Context, ptr unsafe.Pointer, iter *Iterator) {
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
	fieldDecoder.Decode(ctx, ptr, iter)
}

type skipObjectDecoder struct{}

func (d *skipObjectDecoder) Decode(_ context.Context, _ unsafe.Pointer, iter *Iterator) {
	valueType := iter.WhatIsNext()
	if valueType != ObjectValue && valueType != NilValue {
		iter.ReportError("skipObjectDecoder", "expect object or null")
		return
	}
	iter.Skip()
}

type fieldsStructDecoder struct {
	typ          reflect2.Type
	HashDecoders map[int64]*structFieldDecoder
}

func (d *fieldsStructDecoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() || !iter.incrementDepth() {
		return
	}
	for {
		fieldHash := iter.readFieldHash()
		if decoder, ok := d.HashDecoders[fieldHash]; ok {
			decoder.Decode(ctx, ptr, iter)
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

type structFieldDecoder struct {
	field        reflect2.StructField
	fieldDecoder ValDecoder
}

func (d *structFieldDecoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *Iterator) {
	fieldPtr := d.field.UnsafeGet(ptr)
	d.fieldDecoder.Decode(ctx, fieldPtr, iter)
	if iter.Error != nil && iter.Error != io.EOF {
		iter.Error = fmt.Errorf("%s: %s", d.field.Name(), iter.Error.Error())
	}
}

type stringModeStringDecoder struct {
	elemDecoder ValDecoder
	cfg         *frozenConfig
}

func (d *stringModeStringDecoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *Iterator) {
	d.elemDecoder.Decode(ctx, ptr, iter)
	str := *((*string)(ptr))
	tempIter := d.cfg.BorrowIterator([]byte(str))
	defer d.cfg.ReturnIterator(tempIter)
	*((*string)(ptr)) = tempIter.ReadString()
}

type stringModeNumberDecoder struct {
	decoder ValDecoder
}

func (d *stringModeNumberDecoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *Iterator) {
	if iter.WhatIsNext() == NilValue {
		d.decoder.Decode(ctx, ptr, iter)
		return
	}

	if c := iter.nextToken(); c != '"' {
		iter.ReportError("stringModeNumberDecoder", `expect ", but found `+string([]byte{c}))
		return
	}
	d.decoder.Decode(ctx, ptr, iter)
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

func (d *stringModeNumberCompatibleDecoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *Iterator) {
	if iter.WhatIsNext() == NilValue {
		d.decoder.Decode(ctx, ptr, iter)
		return
	}

	isString := iter.nextToken() == '"'
	if !isString {
		iter.unreadByte()
	}
	d.decoder.Decode(ctx, ptr, iter)
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
