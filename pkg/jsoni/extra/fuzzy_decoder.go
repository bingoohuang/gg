package extra

import (
	"context"
	"encoding/json"
	"io"
	"math"
	"reflect"
	"strings"
	"unsafe"

	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/modern-go/reflect2"
)

const maxUint = ^uint(0)
const maxInt = int(maxUint >> 1)
const minInt = -maxInt - 1

// RegisterFuzzyDecoders decode input from PHP with tolerance.
// It will handle string/number auto conversation, and treat empty [] as empty struct.
func RegisterFuzzyDecoders() {
	jsoni.RegisterExtension(&tolerateEmptyArrayExtension{})
	jsoni.RegisterTypeDecoder("string", &fuzzyStringDecoder{})
	jsoni.RegisterTypeDecoder("float32", &fuzzyFloat32Decoder{})
	jsoni.RegisterTypeDecoder("float64", &fuzzyFloat64Decoder{})
	jsoni.RegisterTypeDecoder("int", &fuzzyIntegerDecoder{func(isFloat bool, ptr unsafe.Pointer, iter *jsoni.Iterator) {
		if isFloat {
			val := iter.ReadFloat64()
			if val > float64(maxInt) || val < float64(minInt) {
				iter.ReportError("fuzzy decode int", "exceed range")
				return
			}
			*((*int)(ptr)) = int(val)
		} else {
			*((*int)(ptr)) = iter.ReadInt()
		}
	}})
	jsoni.RegisterTypeDecoder("uint", &fuzzyIntegerDecoder{func(isFloat bool, ptr unsafe.Pointer, iter *jsoni.Iterator) {
		if isFloat {
			val := iter.ReadFloat64()
			if val > float64(maxUint) || val < 0 {
				iter.ReportError("fuzzy decode uint", "exceed range")
				return
			}
			*((*uint)(ptr)) = uint(val)
		} else {
			*((*uint)(ptr)) = iter.ReadUint()
		}
	}})
	jsoni.RegisterTypeDecoder("int8", &fuzzyIntegerDecoder{func(isFloat bool, ptr unsafe.Pointer, iter *jsoni.Iterator) {
		if isFloat {
			val := iter.ReadFloat64()
			if val > float64(math.MaxInt8) || val < float64(math.MinInt8) {
				iter.ReportError("fuzzy decode int8", "exceed range")
				return
			}
			*((*int8)(ptr)) = int8(val)
		} else {
			*((*int8)(ptr)) = iter.ReadInt8()
		}
	}})
	jsoni.RegisterTypeDecoder("uint8", &fuzzyIntegerDecoder{func(isFloat bool, ptr unsafe.Pointer, iter *jsoni.Iterator) {
		if isFloat {
			val := iter.ReadFloat64()
			if val > float64(math.MaxUint8) || val < 0 {
				iter.ReportError("fuzzy decode uint8", "exceed range")
				return
			}
			*((*uint8)(ptr)) = uint8(val)
		} else {
			*((*uint8)(ptr)) = iter.ReadUint8()
		}
	}})
	jsoni.RegisterTypeDecoder("int16", &fuzzyIntegerDecoder{func(isFloat bool, ptr unsafe.Pointer, iter *jsoni.Iterator) {
		if isFloat {
			val := iter.ReadFloat64()
			if val > float64(math.MaxInt16) || val < float64(math.MinInt16) {
				iter.ReportError("fuzzy decode int16", "exceed range")
				return
			}
			*((*int16)(ptr)) = int16(val)
		} else {
			*((*int16)(ptr)) = iter.ReadInt16()
		}
	}})
	jsoni.RegisterTypeDecoder("uint16", &fuzzyIntegerDecoder{func(isFloat bool, ptr unsafe.Pointer, iter *jsoni.Iterator) {
		if isFloat {
			val := iter.ReadFloat64()
			if val > float64(math.MaxUint16) || val < 0 {
				iter.ReportError("fuzzy decode uint16", "exceed range")
				return
			}
			*((*uint16)(ptr)) = uint16(val)
		} else {
			*((*uint16)(ptr)) = iter.ReadUint16()
		}
	}})
	jsoni.RegisterTypeDecoder("int32", &fuzzyIntegerDecoder{func(isFloat bool, ptr unsafe.Pointer, iter *jsoni.Iterator) {
		if isFloat {
			val := iter.ReadFloat64()
			if val > float64(math.MaxInt32) || val < float64(math.MinInt32) {
				iter.ReportError("fuzzy decode int32", "exceed range")
				return
			}
			*((*int32)(ptr)) = int32(val)
		} else {
			*((*int32)(ptr)) = iter.ReadInt32()
		}
	}})
	jsoni.RegisterTypeDecoder("uint32", &fuzzyIntegerDecoder{func(isFloat bool, ptr unsafe.Pointer, iter *jsoni.Iterator) {
		if isFloat {
			val := iter.ReadFloat64()
			if val > float64(math.MaxUint32) || val < 0 {
				iter.ReportError("fuzzy decode uint32", "exceed range")
				return
			}
			*((*uint32)(ptr)) = uint32(val)
		} else {
			*((*uint32)(ptr)) = iter.ReadUint32()
		}
	}})
	jsoni.RegisterTypeDecoder("int64", &fuzzyIntegerDecoder{func(isFloat bool, ptr unsafe.Pointer, iter *jsoni.Iterator) {
		if isFloat {
			val := iter.ReadFloat64()
			if val > float64(math.MaxInt64) || val < float64(math.MinInt64) {
				iter.ReportError("fuzzy decode int64", "exceed range")
				return
			}
			*((*int64)(ptr)) = int64(val)
		} else {
			*((*int64)(ptr)) = iter.ReadInt64()
		}
	}})
	jsoni.RegisterTypeDecoder("uint64", &fuzzyIntegerDecoder{func(isFloat bool, ptr unsafe.Pointer, iter *jsoni.Iterator) {
		if isFloat {
			val := iter.ReadFloat64()
			if val > float64(math.MaxUint64) || val < 0 {
				iter.ReportError("fuzzy decode uint64", "exceed range")
				return
			}
			*((*uint64)(ptr)) = uint64(val)
		} else {
			*((*uint64)(ptr)) = iter.ReadUint64()
		}
	}})
}

type tolerateEmptyArrayExtension struct {
	jsoni.DummyExtension
}

func (extension *tolerateEmptyArrayExtension) DecorateDecoder(typ reflect2.Type, decoder jsoni.ValDecoder) jsoni.ValDecoder {
	if typ.Kind() == reflect.Struct || typ.Kind() == reflect.Map {
		return &tolerateEmptyArrayDecoder{decoder}
	}
	return decoder
}

type tolerateEmptyArrayDecoder struct {
	valDecoder jsoni.ValDecoder
}

func (decoder *tolerateEmptyArrayDecoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *jsoni.Iterator) {
	if iter.WhatIsNext() == jsoni.ArrayValue {
		iter.Skip()
		newIter := iter.Pool().BorrowIterator([]byte("{}"))
		defer iter.Pool().ReturnIterator(newIter)
		decoder.valDecoder.Decode(ctx, ptr, newIter)
	} else {
		decoder.valDecoder.Decode(ctx, ptr, iter)
	}
}

type fuzzyStringDecoder struct {
}

func (decoder *fuzzyStringDecoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *jsoni.Iterator) {
	valueType := iter.WhatIsNext()
	switch valueType {
	case jsoni.NumberValue:
		var number json.Number
		iter.ReadVal(ctx, &number)
		*((*string)(ptr)) = string(number)
	case jsoni.StringValue:
		*((*string)(ptr)) = iter.ReadString()
	case jsoni.NilValue:
		iter.Skip()
		*((*string)(ptr)) = ""
	default:
		iter.ReportError("fuzzyStringDecoder", "not number or string")
	}
}

type fuzzyIntegerDecoder struct {
	fun func(isFloat bool, ptr unsafe.Pointer, iter *jsoni.Iterator)
}

func (decoder *fuzzyIntegerDecoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *jsoni.Iterator) {
	valueType := iter.WhatIsNext()
	var str string
	switch valueType {
	case jsoni.NumberValue:
		var number json.Number
		iter.ReadVal(ctx, &number)
		str = string(number)
	case jsoni.StringValue:
		str = iter.ReadString()
	case jsoni.BoolValue:
		if iter.ReadBool() {
			str = "1"
		} else {
			str = "0"
		}
	case jsoni.NilValue:
		iter.Skip()
		str = "0"
	default:
		iter.ReportError("fuzzyIntegerDecoder", "not number or string")
	}
	if len(str) == 0 {
		str = "0"
	}
	newIter := iter.Pool().BorrowIterator([]byte(str))
	defer iter.Pool().ReturnIterator(newIter)
	isFloat := strings.IndexByte(str, '.') != -1
	decoder.fun(isFloat, ptr, newIter)
	if newIter.Error != nil && newIter.Error != io.EOF {
		iter.Error = newIter.Error
	}
}

type fuzzyFloat32Decoder struct {
}

func (decoder *fuzzyFloat32Decoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *jsoni.Iterator) {
	valueType := iter.WhatIsNext()
	var str string
	switch valueType {
	case jsoni.NumberValue:
		*((*float32)(ptr)) = iter.ReadFloat32()
	case jsoni.StringValue:
		str = iter.ReadString()
		newIter := iter.Pool().BorrowIterator([]byte(str))
		defer iter.Pool().ReturnIterator(newIter)
		*((*float32)(ptr)) = newIter.ReadFloat32()
		if newIter.Error != nil && newIter.Error != io.EOF {
			iter.Error = newIter.Error
		}
	case jsoni.BoolValue:
		// support bool to float32
		if iter.ReadBool() {
			*((*float32)(ptr)) = 1
		} else {
			*((*float32)(ptr)) = 0
		}
	case jsoni.NilValue:
		iter.Skip()
		*((*float32)(ptr)) = 0
	default:
		iter.ReportError("fuzzyFloat32Decoder", "not number or string")
	}
}

type fuzzyFloat64Decoder struct {
}

func (decoder *fuzzyFloat64Decoder) Decode(ctx context.Context, ptr unsafe.Pointer, iter *jsoni.Iterator) {
	valueType := iter.WhatIsNext()
	var str string
	switch valueType {
	case jsoni.NumberValue:
		*((*float64)(ptr)) = iter.ReadFloat64()
	case jsoni.StringValue:
		str = iter.ReadString()
		newIter := iter.Pool().BorrowIterator([]byte(str))
		defer iter.Pool().ReturnIterator(newIter)
		*((*float64)(ptr)) = newIter.ReadFloat64()
		if newIter.Error != nil && newIter.Error != io.EOF {
			iter.Error = newIter.Error
		}
	case jsoni.BoolValue:
		// support bool to float64
		if iter.ReadBool() {
			*((*float64)(ptr)) = 1
		} else {
			*((*float64)(ptr)) = 0
		}
	case jsoni.NilValue:
		iter.Skip()
		*((*float64)(ptr)) = 0
	default:
		iter.ReportError("fuzzyFloat64Decoder", "not number or string")
	}
}
