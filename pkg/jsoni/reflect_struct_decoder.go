package jsoni

import (
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
	knownHash := map[int64]struct{}{0: {}}

	switch len(fields) {
	case 0:
		return &skipObjectDecoder{typ: typ}
	case 1:
		for fieldName, fieldDecoder := range fields {
			fieldHash := calcHash(fieldName, ctx.caseSensitive())
			if _, known := knownHash[fieldHash]; known {
				return &generalStructDecoder{typ: typ, fields: fields}
			}
			knownHash[fieldHash] = struct{}{}
			return &oneFieldStructDecoder{typ: typ, fieldHash: fieldHash, fieldDecoder: fieldDecoder}
		}
	case 2:
		var fieldHash1, fieldHash2 int64
		var fieldDecoder1, fieldDecoder2 *structFieldDecoder
		for fieldName, fieldDecoder := range fields {
			fieldHash := calcHash(fieldName, ctx.caseSensitive())
			if _, known := knownHash[fieldHash]; known {
				return &generalStructDecoder{typ: typ, fields: fields}
			}
			knownHash[fieldHash] = struct{}{}
			if fieldHash1 == 0 {
				fieldHash1 = fieldHash
				fieldDecoder1 = fieldDecoder
			} else {
				fieldHash2 = fieldHash
				fieldDecoder2 = fieldDecoder
			}
		}
		return &twoFieldsStructDecoder{typ: typ, fieldHash1: fieldHash1, fieldDecoder1: fieldDecoder1, fieldHash2: fieldHash2, fieldDecoder2: fieldDecoder2}
	case 3:
		var fieldName1, fieldName2, fieldName3 int64
		var fieldDecoder1, fieldDecoder2, fieldDecoder3 *structFieldDecoder
		for fieldName, fieldDecoder := range fields {
			fieldHash := calcHash(fieldName, ctx.caseSensitive())
			if _, known := knownHash[fieldHash]; known {
				return &generalStructDecoder{typ, fields, false}
			}
			knownHash[fieldHash] = struct{}{}
			if fieldName1 == 0 {
				fieldName1 = fieldHash
				fieldDecoder1 = fieldDecoder
			} else if fieldName2 == 0 {
				fieldName2 = fieldHash
				fieldDecoder2 = fieldDecoder
			} else {
				fieldName3 = fieldHash
				fieldDecoder3 = fieldDecoder
			}
		}
		return &threeFieldsStructDecoder{typ: typ,
			fieldHash1: fieldName1, fieldDecoder1: fieldDecoder1,
			fieldHash2: fieldName2, fieldDecoder2: fieldDecoder2,
			fieldHash3: fieldName3, fieldDecoder3: fieldDecoder3}
	case 4:
		var fieldName1, fieldName2, fieldName3, fieldName4 int64
		var fieldDecoder1, fieldDecoder2, fieldDecoder3, fieldDecoder4 *structFieldDecoder
		for fieldName, fieldDecoder := range fields {
			fieldHash := calcHash(fieldName, ctx.caseSensitive())
			if _, known := knownHash[fieldHash]; known {
				return &generalStructDecoder{typ: typ, fields: fields}
			}
			knownHash[fieldHash] = struct{}{}
			if fieldName1 == 0 {
				fieldName1 = fieldHash
				fieldDecoder1 = fieldDecoder
			} else if fieldName2 == 0 {
				fieldName2 = fieldHash
				fieldDecoder2 = fieldDecoder
			} else if fieldName3 == 0 {
				fieldName3 = fieldHash
				fieldDecoder3 = fieldDecoder
			} else {
				fieldName4 = fieldHash
				fieldDecoder4 = fieldDecoder
			}
		}
		return &fourFieldsStructDecoder{typ: typ,
			fieldHash1: fieldName1, fieldDecoder1: fieldDecoder1,
			fieldHash2: fieldName2, fieldDecoder2: fieldDecoder2,
			fieldHash3: fieldName3, fieldDecoder3: fieldDecoder3,
			fieldHash4: fieldName4, fieldDecoder4: fieldDecoder4}
	case 5:
		var fieldName1, fieldName2, fieldName3, fieldName4, fieldName5 int64
		var fieldDecoder1, fieldDecoder2, fieldDecoder3, fieldDecoder4, fieldDecoder5 *structFieldDecoder
		for fieldName, fieldDecoder := range fields {
			fieldHash := calcHash(fieldName, ctx.caseSensitive())
			if _, known := knownHash[fieldHash]; known {
				return &generalStructDecoder{typ, fields, false}
			}
			knownHash[fieldHash] = struct{}{}
			if fieldName1 == 0 {
				fieldName1 = fieldHash
				fieldDecoder1 = fieldDecoder
			} else if fieldName2 == 0 {
				fieldName2 = fieldHash
				fieldDecoder2 = fieldDecoder
			} else if fieldName3 == 0 {
				fieldName3 = fieldHash
				fieldDecoder3 = fieldDecoder
			} else if fieldName4 == 0 {
				fieldName4 = fieldHash
				fieldDecoder4 = fieldDecoder
			} else {
				fieldName5 = fieldHash
				fieldDecoder5 = fieldDecoder
			}
		}
		return &fiveFieldsStructDecoder{typ: typ,
			fieldHash1: fieldName1, fieldDecoder1: fieldDecoder1,
			fieldHash2: fieldName2, fieldDecoder2: fieldDecoder2,
			fieldHash3: fieldName3, fieldDecoder3: fieldDecoder3,
			fieldHash4: fieldName4, fieldDecoder4: fieldDecoder4,
			fieldHash5: fieldName5, fieldDecoder5: fieldDecoder5}
	case 6:
		var fieldName1, fieldName2, fieldName3, fieldName4, fieldName5, fieldName6 int64
		var fieldDecoder1, fieldDecoder2, fieldDecoder3, fieldDecoder4, fieldDecoder5, fieldDecoder6 *structFieldDecoder
		for fieldName, fieldDecoder := range fields {
			fieldHash := calcHash(fieldName, ctx.caseSensitive())
			if _, known := knownHash[fieldHash]; known {
				return &generalStructDecoder{typ, fields, false}
			}
			knownHash[fieldHash] = struct{}{}
			if fieldName1 == 0 {
				fieldName1 = fieldHash
				fieldDecoder1 = fieldDecoder
			} else if fieldName2 == 0 {
				fieldName2 = fieldHash
				fieldDecoder2 = fieldDecoder
			} else if fieldName3 == 0 {
				fieldName3 = fieldHash
				fieldDecoder3 = fieldDecoder
			} else if fieldName4 == 0 {
				fieldName4 = fieldHash
				fieldDecoder4 = fieldDecoder
			} else if fieldName5 == 0 {
				fieldName5 = fieldHash
				fieldDecoder5 = fieldDecoder
			} else {
				fieldName6 = fieldHash
				fieldDecoder6 = fieldDecoder
			}
		}
		return &sixFieldsStructDecoder{typ: typ,
			fieldHash1: fieldName1, fieldDecoder1: fieldDecoder1,
			fieldHash2: fieldName2, fieldDecoder2: fieldDecoder2,
			fieldHash3: fieldName3, fieldDecoder3: fieldDecoder3,
			fieldHash4: fieldName4, fieldDecoder4: fieldDecoder4,
			fieldHash5: fieldName5, fieldDecoder5: fieldDecoder5,
			fieldHash6: fieldName6, fieldDecoder6: fieldDecoder6}
	case 7:
		var fieldName1, fieldName2, fieldName3, fieldName4, fieldName5, fieldName6, fieldName7 int64
		var fieldDecoder1, fieldDecoder2, fieldDecoder3, fieldDecoder4, fieldDecoder5, fieldDecoder6, fieldDecoder7 *structFieldDecoder
		for fieldName, fieldDecoder := range fields {
			fieldHash := calcHash(fieldName, ctx.caseSensitive())
			if _, known := knownHash[fieldHash]; known {
				return &generalStructDecoder{typ: typ, fields: fields}
			}
			knownHash[fieldHash] = struct{}{}
			if fieldName1 == 0 {
				fieldName1 = fieldHash
				fieldDecoder1 = fieldDecoder
			} else if fieldName2 == 0 {
				fieldName2 = fieldHash
				fieldDecoder2 = fieldDecoder
			} else if fieldName3 == 0 {
				fieldName3 = fieldHash
				fieldDecoder3 = fieldDecoder
			} else if fieldName4 == 0 {
				fieldName4 = fieldHash
				fieldDecoder4 = fieldDecoder
			} else if fieldName5 == 0 {
				fieldName5 = fieldHash
				fieldDecoder5 = fieldDecoder
			} else if fieldName6 == 0 {
				fieldName6 = fieldHash
				fieldDecoder6 = fieldDecoder
			} else {
				fieldName7 = fieldHash
				fieldDecoder7 = fieldDecoder
			}
		}
		return &sevenFieldsStructDecoder{typ: typ,
			fieldHash1: fieldName1, fieldDecoder1: fieldDecoder1,
			fieldHash2: fieldName2, fieldDecoder2: fieldDecoder2,
			fieldHash3: fieldName3, fieldDecoder3: fieldDecoder3,
			fieldHash4: fieldName4, fieldDecoder4: fieldDecoder4,
			fieldHash5: fieldName5, fieldDecoder5: fieldDecoder5,
			fieldHash6: fieldName6, fieldDecoder6: fieldDecoder6,
			fieldHash7: fieldName7, fieldDecoder7: fieldDecoder7}
	case 8:
		var fieldName1, fieldName2, fieldName3, fieldName4, fieldName5, fieldName6, fieldName7, fieldName8 int64
		var fieldDecoder1, fieldDecoder2, fieldDecoder3, fieldDecoder4, fieldDecoder5, fieldDecoder6, fieldDecoder7, fieldDecoder8 *structFieldDecoder
		for fieldName, fieldDecoder := range fields {
			fieldHash := calcHash(fieldName, ctx.caseSensitive())
			if _, known := knownHash[fieldHash]; known {
				return &generalStructDecoder{typ, fields, false}
			}
			knownHash[fieldHash] = struct{}{}
			if fieldName1 == 0 {
				fieldName1 = fieldHash
				fieldDecoder1 = fieldDecoder
			} else if fieldName2 == 0 {
				fieldName2 = fieldHash
				fieldDecoder2 = fieldDecoder
			} else if fieldName3 == 0 {
				fieldName3 = fieldHash
				fieldDecoder3 = fieldDecoder
			} else if fieldName4 == 0 {
				fieldName4 = fieldHash
				fieldDecoder4 = fieldDecoder
			} else if fieldName5 == 0 {
				fieldName5 = fieldHash
				fieldDecoder5 = fieldDecoder
			} else if fieldName6 == 0 {
				fieldName6 = fieldHash
				fieldDecoder6 = fieldDecoder
			} else if fieldName7 == 0 {
				fieldName7 = fieldHash
				fieldDecoder7 = fieldDecoder
			} else {
				fieldName8 = fieldHash
				fieldDecoder8 = fieldDecoder
			}
		}
		return &eightFieldsStructDecoder{typ: typ,
			fieldHash1: fieldName1, fieldDecoder1: fieldDecoder1,
			fieldHash2: fieldName2, fieldDecoder2: fieldDecoder2,
			fieldHash3: fieldName3, fieldDecoder3: fieldDecoder3,
			fieldHash4: fieldName4, fieldDecoder4: fieldDecoder4,
			fieldHash5: fieldName5, fieldDecoder5: fieldDecoder5,
			fieldHash6: fieldName6, fieldDecoder6: fieldDecoder6,
			fieldHash7: fieldName7, fieldDecoder7: fieldDecoder7,
			fieldHash8: fieldName8, fieldDecoder8: fieldDecoder8}
	case 9:
		var fieldName1, fieldName2, fieldName3, fieldName4, fieldName5, fieldName6, fieldName7, fieldName8, fieldName9 int64
		var fieldDecoder1, fieldDecoder2, fieldDecoder3, fieldDecoder4, fieldDecoder5, fieldDecoder6, fieldDecoder7, fieldDecoder8, fieldDecoder9 *structFieldDecoder
		for fieldName, fieldDecoder := range fields {
			fieldHash := calcHash(fieldName, ctx.caseSensitive())
			if _, known := knownHash[fieldHash]; known {
				return &generalStructDecoder{typ: typ, fields: fields}
			}
			knownHash[fieldHash] = struct{}{}
			if fieldName1 == 0 {
				fieldName1 = fieldHash
				fieldDecoder1 = fieldDecoder
			} else if fieldName2 == 0 {
				fieldName2 = fieldHash
				fieldDecoder2 = fieldDecoder
			} else if fieldName3 == 0 {
				fieldName3 = fieldHash
				fieldDecoder3 = fieldDecoder
			} else if fieldName4 == 0 {
				fieldName4 = fieldHash
				fieldDecoder4 = fieldDecoder
			} else if fieldName5 == 0 {
				fieldName5 = fieldHash
				fieldDecoder5 = fieldDecoder
			} else if fieldName6 == 0 {
				fieldName6 = fieldHash
				fieldDecoder6 = fieldDecoder
			} else if fieldName7 == 0 {
				fieldName7 = fieldHash
				fieldDecoder7 = fieldDecoder
			} else if fieldName8 == 0 {
				fieldName8 = fieldHash
				fieldDecoder8 = fieldDecoder
			} else {
				fieldName9 = fieldHash
				fieldDecoder9 = fieldDecoder
			}
		}
		return &nineFieldsStructDecoder{typ: typ,
			fieldHash1: fieldName1, fieldDecoder1: fieldDecoder1,
			fieldHash2: fieldName2, fieldDecoder2: fieldDecoder2,
			fieldHash3: fieldName3, fieldDecoder3: fieldDecoder3,
			fieldHash4: fieldName4, fieldDecoder4: fieldDecoder4,
			fieldHash5: fieldName5, fieldDecoder5: fieldDecoder5,
			fieldHash6: fieldName6, fieldDecoder6: fieldDecoder6,
			fieldHash7: fieldName7, fieldDecoder7: fieldDecoder7,
			fieldHash8: fieldName8, fieldDecoder8: fieldDecoder8,
			fieldHash9: fieldName9, fieldDecoder9: fieldDecoder9}
	case 10:
		var fieldName1, fieldName2, fieldName3, fieldName4, fieldName5, fieldName6, fieldName7, fieldName8, fieldName9, fieldName10 int64
		var fieldDecoder1, fieldDecoder2, fieldDecoder3, fieldDecoder4, fieldDecoder5, fieldDecoder6, fieldDecoder7, fieldDecoder8, fieldDecoder9, fieldDecoder10 *structFieldDecoder
		for fieldName, fieldDecoder := range fields {
			fieldHash := calcHash(fieldName, ctx.caseSensitive())
			if _, known := knownHash[fieldHash]; known {
				return &generalStructDecoder{typ: typ, fields: fields}
			}
			knownHash[fieldHash] = struct{}{}
			if fieldName1 == 0 {
				fieldName1 = fieldHash
				fieldDecoder1 = fieldDecoder
			} else if fieldName2 == 0 {
				fieldName2 = fieldHash
				fieldDecoder2 = fieldDecoder
			} else if fieldName3 == 0 {
				fieldName3 = fieldHash
				fieldDecoder3 = fieldDecoder
			} else if fieldName4 == 0 {
				fieldName4 = fieldHash
				fieldDecoder4 = fieldDecoder
			} else if fieldName5 == 0 {
				fieldName5 = fieldHash
				fieldDecoder5 = fieldDecoder
			} else if fieldName6 == 0 {
				fieldName6 = fieldHash
				fieldDecoder6 = fieldDecoder
			} else if fieldName7 == 0 {
				fieldName7 = fieldHash
				fieldDecoder7 = fieldDecoder
			} else if fieldName8 == 0 {
				fieldName8 = fieldHash
				fieldDecoder8 = fieldDecoder
			} else if fieldName9 == 0 {
				fieldName9 = fieldHash
				fieldDecoder9 = fieldDecoder
			} else {
				fieldName10 = fieldHash
				fieldDecoder10 = fieldDecoder
			}
		}
		return &tenFieldsStructDecoder{typ: typ,
			fieldHash1: fieldName1, fieldDecoder1: fieldDecoder1,
			fieldHash2: fieldName2, fieldDecoder2: fieldDecoder2,
			fieldHash3: fieldName3, fieldDecoder3: fieldDecoder3,
			fieldHash4: fieldName4, fieldDecoder4: fieldDecoder4,
			fieldHash5: fieldName5, fieldDecoder5: fieldDecoder5,
			fieldHash6: fieldName6, fieldDecoder6: fieldDecoder6,
			fieldHash7: fieldName7, fieldDecoder7: fieldDecoder7,
			fieldHash8: fieldName8, fieldDecoder8: fieldDecoder8,
			fieldHash9: fieldName9, fieldDecoder9: fieldDecoder9,
			fieldHash10: fieldName10, fieldDecoder10: fieldDecoder10}
	}
	return &generalStructDecoder{typ: typ, fields: fields}
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
	typ          reflect2.Type
	fieldHash    int64
	fieldDecoder *structFieldDecoder
}

func (d *oneFieldStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	if !iter.incrementDepth() {
		return
	}
	for {
		if iter.readFieldHash() == d.fieldHash {
			d.fieldDecoder.Decode(ptr, iter)
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

type twoFieldsStructDecoder struct {
	typ           reflect2.Type
	fieldHash1    int64
	fieldDecoder1 *structFieldDecoder
	fieldHash2    int64
	fieldDecoder2 *structFieldDecoder
}

func (d *twoFieldsStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	if !iter.incrementDepth() {
		return
	}
	for {
		switch iter.readFieldHash() {
		case d.fieldHash1:
			d.fieldDecoder1.Decode(ptr, iter)
		case d.fieldHash2:
			d.fieldDecoder2.Decode(ptr, iter)
		default:
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

type threeFieldsStructDecoder struct {
	typ           reflect2.Type
	fieldHash1    int64
	fieldDecoder1 *structFieldDecoder
	fieldHash2    int64
	fieldDecoder2 *structFieldDecoder
	fieldHash3    int64
	fieldDecoder3 *structFieldDecoder
}

func (d *threeFieldsStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	if !iter.incrementDepth() {
		return
	}
	for {
		switch iter.readFieldHash() {
		case d.fieldHash1:
			d.fieldDecoder1.Decode(ptr, iter)
		case d.fieldHash2:
			d.fieldDecoder2.Decode(ptr, iter)
		case d.fieldHash3:
			d.fieldDecoder3.Decode(ptr, iter)
		default:
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

type fourFieldsStructDecoder struct {
	typ           reflect2.Type
	fieldHash1    int64
	fieldDecoder1 *structFieldDecoder
	fieldHash2    int64
	fieldDecoder2 *structFieldDecoder
	fieldHash3    int64
	fieldDecoder3 *structFieldDecoder
	fieldHash4    int64
	fieldDecoder4 *structFieldDecoder
}

func (d *fourFieldsStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	if !iter.incrementDepth() {
		return
	}
	for {
		switch iter.readFieldHash() {
		case d.fieldHash1:
			d.fieldDecoder1.Decode(ptr, iter)
		case d.fieldHash2:
			d.fieldDecoder2.Decode(ptr, iter)
		case d.fieldHash3:
			d.fieldDecoder3.Decode(ptr, iter)
		case d.fieldHash4:
			d.fieldDecoder4.Decode(ptr, iter)
		default:
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

type fiveFieldsStructDecoder struct {
	typ           reflect2.Type
	fieldHash1    int64
	fieldDecoder1 *structFieldDecoder
	fieldHash2    int64
	fieldDecoder2 *structFieldDecoder
	fieldHash3    int64
	fieldDecoder3 *structFieldDecoder
	fieldHash4    int64
	fieldDecoder4 *structFieldDecoder
	fieldHash5    int64
	fieldDecoder5 *structFieldDecoder
}

func (d *fiveFieldsStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	if !iter.incrementDepth() {
		return
	}
	for {
		switch iter.readFieldHash() {
		case d.fieldHash1:
			d.fieldDecoder1.Decode(ptr, iter)
		case d.fieldHash2:
			d.fieldDecoder2.Decode(ptr, iter)
		case d.fieldHash3:
			d.fieldDecoder3.Decode(ptr, iter)
		case d.fieldHash4:
			d.fieldDecoder4.Decode(ptr, iter)
		case d.fieldHash5:
			d.fieldDecoder5.Decode(ptr, iter)
		default:
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

type sixFieldsStructDecoder struct {
	typ           reflect2.Type
	fieldHash1    int64
	fieldDecoder1 *structFieldDecoder
	fieldHash2    int64
	fieldDecoder2 *structFieldDecoder
	fieldHash3    int64
	fieldDecoder3 *structFieldDecoder
	fieldHash4    int64
	fieldDecoder4 *structFieldDecoder
	fieldHash5    int64
	fieldDecoder5 *structFieldDecoder
	fieldHash6    int64
	fieldDecoder6 *structFieldDecoder
}

func (d *sixFieldsStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	if !iter.incrementDepth() {
		return
	}
	for {
		switch iter.readFieldHash() {
		case d.fieldHash1:
			d.fieldDecoder1.Decode(ptr, iter)
		case d.fieldHash2:
			d.fieldDecoder2.Decode(ptr, iter)
		case d.fieldHash3:
			d.fieldDecoder3.Decode(ptr, iter)
		case d.fieldHash4:
			d.fieldDecoder4.Decode(ptr, iter)
		case d.fieldHash5:
			d.fieldDecoder5.Decode(ptr, iter)
		case d.fieldHash6:
			d.fieldDecoder6.Decode(ptr, iter)
		default:
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

type sevenFieldsStructDecoder struct {
	typ           reflect2.Type
	fieldHash1    int64
	fieldDecoder1 *structFieldDecoder
	fieldHash2    int64
	fieldDecoder2 *structFieldDecoder
	fieldHash3    int64
	fieldDecoder3 *structFieldDecoder
	fieldHash4    int64
	fieldDecoder4 *structFieldDecoder
	fieldHash5    int64
	fieldDecoder5 *structFieldDecoder
	fieldHash6    int64
	fieldDecoder6 *structFieldDecoder
	fieldHash7    int64
	fieldDecoder7 *structFieldDecoder
}

func (d *sevenFieldsStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	if !iter.incrementDepth() {
		return
	}
	for {
		switch iter.readFieldHash() {
		case d.fieldHash1:
			d.fieldDecoder1.Decode(ptr, iter)
		case d.fieldHash2:
			d.fieldDecoder2.Decode(ptr, iter)
		case d.fieldHash3:
			d.fieldDecoder3.Decode(ptr, iter)
		case d.fieldHash4:
			d.fieldDecoder4.Decode(ptr, iter)
		case d.fieldHash5:
			d.fieldDecoder5.Decode(ptr, iter)
		case d.fieldHash6:
			d.fieldDecoder6.Decode(ptr, iter)
		case d.fieldHash7:
			d.fieldDecoder7.Decode(ptr, iter)
		default:
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

type eightFieldsStructDecoder struct {
	typ           reflect2.Type
	fieldHash1    int64
	fieldDecoder1 *structFieldDecoder
	fieldHash2    int64
	fieldDecoder2 *structFieldDecoder
	fieldHash3    int64
	fieldDecoder3 *structFieldDecoder
	fieldHash4    int64
	fieldDecoder4 *structFieldDecoder
	fieldHash5    int64
	fieldDecoder5 *structFieldDecoder
	fieldHash6    int64
	fieldDecoder6 *structFieldDecoder
	fieldHash7    int64
	fieldDecoder7 *structFieldDecoder
	fieldHash8    int64
	fieldDecoder8 *structFieldDecoder
}

func (d *eightFieldsStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	if !iter.incrementDepth() {
		return
	}
	for {
		switch iter.readFieldHash() {
		case d.fieldHash1:
			d.fieldDecoder1.Decode(ptr, iter)
		case d.fieldHash2:
			d.fieldDecoder2.Decode(ptr, iter)
		case d.fieldHash3:
			d.fieldDecoder3.Decode(ptr, iter)
		case d.fieldHash4:
			d.fieldDecoder4.Decode(ptr, iter)
		case d.fieldHash5:
			d.fieldDecoder5.Decode(ptr, iter)
		case d.fieldHash6:
			d.fieldDecoder6.Decode(ptr, iter)
		case d.fieldHash7:
			d.fieldDecoder7.Decode(ptr, iter)
		case d.fieldHash8:
			d.fieldDecoder8.Decode(ptr, iter)
		default:
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

type nineFieldsStructDecoder struct {
	typ           reflect2.Type
	fieldHash1    int64
	fieldDecoder1 *structFieldDecoder
	fieldHash2    int64
	fieldDecoder2 *structFieldDecoder
	fieldHash3    int64
	fieldDecoder3 *structFieldDecoder
	fieldHash4    int64
	fieldDecoder4 *structFieldDecoder
	fieldHash5    int64
	fieldDecoder5 *structFieldDecoder
	fieldHash6    int64
	fieldDecoder6 *structFieldDecoder
	fieldHash7    int64
	fieldDecoder7 *structFieldDecoder
	fieldHash8    int64
	fieldDecoder8 *structFieldDecoder
	fieldHash9    int64
	fieldDecoder9 *structFieldDecoder
}

func (d *nineFieldsStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	if !iter.incrementDepth() {
		return
	}
	for {
		switch iter.readFieldHash() {
		case d.fieldHash1:
			d.fieldDecoder1.Decode(ptr, iter)
		case d.fieldHash2:
			d.fieldDecoder2.Decode(ptr, iter)
		case d.fieldHash3:
			d.fieldDecoder3.Decode(ptr, iter)
		case d.fieldHash4:
			d.fieldDecoder4.Decode(ptr, iter)
		case d.fieldHash5:
			d.fieldDecoder5.Decode(ptr, iter)
		case d.fieldHash6:
			d.fieldDecoder6.Decode(ptr, iter)
		case d.fieldHash7:
			d.fieldDecoder7.Decode(ptr, iter)
		case d.fieldHash8:
			d.fieldDecoder8.Decode(ptr, iter)
		case d.fieldHash9:
			d.fieldDecoder9.Decode(ptr, iter)
		default:
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

type tenFieldsStructDecoder struct {
	typ            reflect2.Type
	fieldHash1     int64
	fieldDecoder1  *structFieldDecoder
	fieldHash2     int64
	fieldDecoder2  *structFieldDecoder
	fieldHash3     int64
	fieldDecoder3  *structFieldDecoder
	fieldHash4     int64
	fieldDecoder4  *structFieldDecoder
	fieldHash5     int64
	fieldDecoder5  *structFieldDecoder
	fieldHash6     int64
	fieldDecoder6  *structFieldDecoder
	fieldHash7     int64
	fieldDecoder7  *structFieldDecoder
	fieldHash8     int64
	fieldDecoder8  *structFieldDecoder
	fieldHash9     int64
	fieldDecoder9  *structFieldDecoder
	fieldHash10    int64
	fieldDecoder10 *structFieldDecoder
}

func (d *tenFieldsStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	if !iter.incrementDepth() {
		return
	}
	for {
		switch iter.readFieldHash() {
		case d.fieldHash1:
			d.fieldDecoder1.Decode(ptr, iter)
		case d.fieldHash2:
			d.fieldDecoder2.Decode(ptr, iter)
		case d.fieldHash3:
			d.fieldDecoder3.Decode(ptr, iter)
		case d.fieldHash4:
			d.fieldDecoder4.Decode(ptr, iter)
		case d.fieldHash5:
			d.fieldDecoder5.Decode(ptr, iter)
		case d.fieldHash6:
			d.fieldDecoder6.Decode(ptr, iter)
		case d.fieldHash7:
			d.fieldDecoder7.Decode(ptr, iter)
		case d.fieldHash8:
			d.fieldDecoder8.Decode(ptr, iter)
		case d.fieldHash9:
			d.fieldDecoder9.Decode(ptr, iter)
		case d.fieldHash10:
			d.fieldDecoder10.Decode(ptr, iter)
		default:
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
