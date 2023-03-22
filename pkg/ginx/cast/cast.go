package cast

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/bingoohuang/gg/pkg/strcase"
)

// TryFind tries to find value by field name or tag value.
func TryFind(filedName, tagValue string, getter func(string) (interface{}, bool)) (interface{}, bool) {
	if tagValue != "" {
		return getter(tagValue)
	}

	return TryAnyCase(filedName, getter)
}

// TryAnyCase tries to find value by name in any-case.
func TryAnyCase(name string, getter func(string) (interface{}, bool)) (interface{}, bool) {
	if value, ok := getter(name); ok {
		return value, true
	}

	if value, ok := getter(strcase.ToCamelLower(name)); ok {
		return value, true
	}

	if value, ok := getter(strcase.ToSnake(name)); ok {
		return value, true
	}

	if value, ok := getter(strcase.ToSnakeUpper(name)); ok {
		return value, true
	}

	if value, ok := getter(strcase.ToKebab(name)); ok {
		return value, true
	}

	if value, ok := getter(strcase.ToKebabUpper(name)); ok {
		return value, true
	}

	return "", false
}

// ToStruct populates the properties to the structure's field.
func ToStruct(b interface{}, tagName string, getter func(filedName, tagValue string) (interface{}, bool)) error {
	v := reflect.ValueOf(b)
	vt := v.Type()

	if vt.Kind() != reflect.Ptr {
		return errors.New("only argument of pointer of structure  supported")
	}

	v = v.Elem()
	if vt = v.Type(); vt.Kind() != reflect.Struct {
		return errors.New("only argument of pointer of structure  supported")
	}

	for i := 0; i < vt.NumField(); i++ {
		structField := vt.Field(i)
		if structField.PkgPath != "" { // bypass non-exported fields
			continue
		}

		fieldType := structField.Type
		fieldPtr := fieldType.Kind() == reflect.Ptr

		if fieldPtr {
			fieldType = fieldType.Elem()
		}

		field := v.Field(i)

		if fieldType.Kind() == reflect.Struct {
			if err := parseStruct(fieldType, tagName, fieldPtr, field, getter); err != nil {
				return err
			}

			continue
		}

		v, ok := getter(structField.Name, structField.Tag.Get(tagName))
		if !ok {
			continue
		}

		if vs, ok := v.(string); ok {
			vv, err := To(vs, structField.Type)
			if err != nil {
				return err
			}

			field.Set(vv)

			continue
		}

		if reflect.TypeOf(v) == structField.Type {
			field.Set(reflect.ValueOf(v))
			continue
		}

		return fmt.Errorf("unable to convert %v to %v", v, structField.Type)
	}

	return nil
}

func parseStruct(fieldType reflect.Type, tag string, ptr bool, field reflect.Value,
	getter func(name string, tagValue string) (interface{}, bool),
) error {
	fv := reflect.New(fieldType)
	if err := ToStruct(fv.Interface(), tag, getter); err != nil {
		return err
	}

	if ptr {
		field.Set(fv)
	} else {
		field.Set(fv.Elem())
	}

	return nil
}

// Caster defines the function prototype for cast string a any type.
type Caster func(s string, asPtr bool) (reflect.Value, error)

// nolint:gochecknoglobals
var (
	InvalidValue = reflect.Value{}
)

// casters defines default for basic types.
// nolint:gochecknoglobals
var casters = map[reflect.Type]Caster{
	reflect.TypeOf(false):            toBool,
	reflect.TypeOf(float32(0)):       toFloat32,
	reflect.TypeOf(float64(0)):       toFloat64,
	reflect.TypeOf(0):                toInt,
	reflect.TypeOf(int8(0)):          toInt8,
	reflect.TypeOf(int16(0)):         toInt16,
	reflect.TypeOf(int32(0)):         toInt32,
	reflect.TypeOf(int64(0)):         toInt64,
	reflect.TypeOf(""):               toString,
	reflect.TypeOf(uint(0)):          toUint,
	reflect.TypeOf(uint8(0)):         toUint8,
	reflect.TypeOf(uint16(0)):        toUint16,
	reflect.TypeOf(uint32(0)):        toUint32,
	reflect.TypeOf(uint64(0)):        toUint64,
	reflect.TypeOf(time.Duration(0)): toTimeDuration,
}

// To cast string a any type.
func To(s string, t reflect.Type) (reflect.Value, error) {
	asPtr := t.Kind() == reflect.Ptr
	if asPtr {
		t = t.Elem()
	}

	if caster, ok := casters[t]; ok {
		return caster(s, asPtr)
	}

	return InvalidValue, errors.New("casting not supported")
}

func toTimeDuration(s string, asPtr bool) (reflect.Value, error) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return InvalidValue, err
	}

	if asPtr {
		return reflect.ValueOf(&d), nil
	}

	return reflect.ValueOf(d), nil
}

func toBool(s string, asPtr bool) (reflect.Value, error) {
	v, err := strconv.ParseBool(s)
	if err != nil {
		switch strings.ToLower(s) {
		case "yes", "ok", "1", "on":
			v = true
			err = nil
		}
	}

	if err != nil {
		return InvalidValue, err
	}

	if asPtr {
		return reflect.ValueOf(&v), nil
	}

	return reflect.ValueOf(v), nil
}

func toFloat32(s string, asPtr bool) (reflect.Value, error) {
	v, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return InvalidValue, err
	}

	vv := float32(v)

	if asPtr {
		return reflect.ValueOf(&vv), nil
	}

	return reflect.ValueOf(vv), nil
}

func toFloat64(s string, asPtr bool) (reflect.Value, error) {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return InvalidValue, err
	}

	vv := v

	if asPtr {
		return reflect.ValueOf(&vv), nil
	}

	return reflect.ValueOf(vv), nil
}

func toInt(s string, asPtr bool) (reflect.Value, error) {
	v, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		return InvalidValue, err
	}

	vv := int(v)

	if asPtr {
		return reflect.ValueOf(&vv), nil
	}

	return reflect.ValueOf(vv), nil
}

func toInt8(s string, asPtr bool) (reflect.Value, error) {
	v, err := strconv.ParseInt(s, 10, 8)
	if err != nil {
		return InvalidValue, err
	}

	vv := int8(v)

	if asPtr {
		return reflect.ValueOf(&vv), nil
	}

	return reflect.ValueOf(vv), nil
}

func toInt16(s string, asPtr bool) (reflect.Value, error) {
	v, err := strconv.ParseInt(s, 10, 16)
	if err != nil {
		return InvalidValue, err
	}

	vv := int16(v)

	if asPtr {
		return reflect.ValueOf(&vv), nil
	}

	return reflect.ValueOf(vv), nil
}

func toInt32(s string, asPtr bool) (reflect.Value, error) {
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return InvalidValue, err
	}

	vv := int32(v)

	if asPtr {
		return reflect.ValueOf(&vv), nil
	}

	return reflect.ValueOf(vv), nil
}

func toInt64(s string, asPtr bool) (reflect.Value, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return InvalidValue, err
	}

	if asPtr {
		return reflect.ValueOf(&v), nil
	}

	return reflect.ValueOf(v), nil
}

func toString(s string, asPtr bool) (reflect.Value, error) {
	if asPtr {
		return reflect.ValueOf(&s), nil
	}

	return reflect.ValueOf(s), nil
}

func toUint(s string, asPtr bool) (reflect.Value, error) {
	v, err := strconv.ParseUint(s, 10, 0)
	if err != nil {
		return InvalidValue, err
	}

	vv := uint(v)

	if asPtr {
		return reflect.ValueOf(&vv), nil
	}

	return reflect.ValueOf(vv), nil
}

func toUint8(s string, asPtr bool) (reflect.Value, error) {
	v, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return InvalidValue, err
	}

	vv := uint8(v)

	if asPtr {
		return reflect.ValueOf(&vv), nil
	}

	return reflect.ValueOf(vv), nil
}

func toUint16(s string, asPtr bool) (reflect.Value, error) {
	v, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return InvalidValue, err
	}

	vv := uint16(v)

	if asPtr {
		return reflect.ValueOf(&vv), nil
	}

	return reflect.ValueOf(vv), nil
}

func toUint32(s string, asPtr bool) (reflect.Value, error) {
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return InvalidValue, err
	}

	vv := uint32(v)

	if asPtr {
		return reflect.ValueOf(&vv), nil
	}

	return reflect.ValueOf(vv), nil
}

func toUint64(s string, asPtr bool) (reflect.Value, error) {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return InvalidValue, err
	}

	if asPtr {
		return reflect.ValueOf(&v), nil
	}

	return reflect.ValueOf(v), nil
}
