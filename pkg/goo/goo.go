package goo

import (
	"errors"
	"fmt"
	"math"
	"math/bits"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"unicode"
)

// https://burakkokenn.medium.com/go-goo-c6921cdac348
// https://github.com/procyon-projects/goo

type Exportable interface {
	IsExported() bool
}

type Invokable interface {
	Invoke(obj interface{}, args ...interface{}) []interface{}
}

type Member interface {
	Exportable
	Name() string
	String() string
}

type New interface {
	New() interface{}
}

type Interface interface {
	Type
	Methods() []Method
	MethodNum() int
}

type intType struct {
	*baseType
}

func newIntType(baseTyp *baseType) intType { return intType{baseType: baseTyp} }

func (t intType) Methods() []Method {
	num := t.MethodNum()
	methods := make([]Method, num)
	for i := 0; i < num; i++ {
		methods[i] = ToMemberMethod(t.typ.Method(i))
	}
	return methods
}

func (t intType) MethodNum() int { return t.typ.NumMethod() }

type Bool interface {
	Type
	New
	ToBool(value string) bool
	ToStr(value bool) string
}

type boolType struct {
	*baseType
}

func newBoolType(baseTyp *baseType) Bool { return boolType{baseType: baseTyp} }

func (b boolType) ToBool(value string) bool {
	if value == "true" {
		return true
	} else if value == "false" {
		return false
	}
	panic("Given value is not true or false")
}

func (b boolType) ToStr(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func (b boolType) New() interface{} { return reflect.New(b.GoType()).Interface() }

type Array interface {
	Type
	New
	ElemType() Type
	Len() int
}

type arrayType struct {
	*baseType
	elementType Type
	length      int
}

func newArrayType(baseTyp *baseType) Array {
	return arrayType{
		baseType:    baseTyp,
		elementType: FromGoType(baseTyp.GoType().Elem()),
		length:      baseTyp.GoType().Len(),
	}
}

func (t arrayType) ElemType() Type   { return t.elementType }
func (t arrayType) Len() int         { return t.length }
func (t arrayType) New() interface{} { return reflect.New(t.GoType()).Interface() }

type Field interface {
	Member
	Taggable
	IsAnonymous() bool
	Type() Type
	CanSet() bool
	Set(instance interface{}, value interface{})
	Get(instance interface{}) interface{}
}

type memberField struct {
	name        string
	typ         Type
	tags        reflect.StructTag
	isAnonymous bool
	isExported  bool
	index       []int
}

func newMemberField(name string, typ Type, isAnonymous, exported bool, tags reflect.StructTag, index []int) memberField {
	return memberField{
		name:        name,
		typ:         typ,
		isAnonymous: isAnonymous,
		tags:        tags,
		isExported:  exported,
		index:       index,
	}
}

func (f memberField) Name() string      { return f.name }
func (f memberField) IsAnonymous() bool { return f.isAnonymous }
func (f memberField) IsExported() bool  { return f.isExported }
func (f memberField) CanSet() bool      { return f.isExported }
func (f memberField) Type() Type        { return f.typ }
func (f memberField) String() string    { return f.name }

func (f memberField) Tags() (fieldTags []Tag) {
	tags := f.tags
	for tags != "" {
		i := 0
		for i < len(tags) && tags[i] == ' ' {
			i++
		}
		tags = tags[i:]
		if tags == "" {
			break
		}

		i = 0
		for i < len(tags) && tags[i] > ' ' && tags[i] != ':' && tags[i] != '"' && tags[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tags) || tags[i] != ':' || tags[i+1] != '"' {
			break
		}
		name := string(tags[:i])
		tags = tags[i+1:]

		i = 1
		for i < len(tags) && tags[i] != '"' {
			if tags[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tags) {
			break
		}
		quotedValue := string(tags[:i+1])
		tags = tags[i+1:]

		value, err := strconv.Unquote(quotedValue)
		if err != nil {
			break
		}

		fieldTags = append(fieldTags, Tag{Name: name, Value: value})
	}
	return fieldTags
}

func (f memberField) TagByName(name string) (Tag, error) {
	if v, ok := f.tags.Lookup(name); ok {
		return Tag{Name: name, Value: v}, nil
	}
	return Tag{}, fmt.Errorf("Tag named %s not found ", name)
}

func (f memberField) Set(instance interface{}, value interface{}) {
	if !f.CanSet() {
		panic("Field cannot be set because of it is an unexported f")
	}
	typ := TypeOf(instance)
	if !typ.IsStruct() {
		panic("Instance must only be a struct")
	}
	if !typ.IsPtr() {
		panic("Instance type must be a pointer")
	}
	v := typ.GoPtrValue().FieldByIndex(f.index)
	v.Set(reflect.ValueOf(value))
}

func (f memberField) Get(instance interface{}) interface{} {
	typ := TypeOf(instance)
	if !typ.IsStruct() {
		panic("Instance must only be a struct")
	}
	structType := typ.GoType()
	structValueType := typ.GoValue()
	if typ.IsPtr() {
		structValueType = typ.GoPtrValue()
	}
	fieldType := structType.FieldByIndex(f.index)
	if !IsFieldExported(fieldType) {
		panic("Field is not exported, you cannot get the value : " + f.name)
	}

	fv := structValueType.FieldByIndex(f.index)
	if fieldType.Type.Kind() != reflect.Ptr {
		return fv.Interface()
	}

	return fv.Addr().Interface()
}

type Func interface {
	Type
	InTypes() []Type
	InNum() int
	OutTypes() []Type
	OutNum() int
	Call(args []interface{}) []interface{}
}

type funcType struct {
	*baseType
}

func newFuncType(baseTyp *baseType) funcType { return funcType{baseType: baseTyp} }

func (t funcType) InTypes() []Type {
	num := t.InNum()
	parameterTypes := make([]Type, num)
	for i := 0; i < num; i++ {
		parameterTypes[i] = FromGoType(t.typ.In(i))
	}
	return parameterTypes
}

func (t funcType) InNum() int  { return t.typ.NumIn() }
func (t funcType) OutNum() int { return t.typ.NumOut() }

func (t funcType) OutTypes() []Type {
	num := t.OutNum()
	returnTypes := make([]Type, num)
	for i := 0; i < num; i++ {
		returnTypes[i] = FromGoType(t.typ.Out(i))
	}
	return returnTypes
}

func (t funcType) Call(args []interface{}) []interface{} {
	num := t.InNum()
	if (args == nil && num != 0) || (args != nil && len(args) != num) {
		panic("Parameter counts don't match argument counts")
	}
	inputs := make([]reflect.Value, num)
	for i, arg := range args {
		if arg != nil {
			inputs[i] = reflect.ValueOf(arg)
			continue
		}

		if t := t.InTypes()[i]; t.IsPtr() {
			inputs[i] = reflect.New(t.PtrType()).Elem()
		} else {
			inputs[i] = reflect.New(t.GoType()).Elem()
		}
	}
	results := t.val.Call(inputs)
	outputs := make([]interface{}, len(results))
	for i, outputParam := range results {
		outputs[i] = outputParam.Interface()
	}
	return outputs
}

type Map interface {
	Type
	New
	KeyType() Type
	ValueType() Type
}

type mapType struct {
	*baseType
	keyType   Type
	valueType Type
}

func newMapType(baseTyp *baseType) Map {
	return mapType{
		baseType:  baseTyp,
		keyType:   FromGoType(baseTyp.GoType().Key()),
		valueType: FromGoType(baseTyp.GoType().Elem()),
	}
}

func (m mapType) KeyType() Type   { return m.keyType }
func (m mapType) ValueType() Type { return m.valueType }

func (m mapType) New() interface{} {
	mapType := reflect.MapOf(m.keyType.GoType(), m.valueType.GoType())
	instance := reflect.MakeMapWithSize(mapType, 0)
	return instance.Interface()
}

type Method interface {
	Member
	Invokable
	OutNum() int
	OutTypes() []Type
	InNum() int
	InTypes() []Type
}

type memberMethod struct {
	typ        reflect.Type
	name       string
	isExported bool
	fun        reflect.Value
	index      int
}

func newMemberMethod(methodType reflect.Type, name string, exported bool, fun reflect.Value, index int) memberMethod {
	return memberMethod{
		typ:        methodType,
		name:       name,
		isExported: exported,
		fun:        fun,
		index:      index,
	}
}

func (m memberMethod) Name() string     { return m.name }
func (m memberMethod) IsExported() bool { return m.isExported }
func (m memberMethod) String() string   { return m.name }
func (m memberMethod) OutNum() int      { return m.typ.NumOut() }
func (m memberMethod) InNum() int       { return m.typ.NumIn() }

func (m memberMethod) Invoke(obj interface{}, args ...interface{}) []interface{} {
	typ := TypeOf(obj)
	if !typ.IsStruct() {
		panic("obj must be a struct instance")
	}

	if num := m.InNum(); (args == nil && num != 1) || (args != nil && len(args) != num-1) {
		panic("Parameter counts don't match argument counts")
	}

	tempArray := []interface{}{obj}

	args = append(tempArray[:], args[:]...)
	inputs := make([]reflect.Value, len(args))
	for i, arg := range args {
		if arg != nil {
			inputs[i] = reflect.ValueOf(arg)
		} else if t := m.InTypes()[i]; t.IsPtr() {
			inputs[i] = reflect.New(t.PtrType()).Elem()
		} else {
			inputs[i] = reflect.New(t.GoType()).Elem()
		}
	}

	results := typ.GoType().Method(m.index).Func.Call(inputs)
	outputs := make([]interface{}, len(results))
	for i, outputParam := range results {
		outputs[i] = outputParam.Interface()
	}
	return outputs
}

func (m memberMethod) OutTypes() []Type {
	num := m.OutNum()
	types := make([]Type, num)
	for i := 0; i < num; i++ {
		types[i] = FromGoType(m.typ.Out(i))
	}
	return types
}

func (m memberMethod) InTypes() []Type {
	num := m.InNum()
	types := make([]Type, num)
	for i := 0; i < num; i++ {
		types[i] = FromGoType(m.typ.In(i))
	}
	return types
}

type NumberType int

const (
	IntType NumberType = iota
	FloatType
	ComplexType
)

type BitSize int

const (
	Bit8   BitSize = 8
	Bit16  BitSize = 16
	Bit32  BitSize = 32
	Bit64  BitSize = 64
	Bit128 BitSize = 128
)

type Number interface {
	Type
	New
	Type() NumberType
	BitSize() BitSize
	Overflow(val interface{}) bool
	ToString(val interface{}) string
}

type Integer interface {
	Number
	IsSigned() bool
}

type signedType struct {
	*baseType
}

func newSignedIntegerType(baseTyp *baseType) signedType { return signedType{baseType: baseTyp} }

func (t signedType) Type() NumberType { return IntType }

func (t signedType) BitSize() BitSize {
	switch t.kind {
	case reflect.Int64:
		return Bit64
	case reflect.Int8:
		return Bit8
	case reflect.Int16:
		return Bit16
	case reflect.Int32:
		return Bit32
	default:
		if bits.UintSize == 32 {
			return Bit32
		}
		return Bit64
	}
}

func (t signedType) IsSigned() bool { return true }

func (t signedType) Overflow(val interface{}) bool {
	valType := TypeOf(val)
	if !valType.IsNumber() || IntType != valType.(Number).Type() || !valType.(Integer).IsSigned() {
		panic("Given type is not compatible with signed t")
	}
	iv, err := strconv.ParseInt(fmt.Sprintf("%d", val), 10, 64)
	PanicIf(err)

	s := t.BitSize()
	return Bit8 == s && (math.MinInt8 > iv || math.MaxInt8 < iv) ||
		Bit16 == s && (math.MinInt16 > iv || math.MaxInt16 < iv) ||
		Bit32 == s && (math.MinInt32 > iv || math.MaxInt32 < iv)
}

func (t signedType) New() interface{} { return reflect.New(t.GoType()).Interface() }

func (t signedType) ToString(val interface{}) string {
	valType := TypeOf(val)
	if !valType.IsNumber() || IntType != valType.(Number).Type() || !valType.(Integer).IsSigned() {
		panic("Incompatible type : " + valType.Name())
	}
	return fmt.Sprintf("%d", val)
}

type unsignedType struct {
	*baseType
}

func newUnsignedIntegerType(baseTyp *baseType) unsignedType { return unsignedType{baseType: baseTyp} }
func (t unsignedType) Type() NumberType                     { return IntType }

func (t unsignedType) BitSize() BitSize {
	switch t.kind {
	case reflect.Uint64:
		return Bit64
	case reflect.Uint8:
		return Bit8
	case reflect.Uint16:
		return Bit16
	case reflect.Uint32:
		return Bit32
	default:
		if bits.UintSize == 32 {
			return Bit32
		}
		return Bit64
	}
}

func (t unsignedType) IsSigned() bool { return false }

func (t unsignedType) Overflow(val interface{}) bool {
	valType := TypeOf(val)
	if !valType.IsNumber() || IntType != valType.(Number).Type() || valType.(Integer).IsSigned() {
		panic("Given type is not compatible with unsigned t")
	}
	v, err := strconv.ParseUint(fmt.Sprintf("%d", val), 10, 64)
	PanicIf(err)

	size := t.BitSize()
	return Bit8 == size && math.MaxUint8 < v ||
		Bit16 == size && math.MaxUint16 < v ||
		Bit32 == size && math.MaxUint32 < v
}

func (t unsignedType) New() interface{} { return reflect.New(t.GoType()).Interface() }

func (t unsignedType) ToString(val interface{}) string {
	typ := TypeOf(val)
	if !typ.IsNumber() || IntType != typ.(Number).Type() || typ.(Integer).IsSigned() {
		panic("Incompatible type : " + typ.Name())
	}
	return fmt.Sprintf("%d", val)
}

type Float interface {
	Number
}

type floatType struct {
	*baseType
}

func newFloatType(baseTyp *baseType) Float { return floatType{baseType: baseTyp} }

func (t floatType) New() interface{} { return reflect.New(t.GoType()).Interface() }
func (t floatType) Type() NumberType { return FloatType }
func (t floatType) BitSize() BitSize { return BitSizeIf(t.kind, reflect.Float32, Bit32, Bit64) }

func (t floatType) Overflow(val interface{}) bool {
	if typ := TypeOf(val); !typ.IsNumber() || FloatType != typ.(Number).Type() {
		panic("Given type is not compatible with t")
	}
	v, err := strconv.ParseFloat(fmt.Sprintf("%f", val), 64)
	PanicIf(err)

	size := t.BitSize()
	return Bit32 == size && math.MaxFloat32 < v || Bit64 == size && math.MaxFloat64 < v
}

func (t floatType) ToString(val interface{}) string {
	valType := TypeOf(val)
	if !valType.IsNumber() || FloatType != valType.(Number).Type() {
		panic("Incompatible type : " + valType.Name())
	}
	return fmt.Sprintf("%f", val)
}

type Complex interface {
	Number
	GetImaginaryData(val interface{}) interface{}
	GetRealData(val interface{}) interface{}
}

type complexType struct {
	*baseType
}

func newComplexType(baseTyp *baseType) Complex    { return complexType{baseType: baseTyp} }
func (t complexType) Type() NumberType            { return ComplexType }
func (t complexType) Overflow(v interface{}) bool { panic("It does not support Overflow for now") }
func (t complexType) BitSize() BitSize            { return BitSizeIf(t.kind, reflect.Complex64, Bit64, Bit128) }

func BitSizeIf(k, ifv reflect.Kind, a, b BitSize) BitSize {
	if k == ifv {
		return a
	}

	return b
}

func (t complexType) GetImaginaryData(val interface{}) interface{} {
	typ := TypeOf(val)
	if !typ.IsNumber() || ComplexType != typ.(Number).Type() {
		panic("Given type is not compatible with t")
	}

	if t.BitSize() == Bit64 {
		return imag(val.(complex64))
	}
	return imag(val.(complex128))
}

func (t complexType) GetRealData(val interface{}) interface{} {
	typ := TypeOf(val)
	if !typ.IsNumber() || ComplexType != typ.(Number).Type() {
		panic("Given type is not compatible with t")
	}

	if t.BitSize() == Bit64 {
		return real(val.(complex64))
	}
	return real(val.(complex128))
}

func (t complexType) New() interface{} { return reflect.New(t.GoType()).Interface() }

func (t complexType) ToString(val interface{}) string {
	if typ := TypeOf(val); !typ.IsNumber() || ComplexType != typ.(Number).Type() {
		panic("Incompatible type : " + typ.Name())
	}
	return fmt.Sprintf("%f", val)
}

type Slice interface {
	Type
	New
	GetElementType() Type
}

type sliceType struct {
	*baseType
	elementType Type
}

func newSliceType(baseTyp *baseType) Slice {
	return sliceType{
		baseType:    baseTyp,
		elementType: FromGoType(baseTyp.GoType().Elem()),
	}
}

func (t sliceType) GetElementType() Type { return t.elementType }

func (t sliceType) New() interface{} {
	return reflect.MakeSlice(t.GoType(), t.val.Len(), t.val.Cap()).Interface()
}

type String interface {
	Type
	New
	ToNumber(val string, number Number) (interface{}, error)
	ToInt(val string) int
	ToInt8(val string) int8
	ToInt16(val string) int16
	ToInt32(val string) int32
	ToInt64(val string) int64
	ToUint(val string) uint
	ToUint8(val string) uint8
	ToUint16(val string) uint16
	ToUint32(val string) uint32
	ToUint64(val string) uint64
	ToFloat32(val string) float32
	ToFloat64(val string) float64
}

type stringType struct {
	*baseType
}

func newStringType(baseTyp *baseType) stringType {
	return stringType{
		baseType: baseTyp,
	}
}

func (t stringType) ToNumber(val string, number Number) (interface{}, error) {
	if number == nil {
		panic("Number must not be null")
	}

	if numberType := number.Type(); IntType == numberType {
		return t.getIntValue(val, number.(Integer))
	} else if FloatType == numberType {
		return t.getFloatValue(val, number.(Float))
	}
	return nil, errors.New("complex numbers does not support for now")
}

func (t stringType) ToInt(val string) int {
	sizeInBits := bits.UintSize
	var result interface{}
	var err error
	if sizeInBits == 32 {
		result, err = t.getIntegerValueByBitSize(val, Bit32, true)
	} else {
		result, err = t.getIntegerValueByBitSize(val, Bit64, true)
	}
	PanicIf(err)

	if sizeInBits == 32 {
		return result.(int)
	}
	return int(result.(int64))
}

func (t stringType) ToInt8(val string) int8 {
	result, err := t.getIntegerValueByBitSize(val, Bit8, true)
	PanicIf(err)
	return result.(int8)
}

func (t stringType) ToInt16(val string) int16 {
	result, err := t.getIntegerValueByBitSize(val, Bit16, true)
	PanicIf(err)
	return result.(int16)
}

func (t stringType) ToInt32(val string) int32 {
	result, err := t.getIntegerValueByBitSize(val, Bit32, true)
	PanicIf(err)
	return result.(int32)
}

func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func (t stringType) ToInt64(val string) int64 {
	result, err := t.getIntegerValueByBitSize(val, Bit64, true)
	PanicIf(err)
	return result.(int64)
}

func (t stringType) ToUint(val string) uint {
	sizeInBits := bits.UintSize
	var result interface{}
	var err error
	if sizeInBits == 32 {
		result, err = t.getIntegerValueByBitSize(val, Bit32, false)
	} else {
		result, err = t.getIntegerValueByBitSize(val, Bit64, false)
	}

	PanicIf(err)

	if sizeInBits == 32 {
		return result.(uint)
	}
	lastValue := result.(uint64)
	return uint(lastValue)
}

func (t stringType) ToUint8(val string) uint8 {
	result, err := t.getIntegerValueByBitSize(val, Bit8, false)
	PanicIf(err)
	return result.(uint8)
}

func (t stringType) ToUint16(val string) uint16 {
	result, err := t.getIntegerValueByBitSize(val, Bit16, false)
	PanicIf(err)
	return result.(uint16)
}

func (t stringType) ToUint32(val string) uint32 {
	result, err := t.getIntegerValueByBitSize(val, Bit32, false)
	PanicIf(err)
	return result.(uint32)
}

func (t stringType) ToUint64(val string) uint64 {
	result, err := t.getIntegerValueByBitSize(val, Bit64, false)
	PanicIf(err)
	return result.(uint64)
}

func (t stringType) ToFloat32(val string) float32 {
	result, err := strconv.ParseFloat(val, int(Bit32))
	PanicIf(err)
	return float32(result)
}

func (t stringType) ToFloat64(val string) float64 {
	result, err := strconv.ParseFloat(val, int(Bit64))
	PanicIf(err)
	return result
}

func (t stringType) getIntValue(strValue string, integer Integer) (resultValue interface{}, err error) {
	var value interface{}
	var signedValue int64
	var unsignedValue uint64
	if integer.IsSigned() {
		signedValue, err = strconv.ParseInt(strValue, 10, 64)
		value = signedValue
	} else {
		unsignedValue, err = strconv.ParseUint(strValue, 10, 64)
		value = unsignedValue
	}

	if err != nil {
		return nil, err
	}

	if integer.Overflow(value) {
		return nil, errors.New("The given value is out of range of the integer type : " + integer.String())
	}

	integerVal := reflect.New(integer.GoType()).Elem()
	if integer.IsSigned() {
		integerVal.SetInt(signedValue)
	} else {
		integerVal.SetUint(unsignedValue)
	}
	resultValue = integerVal.Interface()
	return
}

func (t stringType) getIntegerValueByBitSize(strValue string, bitSize BitSize, isSigned bool) (resultValue interface{}, err error) {
	if Bit128 == bitSize {
		panic("BitSize does not support 128")
	}

	var signedValue int64
	var unsignedValue uint64
	if isSigned {
		signedValue, err = strconv.ParseInt(strValue, 10, 64)
	} else {
		unsignedValue, err = strconv.ParseUint(strValue, 10, 64)
	}
	if err != nil {
		return nil, err
	}

	overflow := false
	if isSigned {
		overflow = Bit8 == bitSize && (math.MinInt8 > signedValue || math.MaxInt8 < signedValue) ||
			Bit16 == bitSize && (math.MinInt16 > signedValue || math.MaxInt16 < signedValue) ||
			Bit32 == bitSize && (math.MinInt32 > signedValue || math.MaxInt32 < signedValue)
	} else {
		overflow = Bit8 == bitSize && math.MaxUint8 < unsignedValue ||
			Bit16 == bitSize && math.MaxUint16 < unsignedValue ||
			Bit32 == bitSize && math.MaxUint32 < unsignedValue
	}

	if overflow {
		return nil, errors.New("the given value is out of range of the integer type")
	}

	if isSigned {
		if Bit8 == bitSize {
			return int8(signedValue), nil
		} else if Bit16 == bitSize {
			return int16(signedValue), nil
		} else if Bit32 == bitSize {
			return int32(signedValue), nil
		}
		return signedValue, nil
	}

	if Bit8 == bitSize {
		return uint8(unsignedValue), nil
	} else if Bit16 == bitSize {
		return uint16(unsignedValue), nil
	} else if Bit32 == bitSize {
		return uint32(unsignedValue), nil
	}
	return unsignedValue, nil
}

func (t stringType) getFloatValue(strValue string, float Float) (resultValue interface{}, err error) {
	var value float64
	value, err = strconv.ParseFloat(strValue, 64)
	if err != nil {
		return nil, err
	}

	if float.Overflow(value) {
		return nil, errors.New("The given value is out of range of the float type : " + float.String())
	}
	floatValue := reflect.New(float.GoType()).Elem()
	floatValue.SetFloat(value)
	resultValue = floatValue.Interface()
	return
}

func (t stringType) New() interface{} { return reflect.New(t.GoType()).Interface() }

type Struct interface {
	Type
	New
	Fields() []Field
	FieldNum() int
	FieldsExported() []Field
	FieldExportedNum() int
	FieldsUnexported() []Field
	FieldUnexportedNum() int
	FieldsAnonymous() []Field
	FieldAnonymousNum() int
	Methods() []Method
	MethodNum() int
	Implements(i Interface) bool
	IsEmbedded(candidate Struct) bool
}

type structType struct {
	*baseType
}

func newStructType(baseTyp *baseType) structType {
	return structType{
		baseType: baseTyp,
	}
}

func (t structType) Fields() []Field {
	num := t.FieldNum()
	fields := make([]Field, num)
	for i := 0; i < num; i++ {
		fields[i] = ConvertGoField(t.typ.Field(i))
	}
	return fields
}

func (t structType) FieldNum() int { return t.typ.NumField() }

func (t structType) FieldsExported() []Field {
	return t.FieldsIf(func(f Field) bool { return f.IsExported() })
}

func (t structType) FieldExportedNum() int {
	return t.NumIf(func(f Field) bool { return f.IsExported() })
}

func (t structType) FieldsUnexported() []Field {
	return t.FieldsIf(func(f Field) bool { return !f.IsExported() })
}

func (t structType) FieldUnexportedNum() int {
	return t.NumIf(func(f Field) bool { return !f.IsExported() })
}

func (t structType) FieldsAnonymous() []Field {
	return t.FieldsIf(func(f Field) bool { return f.IsAnonymous() })
}

func (t structType) FieldAnonymousNum() int {
	return t.NumIf(func(f Field) bool { return f.IsAnonymous() })
}

func (t structType) FieldsIf(predicate func(Field) bool) (fields []Field) {
	for _, field := range t.Fields() {
		if predicate(field) {
			fields = append(fields, field)
		}
	}
	return fields
}

func (t structType) NumIf(predicate func(Field) bool) (n int) {
	for _, field := range t.Fields() {
		if predicate(field) {
			n++
		}
	}
	return n
}

func (t structType) Methods() (methods []Method) {
	for i := 0; i < t.MethodNum(); i++ {
		var method reflect.Method
		if t.isPtr {
			method = t.ptrType.Method(i)
		} else {
			method = t.typ.Method(i)
		}
		methods = append(methods, ToMemberMethod(method))
	}
	return methods
}

func (t structType) MethodNum() int {
	if t.isPtr {
		return t.ptrType.NumMethod()
	}
	return t.typ.NumMethod()
}

func (t structType) Implements(i Interface) bool {
	if t.isPtr {
		return t.PtrType().Implements(i.GoType())
	}
	return t.GoType().Implements(i.GoType())
}

func (t structType) New() interface{} { return reflect.New(t.GoType()).Interface() }

func (t structType) IsEmbedded(candidate Struct) bool {
	if candidate == nil {
		panic("candidate must not be null")
	}
	return t.embeddedStruct(t, candidate)
}

func (t structType) embeddedStruct(parent Struct, candidate Struct) bool {
	for _, f := range parent.Fields() {
		if f.IsAnonymous() && f.Type().IsStruct() {
			if candidate.Equals(f.Type()) {
				return true
			}
			if f.Type().(Struct).FieldNum() > 0 {
				return t.embeddedStruct(f.Type().(Struct), candidate)
			}
		}
	}
	return false
}

type Tag struct {
	Name  string
	Value string
}

func (t Tag) String() string { return t.Name + "->" + t.Value }

type Taggable interface {
	Tags() []Tag
	TagByName(name string) (Tag, error)
}

type TypeConverter interface {
	IsBool() bool
	ToBoolType() Bool
	IsNumber() bool
	ToNumberType() Number
	IsFunc() bool
	ToFuncType() Func
	IsStruct() bool
	ToStructType() Struct
	IsInterface() bool
	ToInterfaceType() Interface
	IsString() bool
	ToStringType() String
	IsMap() bool
	ToMapType() Map
	IsArray() bool
	ToArrayType() Array
	IsSlice() bool
	ToSliceType() Slice
}

type Type interface {
	TypeConverter
	Name() string
	NameFull() string
	PkgName() string
	PkgNameFull() string
	PtrType() reflect.Type
	GoPtrValue() reflect.Value
	GoType() reflect.Type
	GoValue() reflect.Value
	IsPtr() bool
	IsInstantiable() bool
	String() string
	Equals(anotherType Type) bool
}

type baseType struct {
	parentType  interface{}
	name        string
	pkgName     string
	pkgFullName string
	typ         reflect.Type
	val         reflect.Value
	ptrType     reflect.Type
	ptrVal      reflect.Value
	kind        reflect.Kind
	isNumber    bool
	isPtr       bool
}

func newBaseType(typ reflect.Type, val reflect.Value) *baseType {
	return &baseType{
		name:        getTypeName(typ, val),
		pkgName:     getPkgName(typ, val),
		pkgFullName: getPkgFullName(typ, val),
		typ:         typ,
		val:         val,
		kind:        typ.Kind(),
		isNumber:    IsNumber(typ),
	}
}

func (t baseType) NameFull() string {
	if t.pkgFullName == "" {
		return t.name
	}
	return t.pkgFullName + "." + t.name
}
func (t baseType) Name() string               { return t.name }
func (t baseType) PkgName() string            { return t.pkgName }
func (t baseType) PkgNameFull() string        { return t.pkgFullName }
func (t baseType) PtrType() reflect.Type      { return t.ptrType }
func (t baseType) GoPtrValue() reflect.Value  { return t.ptrVal }
func (t baseType) GoType() reflect.Type       { return t.typ }
func (t baseType) GoValue() reflect.Value     { return t.val }
func (t baseType) IsBool() bool               { return reflect.Bool == t.kind }
func (t baseType) IsNumber() bool             { return t.isNumber }
func (t baseType) IsFunc() bool               { return reflect.Func == t.kind }
func (t baseType) IsStruct() bool             { return reflect.Struct == t.kind }
func (t baseType) IsInterface() bool          { return reflect.Interface == t.kind }
func (t baseType) IsString() bool             { return reflect.String == t.kind }
func (t baseType) IsMap() bool                { return reflect.Map == t.kind }
func (t baseType) IsArray() bool              { return reflect.Array == t.kind }
func (t baseType) IsSlice() bool              { return reflect.Slice == t.kind }
func (t baseType) IsPtr() bool                { return t.isPtr }
func (t baseType) IsInstantiable() bool       { return !(t.IsInterface() || t.IsFunc()) }
func (t baseType) ToBoolType() Bool           { return t.parentType.(Bool) }
func (t baseType) ToNumberType() Number       { return t.parentType.(Number) }
func (t baseType) ToFuncType() Func           { return t.parentType.(Func) }
func (t baseType) ToInterfaceType() Interface { return t.parentType.(Interface) }
func (t baseType) ToStringType() String       { return t.parentType.(String) }
func (t baseType) ToMapType() Map             { return t.parentType.(Map) }
func (t baseType) ToArrayType() Array         { return t.parentType.(Array) }
func (t baseType) ToSliceType() Slice         { return t.parentType.(Slice) }
func (t baseType) ToStructType() Struct       { return t.parentType.(Struct) }
func (t baseType) String() string             { return t.name }
func (t baseType) Equals(that Type) bool      { return that != nil && t.typ == that.GoType() }

func TypeOf(obj interface{}) Type {
	typ, val, isPtr := GoTypeAndValue(obj)
	baseTyp := newBaseType(typ, val)
	populatePtrInfo(obj, baseTyp, isPtr)
	actualType := getActualTypeFromBaseType(baseTyp)
	baseTyp.parentType = actualType
	return actualType
}

func populatePtrInfo(obj interface{}, baseType *baseType, isPtr bool) {
	if isPtr {
		typ, val := GoPtrTypeAndValue(obj)
		baseType.isPtr = true
		baseType.ptrType = typ
		baseType.ptrVal = val
	}
}

func FromGoType(typ reflect.Type) Type {
	if typ == nil {
		panic("Type cannot be nil")
	}
	var ptrType reflect.Type
	isPtr := typ.Kind() == reflect.Ptr
	if isPtr {
		ptrType = typ
		typ = typ.Elem()
	}
	baseTyp := newBaseType(typ, reflect.Value{})
	actualType := getActualTypeFromBaseType(baseTyp)
	baseTyp.parentType = actualType
	if isPtr {
		baseTyp.ptrType = ptrType
		baseTyp.isPtr = true
	}
	return actualType
}

func sanitizedName(str string) string {
	n := strings.ReplaceAll(str, "/", ".")
	n = strings.ReplaceAll(n, "-", ".")
	return strings.ReplaceAll(n, "_", ".")
}

func getActualTypeFromBaseType(b *baseType) Type {
	if b.IsFunc() {
		return newFuncType(b)
	} else if b.IsInterface() {
		return newIntType(b)
	} else if b.IsStruct() {
		return newStructType(b)
	} else if b.IsNumber() {
		if IsSigned(b.typ) {
			return newSignedIntegerType(b)
		} else if IsUnsigned(b.typ) {
			return newUnsignedIntegerType(b)
		} else if IsFloat(b.typ) {
			return newFloatType(b)
		} else if IsComplex(b.typ) {
			return newComplexType(b)
		}
	} else if b.IsString() {
		return newStringType(b)
	} else if b.IsBool() {
		return newBoolType(b)
	} else if b.IsMap() {
		return newMapType(b)
	} else if b.IsArray() {
		return newArrayType(b)
	} else if b.IsSlice() {
		return newSliceType(b)
	}
	panic(b.Name() + " isn't supported for now")
}

func getTypeName(typ reflect.Type, val reflect.Value) string {
	defer func() { recover() }()
	switch typ.Kind() {
	case reflect.Struct, reflect.Interface:
		return getBaseTypeName(typ)
	case reflect.Func:
		return getFuncName(val)
	}
	return typ.Name()
}

func GoTypeAndValue(obj interface{}) (reflect.Type, reflect.Value, bool) {
	typ := reflect.TypeOf(obj)
	if typ == nil {
		panic("Type cannot be determined as the given object is nil")
	}
	isPtr := typ.Kind() == reflect.Ptr
	if isPtr {
		typ = typ.Elem()
	}
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	return typ, val, isPtr
}

func GoPtrTypeAndValue(obj interface{}) (reflect.Type, reflect.Value) {
	typ := reflect.TypeOf(obj)
	if typ == nil {
		panic("Type cannot be determined as the given object is nil")
	}

	var ptrType reflect.Type
	if typ.Kind() == reflect.Ptr {
		ptrType = typ
	}

	var ptrVal reflect.Value
	if v := reflect.ValueOf(obj); v.Kind() == reflect.Ptr {
		ptrVal = v.Elem()
	}
	return ptrType, ptrVal
}

func getBaseTypeName(typ reflect.Type) string {
	name := typ.Name()
	if name != "" {
		return name
	}
	return typ.String()
}

func getPkgName(typ reflect.Type, val reflect.Value) string {
	defer func() { recover() }()

	if reflect.Func == typ.Kind() {
		return getFuncPkgName(val)
	}
	if i := strings.LastIndex(typ.String(), "."); i >= 0 {
		return typ.String()[:i]
	}
	return ""
}

func getPkgFullName(typ reflect.Type, val reflect.Value) string {
	defer func() { recover() }()

	if reflect.Func == typ.Kind() {
		return getFuncPkgFullName(val)
	}
	return sanitizedName(typ.PkgPath())
}

func getFuncName(val reflect.Value) string {
	fullName := runtime.FuncForPC(val.Pointer()).Name()
	if pos := strings.LastIndex(fullName, "."); pos < 0 {
		return fullName[pos+1:]
	}
	return fullName
}

func getFuncPkgFullName(val reflect.Value) string {
	fullName := runtime.FuncForPC(val.Pointer()).Name()
	if pos := strings.LastIndex(fullName, "."); pos < 0 {
		return fullName[:pos]
	}
	return sanitizedName(fullName)
}

func getFuncPkgName(val reflect.Value) string {
	fullName := getFuncPkgFullName(val)
	if pos := strings.LastIndex(fullName, "."); pos < 0 {
		return fullName[pos+1:]
	}
	return fullName
}

func ConvertGoField(f reflect.StructField) Field {
	return newMemberField(f.Name, FromGoType(f.Type), f.Anonymous, IsFieldExported(f), f.Tag, f.Index)
}

func IsFieldExported(f reflect.StructField) bool { return unicode.IsUpper(rune(f.Name[0])) }
func IsMethodExported(m reflect.Method) bool     { return unicode.IsUpper(rune(m.Name[0])) }

func ToMemberMethod(m reflect.Method) Method {
	return newMemberMethod(m.Type, m.Name, IsMethodExported(m), m.Func, m.Index)
}

func IsNumber(typ reflect.Type) bool {
	return IsAnyKind(typ, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128)
}

func IsSigned(typ reflect.Type) bool {
	return IsAnyKind(typ, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64)
}

func IsUnsigned(typ reflect.Type) bool {
	return IsAnyKind(typ, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64)
}
func IsFloat(typ reflect.Type) bool   { return IsAnyKind(typ, reflect.Float32, reflect.Float64) }
func IsComplex(typ reflect.Type) bool { return IsAnyKind(typ, reflect.Complex64, reflect.Complex128) }

func IsAnyKind(typ reflect.Type, kinds ...reflect.Kind) bool {
	kind := typ.Kind()
	for _, k := range kinds {
		if kind == k {
			return true
		}
	}

	return false
}
