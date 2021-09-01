package jsoni

import (
	"fmt"
	"strconv"
)

type stringAny struct {
	baseAny
	val string
}

func (a *stringAny) Get(path ...interface{}) Any {
	if len(path) == 0 {
		return a
	}
	return &invalidAny{baseAny{}, fmt.Errorf("GetIndex %v from simple value", path)}
}

func (a *stringAny) Parse() *Iterator     { return nil }
func (a *stringAny) ValueType() ValueType { return StringValue }
func (a *stringAny) MustBeValid() Any     { return a }
func (a *stringAny) LastError() error     { return nil }
func (a *stringAny) ToBool() bool {
	str := a.ToString()
	if str == "0" {
		return false
	}
	for _, c := range str {
		switch c {
		case ' ', '\n', '\r', '\t':
		default:
			return true
		}
	}
	return false
}

func (a *stringAny) ToInt() int     { return int(a.ToInt64()) }
func (a *stringAny) ToInt32() int32 { return int32(a.ToInt64()) }

func (a *stringAny) ToInt64() int64 {
	if a.val == "" {
		return 0
	}

	flag := 1
	startPos := 0
	if a.val[0] == '+' || a.val[0] == '-' {
		startPos = 1
	}

	if a.val[0] == '-' {
		flag = -1
	}

	endPos := startPos
	for i := startPos; i < len(a.val); i++ {
		if a.val[i] >= '0' && a.val[i] <= '9' {
			endPos = i + 1
		} else {
			break
		}
	}
	parsed, _ := strconv.ParseInt(a.val[startPos:endPos], 10, 64)
	return int64(flag) * parsed
}

func (a *stringAny) ToUint() uint     { return uint(a.ToUint64()) }
func (a *stringAny) ToUint32() uint32 { return uint32(a.ToUint64()) }

func (a *stringAny) ToUint64() uint64 {
	if a.val == "" {
		return 0
	}

	startPos := 0

	if a.val[0] == '-' {
		return 0
	}
	if a.val[0] == '+' {
		startPos = 1
	}

	endPos := startPos
	for i := startPos; i < len(a.val); i++ {
		if a.val[i] >= '0' && a.val[i] <= '9' {
			endPos = i + 1
		} else {
			break
		}
	}
	parsed, _ := strconv.ParseUint(a.val[startPos:endPos], 10, 64)
	return parsed
}

func (a *stringAny) ToFloat32() float32 { return float32(a.ToFloat64()) }

func (a *stringAny) ToFloat64() float64 {
	if len(a.val) == 0 {
		return 0
	}

	// first char invalid
	if a.val[0] != '+' && a.val[0] != '-' && (a.val[0] > '9' || a.val[0] < '0') {
		return 0
	}

	// extract valid num expression from string
	// eg 123true => 123, -12.12xxa => -12.12
	endPos := 1
	for i := 1; i < len(a.val); i++ {
		if a.val[i] == '.' || a.val[i] == 'e' || a.val[i] == 'E' || a.val[i] == '+' || a.val[i] == '-' {
			endPos = i + 1
			continue
		}

		// end position is the first char which is not digit
		if a.val[i] >= '0' && a.val[i] <= '9' {
			endPos = i + 1
		} else {
			endPos = i
			break
		}
	}
	parsed, _ := strconv.ParseFloat(a.val[:endPos], 64)
	return parsed
}

func (a *stringAny) ToString() string          { return a.val }
func (a *stringAny) WriteTo(stream *Stream)    { stream.WriteString(a.val) }
func (a *stringAny) GetInterface() interface{} { return a.val }
