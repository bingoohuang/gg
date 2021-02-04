package goo_test

import (
	"github.com/bingoohuang/gg/pkg/goo"
)

type MyInterface interface {
	Method1(str string)
	Method2(val int, str string)
}

type MyStruct struct {
	Name  string
	Price float32
}

func MyFunc(string, int, MyStruct) (string, error) {
	return "test", nil
}

func ExampleTypeOf() {
	testFunctionType := goo.TypeOf(MyFunc)
	if testFunctionType.IsFunc() {
		functionType := testFunctionType.ToFuncType()

		args := make([]interface{}, 3)
		args[0] = "test"
		args[1] = 25
		args[2] = MyStruct{}
		outputs := functionType.Call(args)
		if len(outputs) > 0 {
			// ...
		}
	}

	testStructInstance := &MyStruct{}
	testStructType := goo.TypeOf(testStructInstance)

	if testStructType.IsStruct() {
		structType := testStructType.ToStructType()
		for _, method := range structType.Methods() {
			method.Name()
			// method.Invoke(...)
			method.IsExported()
			method.OutTypes()
			method.OutNum()
			method.InTypes()
			method.InNum()
		}

		// ...
		structType.FieldsExported()
		structType.Fields()
		fields := structType.Fields()
		for _, field := range fields {
			field.Name()
			field.Type()
			field.Get(testStructInstance)
			//field.Set(testStructInstance, nil)
			field.Tags()
			tag, err := field.TagByName("json")
			if err == nil {
				if tag.Value != "" && tag.Name != "" {
					// ...
				}
			}
		}
	}

	if testStructType.IsInstantiable() {
		structType := testStructType.ToStructType()
		newStructInstance := structType.New()
		if newStructInstance != nil {
			// ...
		}
	}

	testInterfaceType := goo.TypeOf((*MyInterface)(nil))

	if testInterfaceType.IsInterface() {
		interfaceType := testInterfaceType.ToInterfaceType()
		interfaceType.Methods()
		// ...
		interfaceType.MethodNum()
	}

	signedInt := 25
	testSignedIntType := goo.TypeOf(signedInt)
	if testSignedIntType.IsNumber() {
		numberType := testSignedIntType.ToNumberType()
		if goo.IntType == numberType.Type() {
			integerType := numberType.(goo.Integer)
			if integerType.IsSigned() {
				// ...
			}
		}
	}

	float32Val := float32(42.28)
	testFloat32Type := goo.TypeOf(float32Val)
	if testFloat32Type.IsNumber() {
		numberType := testFloat32Type.ToNumberType()
		if goo.FloatType == numberType.Type() {
			floatType := numberType.(goo.Float)
			if goo.Bit32 == floatType.BitSize() {
				// ...
			}
		}
	}

	testMap := make(map[string]bool, 0)
	testMapType := goo.TypeOf(testMap)
	if testMapType.IsMap() {
		mapType := testMapType.ToMapType()

		keyType := mapType.KeyType()
		if keyType.IsString() {
			// ...
		}

		valueType := mapType.ValueType()
		if valueType.IsBool() {
			// ...
		}
	}

	if testMapType.IsInstantiable() {
		mapType := testMapType.ToMapType()
		newMapInstance := mapType.New()
		if newMapInstance != nil {
			// ...
		}
	}

	str := "test"
	stringTestType := goo.TypeOf(str)
	if stringTestType.IsString() {
		stringType := stringTestType.ToStringType()

		stringType.ToUint8("20")
		// ...
		stringType.ToUint64("58745")

		stringType.ToInt8("-23")
		// ..
		stringType.ToUint64("9823")

		stringType.ToFloat32("23.52")
		// ..
		stringType.ToFloat64("82387.32")
	}

	array := [5]string{}
	testArrayType := goo.TypeOf(array)
	if testArrayType.IsArray() {
		arrayType := testArrayType.ToArrayType()
		arrayType.ElemType()
		// ...
		arrayType.Len()
	}

	testSliceType := goo.TypeOf(array[2:])
	if testSliceType.IsSlice() {
		sliceType := testSliceType.ToSliceType()
		sliceType.GetElementType()
		// ...
		sliceType.New()
	}

	// Output:
}
