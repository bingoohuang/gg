package jsoni

import (
	"context"
	"github.com/modern-go/reflect2"
	"reflect"
	"unsafe"
)

type dynamicEncoder struct {
	valType reflect2.Type
}

func (e *dynamicEncoder) Encode(ctx context.Context, ptr unsafe.Pointer, stream *Stream) {
	obj := e.valType.UnsafeIndirect(ptr)
	stream.WriteVal(ctx, obj)
}

func (e *dynamicEncoder) IsEmpty(_ context.Context, ptr unsafe.Pointer, checkZero bool) bool {
	obj := e.valType.UnsafeIndirect(ptr)
	return reflect2.IsNil(obj) || checkZero && reflect.ValueOf(obj).IsZero()
}

type efaceDecoder struct{}

func (d *efaceDecoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *Iterator) {
	pObj := (*interface{})(ptr)
	obj := *pObj
	if obj == nil {
		*pObj = iter.Read(ctx)
		return
	}
	typ := reflect2.TypeOf(obj)
	if typ.Kind() != reflect.Ptr {
		*pObj = iter.Read(ctx)
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
		iter.ReadVal(ctx, obj)
		*pObj = obj
		return
	}
	iter.ReadVal(ctx, obj)
}

type ifaceDecoder struct {
	valType *reflect2.UnsafeIFaceType
}

func (d *ifaceDecoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *Iterator) {
	if iter.ReadNil() {
		d.valType.UnsafeSet(ptr, d.valType.UnsafeNew())
		return
	}
	obj := d.valType.UnsafeIndirect(ptr)
	if reflect2.IsNil(obj) {
		iter.ReportError("decode non empty interface", "can not unmarshal into nil")
		return
	}
	iter.ReadVal(ctx, obj)
}
