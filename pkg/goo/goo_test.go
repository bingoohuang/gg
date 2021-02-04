package goo_test

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/bits"
	"reflect"
	"testing"

	"github.com/bingoohuang/gg/pkg/goo"
	"github.com/stretchr/testify/assert"
)

func TestBooleanType_NewInstance(t *testing.T) {
	typ := goo.TypeOf(true)
	assert.True(t, typ.IsBool())

	booleanType := typ.ToBoolType()
	instance := booleanType.New()
	assert.NotNil(t, instance)

	boolPtr := instance.(*bool)
	assert.False(t, *boolPtr)
}

func TestBooleanType_ToString(t *testing.T) {
	typ := goo.TypeOf(true)
	assert.True(t, typ.IsBool())
	booleanType := typ.ToBoolType()

	assert.Equal(t, "true", booleanType.ToStr(true))
	assert.Equal(t, "false", booleanType.ToStr(false))
}

func TestBooleanType_ToBoolean(t *testing.T) {
	typ := goo.TypeOf(true)
	assert.True(t, typ.IsBool())
	booleanType := typ.ToBoolType()

	assert.True(t, booleanType.ToBool("true"))
	assert.False(t, booleanType.ToBool("false"))
	assert.Panics(t, func() {
		booleanType.ToBool("test")
	})
}

func TestArrayType_GetElementType(t *testing.T) {
	arr := [5]string{}
	typ := goo.TypeOf(arr)
	assert.True(t, typ.IsArray())

	arrayType := typ.ToArrayType()
	assert.Equal(t, "string", arrayType.ElemType().NameFull())
}

func TestArrayType_GetLength(t *testing.T) {
	arr := [5]string{}
	typ := goo.TypeOf(arr)
	assert.True(t, typ.IsArray())

	arrayType := typ.ToArrayType()
	assert.Equal(t, 5, arrayType.Len())
}

func TestArrayType_NewInstance(t *testing.T) {
	arr := [5]string{}
	typ := goo.TypeOf(arr)
	assert.True(t, typ.IsArray())

	arrayType := typ.ToArrayType()
	instance := arrayType.New().(*[5]string)
	assert.NotNil(t, instance)
	assert.Len(t, *instance, 5)
}

type Stock struct {
	quantity uint32 `json:"quantity" yaml:"quantity"`
}

type Product struct {
	Name     string  `json:"name" yaml:"name" customTag:"customTagValue"`
	price    float64 `json:"price" yaml:"price"`
	Stock    `json:"stock" yaml:"stock" customTag:"customTagValue"`
	supplier Supplier `json:invalid`
}

type Supplier struct {
	Name string `json:"name" yaml:"name" `
}

func TestMemberField_GetName(t *testing.T) {
	typ := goo.TypeOf(&Product{})
	assert.True(t, typ.IsStruct())

	structType := typ.ToStructType()
	assert.Equal(t, 4, structType.FieldNum())

	assert.Equal(t, "Name", structType.Fields()[0].Name())
	assert.Equal(t, "price", structType.Fields()[1].Name())
	assert.Equal(t, "Stock", structType.Fields()[2].Name())
	assert.Equal(t, "supplier", structType.Fields()[3].Name())
}

func TestMemberField_String(t *testing.T) {
	typ := goo.TypeOf(&Product{})
	assert.True(t, typ.IsStruct())

	structType := typ.ToStructType()
	assert.Equal(t, 4, structType.FieldNum())

	assert.Equal(t, "Name", structType.Fields()[0].String())
	assert.Equal(t, "price", structType.Fields()[1].String())
	assert.Equal(t, "Stock", structType.Fields()[2].String())
	assert.Equal(t, "supplier", structType.Fields()[3].String())
}

func TestMemberField_IsAnonymous(t *testing.T) {
	typ := goo.TypeOf(&Product{})
	assert.True(t, typ.IsStruct())

	structType := typ.ToStructType()
	assert.Equal(t, 4, structType.FieldNum())
	assert.False(t, structType.Fields()[0].IsAnonymous())
	assert.False(t, structType.Fields()[1].IsAnonymous())
	assert.True(t, structType.Fields()[2].IsAnonymous())
	assert.False(t, structType.Fields()[3].IsAnonymous())
}

func TestMemberField_IsExported(t *testing.T) {
	typ := goo.TypeOf(&Product{})
	assert.True(t, typ.IsStruct())

	structType := typ.ToStructType()
	assert.Equal(t, 4, structType.FieldNum())
	assert.True(t, structType.Fields()[0].IsExported())
	assert.False(t, structType.Fields()[1].IsExported())
	assert.True(t, structType.Fields()[2].IsExported())
	assert.False(t, structType.Fields()[3].IsExported())
}

func TestMemberField_CanSet(t *testing.T) {
	typ := goo.TypeOf(&Product{})
	assert.True(t, typ.IsStruct())

	structType := typ.ToStructType()
	assert.Equal(t, 4, structType.FieldNum())
	assert.True(t, structType.Fields()[0].CanSet())
	assert.False(t, structType.Fields()[1].CanSet())
	assert.True(t, structType.Fields()[2].CanSet())
	assert.False(t, structType.Fields()[3].CanSet())
}

func TestMemberField_SetValue(t *testing.T) {
	product := &Product{
		Name:  "test-product",
		price: 39.90,
		Stock: Stock{
			quantity: 20,
		},
		supplier: Supplier{
			Name: "test-supplier",
		},
	}
	typ := goo.TypeOf(product)
	assert.True(t, typ.IsStruct())

	structType := typ.ToStructType()
	assert.Equal(t, 4, structType.FieldNum())

	name := structType.Fields()[0].Get(product).(string)
	assert.Equal(t, "test-product", name)

	assert.Panics(t, func() { structType.Fields()[1].Get(product) })

	stock := structType.Fields()[2].Get(product).(Stock)
	assert.Equal(t, uint32(20), stock.quantity)

	assert.Panics(t, func() { structType.Fields()[3].Get(product) })
	assert.Panics(t, func() { structType.Fields()[0].Set(23, nil) })
	assert.Panics(t, func() { structType.Fields()[0].Set(*product, nil) })
}

func TestMemberField_GetValue(t *testing.T) {
	product := &Product{
		Name:  "test-product",
		price: 39.90,
		Stock: Stock{
			quantity: 20,
		},
		supplier: Supplier{
			Name: "test-supplier",
		},
	}
	typ := goo.TypeOf(product)
	assert.True(t, typ.IsStruct())

	structType := typ.ToStructType()
	assert.Equal(t, 4, structType.FieldNum())

	structType.Fields()[0].Set(product, "test-product-2")
	assert.Equal(t, "test-product-2", product.Name)

	assert.Panics(t, func() { structType.Fields()[1].Set(product, 23.20) })

	structType.Fields()[2].Set(product, Stock{quantity: 30})
	assert.Equal(t, uint32(30), product.quantity)

	assert.Panics(t, func() { structType.Fields()[3].Set(product, Supplier{Name: "test-supplier-2"}) })
	assert.Panics(t, func() { structType.Fields()[0].Get(23) })
	assert.Panics(t, func() { structType.Fields()[1].Get(Stock{}) })
}

func TestMemberField_GetTags(t *testing.T) {
	product := &Product{
		Name:  "test-product",
		price: 39.90,
		Stock: Stock{
			quantity: 20,
		},
		supplier: Supplier{
			Name: "test-supplier",
		},
	}
	typ := goo.TypeOf(product)
	assert.True(t, typ.IsStruct())

	structType := typ.ToStructType()
	assert.Equal(t, 4, structType.FieldNum())

	assert.Equal(t, 3, len(structType.Fields()[0].Tags()))
	assert.Equal(t, 2, len(structType.Fields()[1].Tags()))
	assert.Equal(t, 3, len(structType.Fields()[2].Tags()))
	assert.Equal(t, 0, len(structType.Fields()[3].Tags()))

	// name
	tag, err := structType.Fields()[0].TagByName("json")
	assert.Nil(t, err)
	assert.Equal(t, "json", tag.Name)
	assert.Equal(t, "name", tag.Value)

	tag, err = structType.Fields()[0].TagByName("yaml")
	assert.Nil(t, err)
	assert.Equal(t, "yaml", tag.Name)
	assert.Equal(t, "name", tag.Value)

	tag, err = structType.Fields()[0].TagByName("customTag")
	assert.Nil(t, err)
	assert.Equal(t, "customTag", tag.Name)
	assert.Equal(t, "customTagValue", tag.Value)

	tag, err = structType.Fields()[0].TagByName("nonExistTag")
	assert.NotNil(t, err)

	// price
	tag, err = structType.Fields()[1].TagByName("json")
	assert.Nil(t, err)
	assert.Equal(t, "json", tag.Name)
	assert.Equal(t, "price", tag.Value)

	tag, err = structType.Fields()[1].TagByName("yaml")
	assert.Nil(t, err)
	assert.Equal(t, "yaml", tag.Name)
	assert.Equal(t, "price", tag.Value)

	// stock
	tag, err = structType.Fields()[2].TagByName("json")
	assert.Nil(t, err)
	assert.Equal(t, "json", tag.Name)
	assert.Equal(t, "stock", tag.Value)

	tag, err = structType.Fields()[2].TagByName("yaml")
	assert.Nil(t, err)
	assert.Equal(t, "yaml", tag.Name)
	assert.Equal(t, "stock", tag.Value)

	tag, err = structType.Fields()[2].TagByName("customTag")
	assert.Nil(t, err)
	assert.Equal(t, "customTag", tag.Name)
	assert.Equal(t, "customTagValue", tag.Value)
}

func testFunction(name string, i int, val bool, test *string) (string, error) {
	return "test", errors.New("test error")
}

func TestFunctionType_GetFunctionParameterCount(t *testing.T) {
	typ := goo.TypeOf(testFunction)
	assert.True(t, typ.IsFunc())

	functionType := typ.ToFuncType()
	assert.Equal(t, 4, functionType.InNum())
}

func TestFunctionType_GetFunctionParameterTypes(t *testing.T) {
	typ := goo.TypeOf(testFunction)
	assert.True(t, typ.IsFunc())

	functionType := typ.ToFuncType()
	parameterTypes := functionType.InTypes()
	assert.Equal(t, goo.TypeOf("").GoType(), parameterTypes[0].GoType())
	assert.Equal(t, goo.TypeOf(0).GoType(), parameterTypes[1].GoType())
	assert.Equal(t, goo.TypeOf(true).GoType(), parameterTypes[2].GoType())
}

func TestFunctionType_GetFunctionReturnTypeCount(t *testing.T) {
	typ := goo.TypeOf(testFunction)
	assert.True(t, typ.IsFunc())

	functionType := typ.ToFuncType()
	assert.Equal(t, 2, functionType.OutNum())
}

func TestFunctionType_GetFunctionReturnTypes(t *testing.T) {
	typ := goo.TypeOf(testFunction)
	assert.True(t, typ.IsFunc())

	functionType := typ.ToFuncType()
	parameterTypes := functionType.OutTypes()
	assert.Equal(t, goo.TypeOf("").GoType(), parameterTypes[0].GoType())
	assert.Equal(t, "error", parameterTypes[1].NameFull())
}

func TestFunctionType_Call(t *testing.T) {
	typ := goo.TypeOf(testFunction)
	assert.True(t, typ.IsFunc())
	functionType := typ.ToFuncType()

	args := make([]interface{}, 0)
	args = append(args, "test")
	args = append(args, 1)
	args = append(args, nil)
	args = append(args, nil)

	outputs := functionType.Call(args)
	assert.Len(t, outputs, 2)

	assert.Equal(t, "test", outputs[0])
	assert.Equal(t, "test error", outputs[1].(error).Error())
}

func TestFunctionType_Call_WithMissingParameters(t *testing.T) {
	typ := goo.TypeOf(testFunction)
	assert.True(t, typ.IsFunc())
	functionType := typ.ToFuncType()

	args := make([]interface{}, 0)
	args = append(args, "test")
	args = append(args, 1)

	assert.Panics(t, func() {
		functionType.Call(args)
	})
}

type testInterface interface {
	testMethod(name string, i int, val bool) (string, error)
	testMethod2()
	TestMethod3()
}

func TestInterfaceType_GetMethodCount(t *testing.T) {
	typ := goo.TypeOf((*testInterface)(nil))
	assert.True(t, typ.IsInterface())

	interfaceType := typ.ToInterfaceType()
	assert.Equal(t, 3, interfaceType.MethodNum())
}

func TestInterfaceType_GetMethods(t *testing.T) {
	typ := goo.TypeOf((*testInterface)(nil))
	assert.True(t, typ.IsInterface())

	interfaceType := typ.ToInterfaceType()
	assert.Len(t, interfaceType.Methods(), 3)
}

func TestMapType_NewInstance(t *testing.T) {
	m := make(map[string]bool)
	typ := goo.TypeOf(m)
	assert.True(t, typ.IsMap())

	mapType := typ.ToMapType()
	newMapInstance := mapType.New().(map[string]bool)
	assert.NotNil(t, newMapInstance)
}

func TestMapType_GetKeyType(t *testing.T) {
	m := make(map[string]bool)
	typ := goo.TypeOf(m)
	assert.True(t, typ.IsMap())

	mapType := typ.ToMapType()
	assert.Equal(t, "string", mapType.KeyType().NameFull())
}

func TestMapType_GetValueType(t *testing.T) {
	m := make(map[string]bool)
	typ := goo.TypeOf(m)
	assert.True(t, typ.IsMap())

	mapType := typ.ToMapType()
	assert.Equal(t, "bool", mapType.ValueType().NameFull())
}

func TestMemberMethod_GetName(t *testing.T) {
	typ := goo.TypeOf((*testInterface)(nil))
	assert.True(t, typ.IsInterface())

	interfaceType := typ.ToInterfaceType()
	assert.Len(t, interfaceType.Methods(), 3)

	assert.Equal(t, "TestMethod3", interfaceType.Methods()[0].Name())
	assert.Equal(t, "testMethod", interfaceType.Methods()[1].Name())
	assert.Equal(t, "testMethod2", interfaceType.Methods()[2].Name())
}

func TestMemberMethod_GetMethodParameterCount(t *testing.T) {
	typ := goo.TypeOf((*testInterface)(nil))
	assert.True(t, typ.IsInterface())

	interfaceType := typ.ToInterfaceType()
	assert.Len(t, interfaceType.Methods(), 3)

	assert.Equal(t, 0, interfaceType.Methods()[0].InNum())
	assert.Equal(t, 3, interfaceType.Methods()[1].InNum())
	assert.Equal(t, 0, interfaceType.Methods()[2].InNum())
}

func TestMemberMethod_GetMethodParameterTypes(t *testing.T) {
	typ := goo.TypeOf((*testInterface)(nil))
	assert.True(t, typ.IsInterface())

	interfaceType := typ.ToInterfaceType()
	assert.Len(t, interfaceType.Methods(), 3)

	assert.Equal(t, 0, len(interfaceType.Methods()[0].InTypes()))
	assert.Equal(t, 3, len(interfaceType.Methods()[1].InTypes()))
	assert.Equal(t, 0, len(interfaceType.Methods()[2].InTypes()))

	types := interfaceType.Methods()[1].InTypes()
	assert.Equal(t, "string", types[0].NameFull())
	assert.Equal(t, "int", types[1].NameFull())
	assert.Equal(t, "bool", types[2].NameFull())
}

func TestMemberMethod_GetMethodReturnTypeCount(t *testing.T) {
	typ := goo.TypeOf((*testInterface)(nil))
	assert.True(t, typ.IsInterface())

	interfaceType := typ.ToInterfaceType()
	assert.Len(t, interfaceType.Methods(), 3)

	assert.Equal(t, 0, interfaceType.Methods()[0].OutNum())
	assert.Equal(t, 2, interfaceType.Methods()[1].OutNum())
	assert.Equal(t, 0, interfaceType.Methods()[2].OutNum())
}

func TestMemberMethod_GetMethodReturnTypes(t *testing.T) {
	typ := goo.TypeOf((*testInterface)(nil))
	assert.True(t, typ.IsInterface())

	interfaceType := typ.ToInterfaceType()
	assert.Len(t, interfaceType.Methods(), 3)

	assert.Equal(t, 0, len(interfaceType.Methods()[0].OutTypes()))
	assert.Equal(t, 2, len(interfaceType.Methods()[1].OutTypes()))
	assert.Equal(t, 0, len(interfaceType.Methods()[2].OutTypes()))

	types := interfaceType.Methods()[1].OutTypes()
	assert.Equal(t, "string", types[0].NameFull())
	assert.Equal(t, "error", types[1].NameFull())
}

func TestMemberMethod_Invoke(t *testing.T) {
	typ := goo.TypeOf(Dog{})
	structType := typ.ToStructType()
	methods := structType.Methods()

	assert.NotPanics(t, func() {
		methods[0].Invoke(Dog{})
		methods[2].Invoke(Dog{}, nil, nil)
		methods[2].Invoke(Dog{}, "test", nil)
		outputs := methods[3].Invoke(Dog{})
		assert.Len(t, outputs, 1)
	})

	assert.Panics(t, func() {
		methods[0].Invoke(2)
	})

	assert.Panics(t, func() {
		methods[0].Invoke(Dog{}, "arg1", "arg2")
	})

	assert.Panics(t, func() {
		methods[0].Invoke(Product{})
	})
}

func TestMemberMethod_IsExported(t *testing.T) {
	typ := goo.TypeOf((*testInterface)(nil))
	assert.True(t, typ.IsInterface())

	interfaceType := typ.ToInterfaceType()
	assert.Len(t, interfaceType.Methods(), 3)

	assert.True(t, interfaceType.Methods()[0].IsExported())
	assert.False(t, interfaceType.Methods()[1].IsExported())
	assert.False(t, interfaceType.Methods()[2].IsExported())
}

func TestMemberMethod_String(t *testing.T) {
	typ := goo.TypeOf((*testInterface)(nil))
	assert.True(t, typ.IsInterface())

	interfaceType := typ.ToInterfaceType()
	assert.Len(t, interfaceType.Methods(), 3)

	assert.Equal(t, "TestMethod3", interfaceType.Methods()[0].String())
	assert.Equal(t, "testMethod", interfaceType.Methods()[1].String())
	assert.Equal(t, "testMethod2", interfaceType.Methods()[2].String())
}

func TestSignedIntegerType(t *testing.T) {
	intType := goo.TypeOf(8)
	assert.True(t, intType.IsNumber())

	intNumberType := intType.ToNumberType()
	assert.Equal(t, goo.IntType, intNumberType.Type())
	if bits.UintSize == goo.Bit32 {
		assert.Equal(t, goo.Bit32, intNumberType.BitSize())
	} else {
		assert.Equal(t, goo.Bit64, intNumberType.BitSize())
	}

	assert.Panics(t, func() {
		intNumberType.Overflow(&Animal{})
	})

	assert.Panics(t, func() {
		intNumberType.ToString("test")
	})

	int8Type := goo.TypeOf(int8(8))
	assert.True(t, int8Type.IsNumber())

	int8NumberType := int8Type.ToNumberType()
	assert.Equal(t, goo.IntType, int8NumberType.Type())
	assert.Equal(t, goo.Bit8, int8NumberType.BitSize())

	assert.True(t, int8NumberType.Overflow(129))
	assert.True(t, int8NumberType.Overflow(-150))
	assert.Equal(t, "120", int8NumberType.ToString(120))

	assert.Panics(t, func() {
		int8NumberType.ToString("test")
	})

	int16Type := goo.TypeOf(int16(25))
	assert.True(t, int16Type.IsNumber())

	int16NumberType := int16Type.ToNumberType()
	assert.Equal(t, goo.IntType, int16NumberType.Type())
	assert.Equal(t, goo.Bit16, int16NumberType.BitSize())

	assert.True(t, int16NumberType.Overflow(35974))
	assert.True(t, int16NumberType.Overflow(-39755))
	assert.Equal(t, "1575", int16NumberType.ToString(1575))

	assert.Panics(t, func() {
		int16NumberType.ToString("test")
	})

	int32Type := goo.TypeOf(int32(25))
	assert.True(t, int32Type.IsNumber())

	int32NumberType := int32Type.ToNumberType()
	assert.Equal(t, goo.IntType, int32NumberType.Type())
	assert.Equal(t, goo.Bit32, int32NumberType.BitSize())

	assert.True(t, int32NumberType.Overflow(2443252523))
	assert.True(t, int32NumberType.Overflow(-2443252523))
	assert.Equal(t, "244325", int32NumberType.ToString(244325))

	assert.Panics(t, func() {
		int32NumberType.ToString("test")
	})

	int64Type := goo.TypeOf(int64(25))
	assert.True(t, int32Type.IsNumber())

	int64NumberType := int64Type.ToNumberType()
	assert.Equal(t, goo.IntType, int64NumberType.Type())
	assert.Equal(t, goo.Bit64, int64NumberType.BitSize())
	assert.Equal(t, "244325", int64NumberType.ToString(244325))

	assert.Panics(t, func() {
		int64NumberType.ToString("test")
	})
}

func TestSignedIntegerType_NewInstance(t *testing.T) {
	intType := goo.TypeOf(8)
	intNumberType := intType.ToNumberType()
	val := intNumberType.New()
	assert.NotNil(t, val.(*int))

	int8Type := goo.TypeOf(int8(8))
	int8NumberType := int8Type.ToNumberType()
	val = int8NumberType.New()
	assert.NotNil(t, val.(*int8))

	int16Type := goo.TypeOf(int16(25))
	int16NumberType := int16Type.ToNumberType()
	val = int16NumberType.New()
	assert.NotNil(t, val.(*int16))

	int32Type := goo.TypeOf(int32(25))
	int32NumberType := int32Type.ToNumberType()
	val = int32NumberType.New()
	assert.NotNil(t, val.(*int32))

	int64Type := goo.TypeOf(int64(25))
	int64NumberType := int64Type.ToNumberType()
	val = int64NumberType.New()
	assert.NotNil(t, val.(*int64))
}

func TestUnSignedIntegerType(t *testing.T) {
	intType := goo.TypeOf(uint(8))
	assert.True(t, intType.IsNumber())

	intNumberType := intType.ToNumberType()
	assert.Equal(t, goo.IntType, intNumberType.Type())
	if bits.UintSize == goo.Bit32 {
		assert.Equal(t, goo.Bit32, intNumberType.BitSize())
	} else {
		assert.Equal(t, goo.Bit64, intNumberType.BitSize())
	}

	assert.Panics(t, func() {
		intNumberType.Overflow(&Animal{})
	})

	assert.Panics(t, func() {
		intNumberType.ToString("test")
	})

	int8Type := goo.TypeOf(uint8(8))
	assert.True(t, int8Type.IsNumber())

	int8NumberType := int8Type.ToNumberType()
	assert.Equal(t, goo.IntType, int8NumberType.Type())
	assert.Equal(t, goo.Bit8, int8NumberType.BitSize())

	assert.True(t, int8NumberType.Overflow(uint(280)))
	assert.Equal(t, "120", int8NumberType.ToString(uint(120)))

	assert.Panics(t, func() {
		int8NumberType.ToString("test")
	})

	int16Type := goo.TypeOf(uint16(25))
	assert.True(t, int16Type.IsNumber())

	int16NumberType := int16Type.ToNumberType()
	assert.Equal(t, goo.IntType, int16NumberType.Type())
	assert.Equal(t, goo.Bit16, int16NumberType.BitSize())

	assert.True(t, int16NumberType.Overflow(uint(68954)))
	assert.Equal(t, "1575", int16NumberType.ToString(uint(1575)))

	assert.Panics(t, func() {
		int16NumberType.ToString("test")
	})

	int32Type := goo.TypeOf(uint32(25))
	assert.True(t, int32Type.IsNumber())

	int32NumberType := int32Type.ToNumberType()
	assert.Equal(t, goo.IntType, int32NumberType.Type())
	assert.Equal(t, goo.Bit32, int32NumberType.BitSize())

	assert.True(t, int32NumberType.Overflow(uint(2443252687523)))
	assert.Equal(t, "244325", int32NumberType.ToString(uint(244325)))

	assert.Panics(t, func() {
		int32NumberType.ToString("test")
	})

	int64Type := goo.TypeOf(uint64(25))
	assert.True(t, int32Type.IsNumber())

	int64NumberType := int64Type.ToNumberType()
	assert.Equal(t, goo.IntType, int64NumberType.Type())
	assert.Equal(t, goo.Bit64, int64NumberType.BitSize())
	assert.Equal(t, "244325", int64NumberType.ToString(uint(244325)))

	assert.Panics(t, func() {
		int64NumberType.ToString("test")
	})
}

func TestUnSignedIntegerType_NewInstance(t *testing.T) {
	intType := goo.TypeOf(uint(8))
	intNumberType := intType.ToNumberType()
	val := intNumberType.New()
	assert.NotNil(t, val.(*uint))

	int8Type := goo.TypeOf(uint8(8))
	int8NumberType := int8Type.ToNumberType()
	val = int8NumberType.New()
	assert.NotNil(t, val.(*uint8))

	int16Type := goo.TypeOf(uint16(25))
	int16NumberType := int16Type.ToNumberType()
	val = int16NumberType.New()
	assert.NotNil(t, val.(*uint16))

	int32Type := goo.TypeOf(uint32(25))
	int32NumberType := int32Type.ToNumberType()
	val = int32NumberType.New()
	assert.NotNil(t, val.(*uint32))

	int64Type := goo.TypeOf(uint64(25))
	int64NumberType := int64Type.ToNumberType()
	val = int64NumberType.New()
	assert.NotNil(t, val.(*uint64))
}

func TestComplexType_ToString(t *testing.T) {
	complexNumber := complex(14.3, 22.5)
	typ := goo.TypeOf(complexNumber)
	assert.True(t, typ.IsNumber())

	numberType := typ.ToNumberType()
	assert.Equal(t, goo.ComplexType, numberType.Type())
	assert.Equal(t, "(14.300000+22.500000i)", numberType.ToString(complexNumber))

	assert.Panics(t, func() {
		numberType.ToString(23)
	})
}

func TestComplexType_GetBitSize(t *testing.T) {
	complexNumber := complex(14.3, 22.5)
	typ := goo.TypeOf(complexNumber)
	assert.True(t, typ.IsNumber())

	numberType := typ.ToNumberType()
	assert.Equal(t, goo.ComplexType, numberType.Type())
	assert.Equal(t, goo.Bit128, numberType.BitSize())
}

func TestComplexType_Overflow(t *testing.T) {
	complexNumber := complex(14.3, 22.5)
	typ := goo.TypeOf(complexNumber)
	assert.True(t, typ.IsNumber())

	numberType := typ.ToNumberType()
	assert.Equal(t, goo.ComplexType, numberType.Type())
	assert.Panics(t, func() {
		numberType.Overflow(nil)
	})
}

func TestComplexType_NewInstance(t *testing.T) {
	complexNumber := complex(14.3, 22.5)
	typ := goo.TypeOf(complexNumber)
	assert.True(t, typ.IsNumber())

	numberType := typ.ToNumberType()
	assert.Equal(t, goo.ComplexType, numberType.Type())

	instance := numberType.New()
	assert.NotNil(t, instance)
}

func TestComplexType_GetRealData(t *testing.T) {
	complexNumber := complex(14.3, 22.5)
	typ := goo.TypeOf(complexNumber)
	assert.True(t, typ.IsNumber())

	numberType := typ.ToNumberType()
	assert.Equal(t, goo.ComplexType, numberType.Type())

	complexType := numberType.(goo.Complex)
	assert.Equal(t, 14.3, complexType.GetRealData(complexNumber))

	assert.Panics(t, func() {
		complexType.GetRealData(23)
	})
}

func TestComplexType_GetImaginaryData(t *testing.T) {
	complexNumber := complex(14.3, 22.5)
	typ := goo.TypeOf(complexNumber)
	assert.True(t, typ.IsNumber())

	numberType := typ.ToNumberType()
	assert.Equal(t, goo.ComplexType, numberType.Type())

	complexType := numberType.(goo.Complex)
	assert.Equal(t, 22.5, complexType.GetImaginaryData(complexNumber))

	assert.Panics(t, func() {
		complexType.GetImaginaryData(23)
	})
}

func TestFloatType_GetType(t *testing.T) {
	float32Number := float32(23.2)
	typ := goo.TypeOf(float32Number)
	assert.True(t, typ.IsNumber())

	numberType := typ.ToNumberType()
	assert.Equal(t, goo.FloatType, numberType.Type())

	float64Number := 23.2
	typ = goo.TypeOf(float64Number)
	assert.True(t, typ.IsNumber())

	numberType = typ.ToNumberType()
	assert.Equal(t, goo.FloatType, numberType.Type())
}

func TestFloatType_GetBitSize(t *testing.T) {
	float32Number := float32(23.2)
	typ := goo.TypeOf(float32Number)
	assert.True(t, typ.IsNumber())

	numberType := typ.ToNumberType()
	assert.Equal(t, goo.Bit32, numberType.BitSize())

	float64Number := 23.2
	typ = goo.TypeOf(float64Number)
	assert.True(t, typ.IsNumber())

	numberType = typ.ToNumberType()
	assert.Equal(t, goo.Bit64, numberType.BitSize())
}

func TestFloatType_NewInstance(t *testing.T) {
	float32Number := float32(23.2)
	typ := goo.TypeOf(float32Number)
	assert.True(t, typ.IsNumber())

	numberType := typ.ToNumberType()
	val := numberType.New()
	assert.NotNil(t, val)

	float64Number := 23.2
	typ = goo.TypeOf(float64Number)
	assert.True(t, typ.IsNumber())

	numberType = typ.ToNumberType()
	val = numberType.New()
	assert.NotNil(t, val)
}

func TestFloatType_Overflow(t *testing.T) {
	float32Number := float32(23.2)
	typ := goo.TypeOf(float32Number)
	assert.True(t, typ.IsNumber())

	numberType := typ.ToNumberType()
	assert.Panics(t, func() { numberType.Overflow(&Animal{}) })
}

func TestFloatType_ToString(t *testing.T) {
	floatNumber := 23.2
	typ := goo.TypeOf(floatNumber)
	assert.True(t, typ.IsNumber())

	numberType := typ.ToNumberType()

	assert.Panics(t, func() { numberType.ToString(&Animal{}) })
	assert.Equal(t, "23.200000", numberType.ToString(floatNumber))
}

func TestSliceType_GetElementType(t *testing.T) {
	var slice []string
	typ := goo.TypeOf(slice)
	assert.True(t, typ.IsSlice())

	sliceType := typ.ToSliceType()
	assert.Equal(t, "string", sliceType.GetElementType().NameFull())
}

func TestSliceType_NewInstance(t *testing.T) {
	arr := [5]string{}
	typ := goo.TypeOf(arr[2:3])
	assert.True(t, typ.IsSlice())

	sliceType := typ.ToSliceType()
	instance := sliceType.New().([]string)
	assert.NotNil(t, instance)
}

func TestStringType_NewInstance(t *testing.T) {
	typ := goo.TypeOf("")
	assert.True(t, typ.IsString())

	stringType := typ.ToStringType()
	stringVal := stringType.New()
	assert.NotNil(t, stringVal)
}

func TestStringType_ToFloat32(t *testing.T) {
	typ := goo.TypeOf("")
	assert.True(t, typ.IsString())

	stringType := typ.ToStringType()
	floatVal := stringType.ToFloat32("23.22")
	assert.Equal(t, float32(23.22), floatVal)

	assert.Panics(t, func() { stringType.ToFloat32("") })
}

func TestStringType_ToFloat64(t *testing.T) {
	typ := goo.TypeOf("")
	assert.True(t, typ.IsString())

	stringType := typ.ToStringType()
	floatVal := stringType.ToFloat64("23.22")
	assert.Equal(t, 23.22, floatVal)

	assert.Panics(t, func() { stringType.ToFloat64("") })
}

func TestStringType_ToNumber(t *testing.T) {
	typ := goo.TypeOf("")
	assert.True(t, typ.IsString())

	stringType := typ.ToStringType()

	// float32
	val, err := stringType.ToNumber("23.75", goo.TypeOf(float32(0)).ToNumberType())
	assert.Nil(t, err)
	assert.Equal(t, float32(23.75), val)

	val, err = stringType.ToNumber(fmt.Sprintf("%f", math.MaxFloat64-1.0), goo.TypeOf(float32(0)).ToNumberType())
	assert.NotNil(t, err)

	val, err = stringType.ToNumber("", goo.TypeOf(float32(0)).ToNumberType())
	assert.NotNil(t, err)

	// float64
	val, err = stringType.ToNumber("23.75", goo.TypeOf(float64(0)).ToNumberType())
	assert.Nil(t, err)
	assert.Equal(t, 23.75, val)

	val, err = stringType.ToNumber("", goo.TypeOf(float64(0)).ToNumberType())
	assert.NotNil(t, err)

	// int
	val, err = stringType.ToNumber("23", goo.TypeOf(0).ToNumberType())
	assert.Nil(t, err)
	assert.Equal(t, 23, val)

	val, err = stringType.ToNumber("", goo.TypeOf(0).ToNumberType())
	assert.NotNil(t, err)

	assert.Panics(t, func() {
		stringType.ToNumber("", nil)
	})

	// int8
	val, err = stringType.ToNumber("23", goo.TypeOf(int8(0)).ToNumberType())
	assert.Nil(t, err)
	assert.Equal(t, int8(23), val)

	val, err = stringType.ToNumber("-128", goo.TypeOf(int8(0)).ToNumberType())
	assert.Nil(t, err)
	assert.Equal(t, int8(-128), val)

	val, err = stringType.ToNumber("-150", goo.TypeOf(int8(0)).ToNumberType())
	assert.NotNil(t, err)

	val, err = stringType.ToNumber("", goo.TypeOf(int8(0)).ToNumberType())
	assert.NotNil(t, err)

	// int16
	val, err = stringType.ToNumber("19421", goo.TypeOf(int16(0)).ToNumberType())
	assert.Nil(t, err)
	assert.Equal(t, int16(19421), val)

	val, err = stringType.ToNumber("-15040", goo.TypeOf(int16(0)).ToNumberType())
	assert.Nil(t, err)
	assert.Equal(t, int16(-15040), val)

	val, err = stringType.ToNumber("32980", goo.TypeOf(int16(0)).ToNumberType())
	assert.NotNil(t, err)

	val, err = stringType.ToNumber("", goo.TypeOf(int16(0)).ToNumberType())
	assert.NotNil(t, err)

	// int32
	val, err = stringType.ToNumber("243293245", goo.TypeOf(int32(0)).ToNumberType())
	assert.Nil(t, err)
	assert.Equal(t, int32(243293245), val)

	val, err = stringType.ToNumber("-243293245", goo.TypeOf(int32(0)).ToNumberType())
	assert.Nil(t, err)
	assert.Equal(t, int32(-243293245), val)

	val, err = stringType.ToNumber("23243293245", goo.TypeOf(int32(0)).ToNumberType())
	assert.NotNil(t, err)

	val, err = stringType.ToNumber("", goo.TypeOf(int64(0)).ToNumberType())
	assert.NotNil(t, err)

	// int64
	val, err = stringType.ToNumber("23243293245", goo.TypeOf(int64(0)).ToNumberType())
	assert.Nil(t, err)
	assert.Equal(t, int64(23243293245), val)

	val, err = stringType.ToNumber("-23243293245", goo.TypeOf(int64(0)).ToNumberType())
	assert.Nil(t, err)
	assert.Equal(t, int64(-23243293245), val)

	val, err = stringType.ToNumber("23545243293245741354", goo.TypeOf(int64(0)).ToNumberType())
	assert.NotNil(t, err)

	val, err = stringType.ToNumber("", goo.TypeOf(int64(0)).ToNumberType())
	assert.NotNil(t, err)

	// unit8
	val, err = stringType.ToNumber("23", goo.TypeOf(uint8(0)).ToNumberType())
	assert.Nil(t, err)
	assert.Equal(t, uint8(23), val)

	val, err = stringType.ToNumber("-150", goo.TypeOf(uint8(0)).ToNumberType())
	assert.NotNil(t, err)

	val, err = stringType.ToNumber("258", goo.TypeOf(uint8(0)).ToNumberType())
	assert.NotNil(t, err)

	val, err = stringType.ToNumber("", goo.TypeOf(uint8(0)).ToNumberType())
	assert.NotNil(t, err)

	// uint16
	val, err = stringType.ToNumber("19874", goo.TypeOf(uint16(0)).ToNumberType())
	assert.Nil(t, err)
	assert.Equal(t, uint16(19874), val)

	val, err = stringType.ToNumber("-150", goo.TypeOf(uint16(0)).ToNumberType())
	assert.NotNil(t, err)

	val, err = stringType.ToNumber("68419", goo.TypeOf(uint16(0)).ToNumberType())
	assert.NotNil(t, err)

	val, err = stringType.ToNumber("", goo.TypeOf(uint16(0)).ToNumberType())
	assert.NotNil(t, err)

	// uint32
	val, err = stringType.ToNumber("68941", goo.TypeOf(uint32(0)).ToNumberType())
	assert.Nil(t, err)
	assert.Equal(t, uint32(68941), val)

	val, err = stringType.ToNumber("-150", goo.TypeOf(uint32(0)).ToNumberType())
	assert.NotNil(t, err)

	val, err = stringType.ToNumber("254684571411", goo.TypeOf(uint32(0)).ToNumberType())
	assert.NotNil(t, err)

	val, err = stringType.ToNumber("", goo.TypeOf(uint32(0)).ToNumberType())
	assert.NotNil(t, err)

	// uint64
	val, err = stringType.ToNumber("254684571411", goo.TypeOf(uint64(0)).ToNumberType())
	assert.Nil(t, err)
	assert.Equal(t, uint64(254684571411), val)

	val, err = stringType.ToNumber("-150", goo.TypeOf(uint64(0)).ToNumberType())
	assert.NotNil(t, err)

	val, err = stringType.ToNumber("254684571411656202321", goo.TypeOf(uint64(0)).ToNumberType())
	assert.NotNil(t, err)

	val, err = stringType.ToNumber("", goo.TypeOf(uint64(0)).ToNumberType())
	assert.NotNil(t, err)

	val, err = stringType.ToNumber("", goo.TypeOf(complex(1, 2)).ToNumberType())
	assert.NotNil(t, err)
}

func TestStringType_ToInt(t *testing.T) {
	typ := goo.TypeOf("")
	assert.True(t, typ.IsString())

	stringType := typ.ToStringType()
	result := stringType.ToInt("23")

	assert.Equal(t, 23, result)

	assert.Panics(t, func() {
		stringType.ToInt("")
	})
}

func TestStringType_ToInt8(t *testing.T) {
	typ := goo.TypeOf("")
	assert.True(t, typ.IsString())

	stringType := typ.ToStringType()

	result := stringType.ToInt8("23")
	assert.Equal(t, int8(23), result)

	result = stringType.ToInt8("-128")
	assert.Equal(t, int8(-128), result)

	assert.Panics(t, func() {
		result = stringType.ToInt8("150")
	})

	assert.Panics(t, func() {
		result = stringType.ToInt8("-130")
	})

	assert.Panics(t, func() {
		stringType.ToInt8("")
	})
}

func TestStringType_ToInt16(t *testing.T) {
	typ := goo.TypeOf("")
	assert.True(t, typ.IsString())

	stringType := typ.ToStringType()

	result := stringType.ToInt16("19421")
	assert.Equal(t, int16(19421), result)

	result = stringType.ToInt16("-15040")
	assert.Equal(t, int16(-15040), result)

	assert.Panics(t, func() { result = stringType.ToInt16("32980") })
	assert.Panics(t, func() { result = stringType.ToInt16("-35874") })

	assert.Panics(t, func() { stringType.ToInt16("") })
}

func TestStringType_ToInt32(t *testing.T) {
	typ := goo.TypeOf("")
	assert.True(t, typ.IsString())

	stringType := typ.ToStringType()

	result := stringType.ToInt32("243293245")
	assert.Equal(t, int32(243293245), result)

	result = stringType.ToInt32("-243293245")
	assert.Equal(t, int32(-243293245), result)

	assert.Panics(t, func() {
		result = stringType.ToInt32("23243293245")
	})

	assert.Panics(t, func() {
		result = stringType.ToInt32("-23243293245")
	})

	assert.Panics(t, func() {
		stringType.ToInt32("")
	})
}

func TestStringType_ToInt64(t *testing.T) {
	typ := goo.TypeOf("")
	assert.True(t, typ.IsString())

	stringType := typ.ToStringType()

	result := stringType.ToInt64("23243293245")
	assert.Equal(t, int64(23243293245), result)

	result = stringType.ToInt64("-23243293245")
	assert.Equal(t, int64(-23243293245), result)

	assert.Panics(t, func() {
		result = stringType.ToInt64("23545243293245741354")
	})

	assert.Panics(t, func() {
		result = stringType.ToInt64("-23545243293245741354")
	})

	assert.Panics(t, func() {
		stringType.ToInt64("")
	})
}

func TestStringType_ToUint(t *testing.T) {
	typ := goo.TypeOf("")
	assert.True(t, typ.IsString())

	stringType := typ.ToStringType()

	result := stringType.ToUint("68941")
	assert.Equal(t, uint(68941), result)

	assert.Panics(t, func() {
		result = stringType.ToUint("-150")
	})
}

func TestStringType_ToUint8(t *testing.T) {
	typ := goo.TypeOf("")
	assert.True(t, typ.IsString())

	stringType := typ.ToStringType()

	result := stringType.ToUint8("23")
	assert.Equal(t, uint8(23), result)

	assert.Panics(t, func() {
		result = stringType.ToUint8("-150")
	})

	assert.Panics(t, func() {
		result = stringType.ToUint8("258")
	})

	assert.Panics(t, func() {
		stringType.ToUint16("")
	})
}

func TestStringType_ToUint16(t *testing.T) {
	typ := goo.TypeOf("")
	assert.True(t, typ.IsString())

	stringType := typ.ToStringType()

	result := stringType.ToUint16("19874")
	assert.Equal(t, uint16(19874), result)

	assert.Panics(t, func() {
		result = stringType.ToUint16("-150")
	})

	assert.Panics(t, func() {
		result = stringType.ToUint16("68419")
	})

	assert.Panics(t, func() {
		stringType.ToUint16("")
	})
}

func TestStringType_ToUint32(t *testing.T) {
	typ := goo.TypeOf("")
	assert.True(t, typ.IsString())

	stringType := typ.ToStringType()

	result := stringType.ToUint32("68941")
	assert.Equal(t, uint32(68941), result)

	assert.Panics(t, func() {
		result = stringType.ToUint32("254684571411")
	})

	assert.Panics(t, func() {
		result = stringType.ToUint32("-150")
	})

	assert.Panics(t, func() {
		stringType.ToUint32("")
	})
}

func TestStringType_ToUint64(t *testing.T) {
	typ := goo.TypeOf("")
	assert.True(t, typ.IsString())

	stringType := typ.ToStringType()

	result := stringType.ToUint64("254684571411")
	assert.Equal(t, uint64(254684571411), result)

	assert.Panics(t, func() {
		result = stringType.ToUint64("254684571411656202321")
	})

	assert.Panics(t, func() {
		result = stringType.ToUint64("-150")
	})

	assert.Panics(t, func() {
		stringType.ToUint64("")
	})
}

func TestStructType_GetNames(t *testing.T) {
	testGetNamesForStruct(t, goo.TypeOf(Animal{}))
	testGetNamesForStruct(t, goo.TypeOf(&Animal{}))
}

func testGetNamesForStruct(t *testing.T, typ goo.Type) {
	assert.Equal(t, "Animal", typ.Name())
	assert.Equal(t, "github.com.bingoohuang.gg.pkg.goo.test.Animal", typ.NameFull())
	assert.Equal(t, "goo_test", typ.PkgName())
	assert.Equal(t, "github.com.bingoohuang.gg.pkg.goo.test", typ.PkgNameFull())
	assert.Equal(t, typ.(goo.Struct), typ.ToStructType())
}

func TestStructType_GetFields(t *testing.T) {
	testGetFieldsForStruct(t, goo.TypeOf(Person{}))
	testGetFieldsForStruct(t, goo.TypeOf(&Person{}))
}

func testGetFieldsForStruct(t *testing.T, typ goo.Type) {
	structType := typ.(goo.Struct)
	// all fields
	fieldCount := structType.FieldNum()
	assert.Equal(t, 5, fieldCount)
	fields := structType.Fields()
	assert.Equal(t, 5, len(fields))

	// exported fields
	fieldCount = structType.FieldExportedNum()
	assert.Equal(t, 3, fieldCount)
	fields = structType.FieldsExported()
	assert.Equal(t, 3, len(fields))

	// unexported fields
	fieldCount = structType.FieldUnexportedNum()
	assert.Equal(t, 2, fieldCount)
	fields = structType.FieldsUnexported()
	assert.Equal(t, 2, len(fields))

	// anonymous fields
	fieldCount = structType.FieldAnonymousNum()
	assert.Equal(t, 1, fieldCount)
	fields = structType.FieldsAnonymous()
	assert.Equal(t, 1, len(fields))
}

func TestStructType_GetStructMethods(t *testing.T) {
	typ := goo.TypeOf(Person{})
	structType := typ.(goo.Struct)
	methodsCount := structType.MethodNum()
	assert.Equal(t, 2, methodsCount)
	methods := structType.Methods()
	assert.Equal(t, 2, len(methods))

	typ = goo.TypeOf(&Person{})
	structType = typ.(goo.Struct)
	methodsCount = structType.MethodNum()
	assert.Equal(t, 3, methodsCount)
	methods = structType.Methods()
	assert.Equal(t, 3, len(methods))
}

func TestStructType_Implements(t *testing.T) {
	x := &Dog{}
	x.Run()
	typ := goo.TypeOf(Dog{})
	structType := typ.(goo.Struct)
	assert.Equal(t, false, structType.Implements(goo.TypeOf((*Bark)(nil)).(goo.Interface)))
	assert.Equal(t, true, structType.Implements(goo.TypeOf((*Run)(nil)).(goo.Interface)))

	typ = goo.TypeOf(&Dog{})
	structType = typ.(goo.Struct)
	assert.Equal(t, true, structType.Implements(goo.TypeOf((*Bark)(nil)).(goo.Interface)))
	assert.Equal(t, true, structType.Implements(goo.TypeOf((*Run)(nil)).(goo.Interface)))
}

func TestStructType_EmbeddedStruct(t *testing.T) {
	typ := goo.TypeOf(Dog{})
	assert.True(t, typ.ToStructType().IsEmbedded(goo.TypeOf(Animal{}).ToStructType()))
	assert.True(t, typ.ToStructType().IsEmbedded(goo.TypeOf(Cell{}).ToStructType()))
	assert.False(t, typ.ToStructType().IsEmbedded(goo.TypeOf(Dog{}).ToStructType()))
	assert.Panics(t, func() {
		typ.ToStructType().IsEmbedded(nil)
	})
}

func TestStructType_NewInstance(t *testing.T) {
	typ := goo.TypeOf(Dog{})
	instance := typ.ToStructType().New()
	assert.NotNil(t, instance)
}

func TestTag_String(t *testing.T) {
	tag := &goo.Tag{
		Name:  "test-tag",
		Value: "test-value",
	}
	assert.Equal(t, tag.Name+"->"+tag.Value, tag.String())
}

type Run interface {
	Run()
}

type Bark interface {
	Bark()
}

type Cell struct {
}

type Animal struct {
	Name string
	Cell
}

func (animal Animal) SayHi() string {
	return "Hi, I'm " + animal.Name
}

type Dog struct {
	Animal
}

func (dog *Dog) Bark() {
	log.Print("Bark")
}

func (dog Dog) Run() {
	log.Print("Run")
}

func (dog Dog) Test(arg string, i *int) {

}

func (dog Dog) TestOutputParam() string {
	return "TestOutputParam"
}

type Person struct {
	name    string
	Surname string
	age     int
	Address Address
	Cell
}

func (person Person) GetName() string {
	return person.name
}

func (person Person) GetSurname() string {
	return person.Surname
}

func (person Person) getAge() int {
	return person.age
}

func (person *Person) GetAddress() Address {
	return person.Address
}

type Address struct {
	city    string
	country string
}

func (address Address) GetCity() string {
	return address.city
}

func (address Address) GetCountry() string {
	return address.country
}

func TestType_IsMethods(t *testing.T) {
	typ := goo.TypeOf(Animal{})
	assert.Equal(t, true, typ.IsStruct())
	assert.Equal(t, false, typ.IsInterface())
	assert.Equal(t, false, typ.IsFunc())
	assert.Equal(t, false, typ.IsNumber())
	assert.Equal(t, true, typ.IsInstantiable())
	assert.Equal(t, false, typ.IsMap())
	assert.Equal(t, false, typ.IsPtr())
	assert.Equal(t, false, typ.IsArray())
	assert.Equal(t, false, typ.IsString())
	assert.Equal(t, false, typ.IsBool())
	assert.Equal(t, false, typ.IsSlice())

	typ = goo.TypeOf(&Animal{})
	assert.Equal(t, true, typ.IsStruct())
	assert.Equal(t, false, typ.IsInterface())
	assert.Equal(t, false, typ.IsFunc())
	assert.Equal(t, false, typ.IsNumber())
	assert.Equal(t, true, typ.IsInstantiable())
	assert.Equal(t, false, typ.IsMap())
	assert.Equal(t, true, typ.IsPtr())
	assert.Equal(t, false, typ.IsArray())
	assert.Equal(t, false, typ.IsString())
	assert.Equal(t, false, typ.IsBool())
	assert.Equal(t, false, typ.IsSlice())
}

func TestType_Instantiable(t *testing.T) {
	typ := goo.TypeOf(Animal{})
	assert.True(t, typ.IsInstantiable())

	typ = goo.TypeOf((*testInterface)(nil))
	assert.False(t, typ.IsInstantiable())

	typ = goo.TypeOf(testFunction)
	assert.False(t, typ.IsInstantiable())
}

func TestBaseType_Equals(t *testing.T) {
	typ := goo.TypeOf(Animal{})
	assert.True(t, typ.Equals(goo.TypeOf(Animal{})))
	assert.False(t, typ.Equals(goo.TypeOf(Dog{})))
	assert.False(t, typ.Equals(nil))
}

func TestType_GetTypeFromGoTypeWithNil(t *testing.T) {
	assert.Panics(t, func() { goo.FromGoType(nil) })
}

func Test_isComplex(t *testing.T) {
	assert.True(t, goo.IsComplex(reflect.TypeOf(complex(14.3, 22.5))))
	assert.False(t, goo.IsComplex(reflect.TypeOf(23)))
}

func Test_getGoPointerTypeAndValueWithNil(t *testing.T) {
	assert.Panics(t, func() { goo.GoPtrTypeAndValue(nil) })
}

func Test_getGoTypeAndValueWithNil(t *testing.T) {
	assert.Panics(t, func() { goo.GoTypeAndValue(nil) })
}
