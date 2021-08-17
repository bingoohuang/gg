package anyfn

import "reflect"

func ConvertPtr(isPtr bool, v reflect.Value) reflect.Value {
	if !isPtr {
		return reflect.Indirect(v)
	}

	if v.Kind() == reflect.Ptr {
		return v
	}

	p := reflect.New(v.Type())
	p.Elem().Set(v)

	return p
}

// IndirectTypeOf returns the non-ptr type of v.
func IndirectTypeOf(v interface{}) reflect.Type {
	if v == nil {
		return reflect.TypeOf(v)
	}

	var t reflect.Type
	if vt, ok1 := v.(reflect.Type); ok1 {
		t = vt
	} else if vt, ok2 := v.(reflect.Value); ok2 {
		t = vt.Type()
	} else {
		t = reflect.TypeOf(v)
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t
}
