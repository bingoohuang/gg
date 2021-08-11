package reflector

import "reflect"

// IsEmpty returns if the object is considered as empty or not.
func IsEmpty(v interface{}) bool {
	var value reflect.Value
	if vv, ok := v.(reflect.Value); ok {
		value = vv
		v = value.Interface()
	} else {
		value = reflect.ValueOf(v)
	}

	switch value.Kind() {
	case reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return value.IsNil()
	}

	zero := reflect.Zero(value.Type()).Interface()
	return reflect.DeepEqual(v, zero)
}

// Indirect returns the value that v points to.
// If v is a nil pointer, Indirect returns a zero Value.
// If v is not a pointer, Indirect returns v.
func Indirect(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if ve := v.Elem(); ve.IsValid() {
			v = ve
		}
	}
	return v
}

// ImplType tells src whether it implements target type.
func ImplType(src, target reflect.Type) bool {
	if src == target {
		return true
	}

	if src.Kind() == reflect.Ptr {
		return src.Implements(target)
	}

	if target.Kind() != reflect.Interface {
		return false
	}

	return reflect.PtrTo(src).Implements(target)
}

// IsError tells t whether it is error type exactly.
func IsError(t reflect.Type) bool { return t == ErrType }

// AsError tells t whether it implements error type exactly.
func AsError(t reflect.Type) bool { return ImplType(t, ErrType) }

// V returns the variadic arguments to slice.
func V(v ...interface{}) []interface{} {
	return v
}

// V0 returns the one of variadic arguments at index 0.
func V0(v ...interface{}) interface{} {
	return v[0]
}
