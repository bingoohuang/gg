package test

func init() {
	testCases = append(testCases,
		(*bool)(nil),
		(*boolAlias)(nil),
		(*byte)(nil),
		(*byteAlias)(nil),
		(*float32)(nil),
		(*float32Alias)(nil),
		(*float64)(nil),
		(*float64Alias)(nil),
		(*int8)(nil),
		(*int8Alias)(nil),
		(*int16)(nil),
		(*int16Alias)(nil),
		(*int32)(nil),
		(*int32Alias)(nil),
		(*int64)(nil),
		(*int64Alias)(nil),
		(*string)(nil),
		(*stringAlias)(nil),
		(*uint8)(nil),
		(*uint8Alias)(nil),
		(*uint16)(nil),
		(*uint16Alias)(nil),
		(*uint32)(nil),
		(*uint32Alias)(nil),
		(*uintptr)(nil),
		(*uintptrAlias)(nil),
		(*struct {
			A int8Alias    `json:"a"`
			B int16Alias   `json:"stream"`
			C int32Alias   `json:"c"`
			D int64Alias   `json:"d"`
			E uintAlias    `json:"e"`
			F uint16Alias  `json:"f"`
			G uint32Alias  `json:"g"`
			H uint64Alias  `json:"h"`
			I float32Alias `json:"i"`
			J float64Alias `json:"j"`
			K stringAlias  `json:"k"`
			L intAlias     `json:"l"`
			M uintAlias    `json:"m"`
			N boolAlias    `json:"n"`
			O uintptrAlias `json:"o"`
		})(nil),
	)
}

type (
	boolAlias       bool
	byteAlias       byte
	float32Alias    float32
	float64Alias    float64
	ptrFloat64Alias *float64
	int8Alias       int8
	int16Alias      int16
	int32Alias      int32
	ptrInt32Alias   *int32
	int64Alias      int64
	stringAlias     string
	ptrStringAlias  *string
	uint8Alias      uint8
	uint16Alias     uint16
	uint32Alias     uint32
	uintptrAlias    uintptr
	uintAlias       uint
	uint64Alias     uint64
	intAlias        int
)
