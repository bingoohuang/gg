package dump_test

import (
	"fmt"
	"sync"

	"github.com/bingoohuang/gg/pkg/dump"
	"github.com/gookit/color"
	"github.com/kortschak/utter"
	"github.com/kr/pretty"
)

func ExampleGotcha2() {
	type cc struct {
		a int
		b [3]byte
	}

	c1 := cc{a: 1, b: [3]byte{1, 2, 3}}
	dump.P(c1)
	// Output:
}

func ExampleGotcha1() {
	type MyStruct struct {
		A int
	}

	var wg sync.WaitGroup

	values := []MyStruct{{1}, {2}, {3}}
	var output []*MyStruct
	for i, v := range values {
		output = append(output, &values[i])
		wg.Add(1)
		go func(m MyStruct) {
			defer wg.Done()
			dump.V(m)
		}(v)
	}

	dump.V(output)
	wg.Wait()

	// Output:
}

func ExampleBasic() {
	dump.P(
		nil, true,
		12, int8(12), int16(12), int32(12), int64(12),
		uint(22), uint8(22), uint16(22), uint32(22), uint64(22),
		float32(23.78), float64(56.45),
		'c', byte('d'),
		"string",
	)

	// Output:
	// <nil>,
	// bool(true),
	// int(12),
	// int8(12),
	// int16(12),
	// int32(12),
	// int64(12),
	// uint(22),
	// uint8(22),
	// uint16(22),
	// uint32(22),
	// uint64(22),
	// float32(23.780000686645508),
	// float64(56.45),
	// int32(99),
	// uint8(100),
	// string("string"), #len=6
	//
}

func ExampleDemo() {
	dump.P(234, int64(56))
	dump.P("abc", "def")
	dump.P([]string{"ab", "cd"})
	dump.P(
		[]interface{}{"ab", 234, []int{1, 3}},
	)

	// Output:
	// PRINT AT github.com/bingoohuang/gg/pkg/dump.ExampleDemo(example_test.go:37)
	//int(234),
	//int64(56),
	//PRINT AT github.com/bingoohuang/gg/pkg/dump.ExampleDemo(example_test.go:38)
	//string("abc"), #len=3
	//string("def"), #len=3
	//PRINT AT github.com/bingoohuang/gg/pkg/dump.ExampleDemo(example_test.go:39)
	//[]string [ #len=2
	//  string("ab"), #len=2
	//  string("cd"), #len=2
	//],
	//PRINT AT github.com/bingoohuang/gg/pkg/dump.ExampleDemo(example_test.go:40)
	//[]interface {} [ #len=3
	//  string("ab"), #len=2
	//  int(234),
	//  []int [ #len=2
	//    int(1),
	//    int(3),
	//  ],
	//],
}

func ExampleSlice() {
	dump.P(
		[]byte("abc"),
		[]int{1, 2, 3},
		[]string{"ab", "cd"},
		[]interface{}{
			"ab",
			234,
			[]int{1, 3},
			[]string{"ab", "cd"},
		},
	)

	// Output:
}

func ExampleStruct() {
	s1 := &struct {
		cannotExport map[string]interface{}
	}{
		cannotExport: map[string]interface{}{
			"key1": 12,
			"key2": "abcd123",
		},
	}

	s2 := struct {
		ab string
		Cd int
	}{
		"ab", 23,
	}

	color.Infoln("- Use fmt.Println:")
	fmt.Println(s1, s2)

	color.Infoln("\n- Use dump.Println:")
	dump.P(
		s1,
		s2,
	)

	// Output:
}

func ExampleMap() {
	dump.P(
		map[string]interface{}{
			"key0": 123,
			"key1": "value1",
			"key2": []int{1, 2, 3},
			"key3": map[string]string{
				"k0": "v0",
				"k1": "v1",
			},
		},
	)
	// Output:
}

func ExampleDemo2() {
	dump.P(
		23,
		[]string{"ab", "cd"},
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		map[string]interface{}{
			"key": "val", "sub": map[string]string{"k": "v"},
		},
		struct {
			ab string
			Cd int
		}{
			"ab", 23,
		},
	)

	// Output:
}

// rum demo:
// 	go run ./refer_kr_pretty.go
// 	go run ./dump/_examples/refer_kr_pretty.go
func ExamleKr() {
	vs := []interface{}{
		23,
		[]string{"ab", "cd"},
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, // len > 10
		map[string]interface{}{
			"key": "val", "sub": map[string]string{"k": "v"},
		},
		struct {
			ab string
			Cd int
		}{
			"ab", 23,
		},
	}

	// print var data
	_, err := pretty.Println(vs...)
	if err != nil {
		panic(err)
	}

	// print var data
	for _, v := range vs {
		utter.Dump(v)
	}
}
