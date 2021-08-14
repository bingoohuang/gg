package jsoni

import (
	"github.com/modern-go/reflect2"
	"reflect"
	"unsafe"
)

type dynamicEncoder struct {
	valType reflect2.Type
}

func (e *dynamicEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	obj := e.valType.UnsafeIndirect(ptr)
	stream.WriteVal(obj)
}

func (e *dynamicEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return e.valType.UnsafeIndirect(ptr) == nil
}

type efaceDecoder struct{}

func (d *efaceDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	pObj := (*interface{})(ptr)
	obj := *pObj
	if obj == nil {
		*pObj = iter.Read()
		return
	}
	typ := reflect2.TypeOf(obj)
	if typ.Kind() != reflect.Ptr {
		*pObj = iter.Read()
		return
	}
	ptrType := typ.(*reflect2.UnsafePtrType)
	ptrElemType := ptrType.Elem()
	if iter.WhatIsNext() == NilValue {
		if ptrElemType.Kind() != reflect.Ptr {
			iter.skip4Bytes('n', 'u', 'l', 'l')
			*pObj = nil
			return
		}
	}
	if reflect2.IsNil(obj) {
		obj := ptrElemType.New()
		iter.ReadVal(obj)
		*pObj = obj
		return
	}
	iter.ReadVal(obj)
}

type ifaceDecoder struct {
	valType *reflect2.UnsafeIFaceType
}

func (d *ifaceDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if iter.ReadNil() {
		d.valType.UnsafeSet(ptr, d.valType.UnsafeNew())
		return
	}
	obj := d.valType.UnsafeIndirect(ptr)
	if reflect2.IsNil(obj) {
		iter.ReportError("decode non empty interface", "can not unmarshal into nil")
		return
	}
	iter.ReadVal(obj)
}
