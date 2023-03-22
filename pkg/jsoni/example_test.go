package jsoni

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func ExampleMarshal() {
	type ColorGroup struct {
		ID     int
		Name   string
		Colors []string
	}
	group := ColorGroup{
		ID:     1,
		Name:   "Reds",
		Colors: []string{"Crimson", "Red", "Ruby", "Maroon"},
	}
	b, err := Marshal(group)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(string(b))

	var p *int
	b, err = json.Marshal(p)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(b))

	b, err = Marshal(p)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(b))
	// Output:
	// {"ID":1,"Name":"Reds","Colors":["Crimson","Red","Ruby","Maroon"]}
	// null
	// null
}

func ExampleUnmarshal() {
	jsonBlob := []byte(`[
		{"Name": "Platypus", "Order": "Monotremata"},
		{"Name": "Quoll",    "Order": "Dasyuromorphia"}
	]`)
	type Animal struct {
		Name  string
		Order string
	}
	var animals []Animal
	err := Unmarshal(jsonBlob, &animals)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%+v\n", animals)

	type A struct {
		Bar string `json:"Bar"`
	}

	var a A
	c := Config{EscapeHTML: true, CaseSensitive: true}.Froze()
	c.Unmarshal(context.Background(), []byte(`{"Bar": "1", "bar": "2" }`), &a)
	fmt.Println(a.Bar)

	// Output:
	// [{Name:Platypus Order:Monotremata} {Name:Quoll Order:Dasyuromorphia}]
	// 1
}

func ExampleConfigFastest_Marshal() {
	type ColorGroup struct {
		ID     int
		Name   string
		Colors []string
	}
	group := ColorGroup{
		ID:     1,
		Name:   "Reds",
		Colors: []string{"Crimson", "Red", "Ruby", "Maroon"},
	}
	stream := ConfigFastest.BorrowStream(nil)
	defer ConfigFastest.ReturnStream(stream)
	stream.WriteVal(context.Background(), group)
	if stream.Error != nil {
		fmt.Println("error:", stream.Error)
	}
	os.Stdout.Write(stream.Buffer())
	// Output:
	// {"ID":1,"Name":"Reds","Colors":["Crimson","Red","Ruby","Maroon"]}
}

func ExampleConfigFastest_Unmarshal() {
	jsonBlob := []byte(`[
		{"Name": "Platypus", "Order": "Monotremata"},
		{"Name": "Quoll",    "Order": "Dasyuromorphia"}
	]`)
	type Animal struct {
		Name  string
		Order string
	}
	var animals []Animal
	iter := ConfigFastest.BorrowIterator(jsonBlob)
	defer ConfigFastest.ReturnIterator(iter)
	iter.ReadVal(context.Background(), &animals)
	if iter.Error != nil {
		fmt.Println("error:", iter.Error)
	}
	fmt.Printf("%+v", animals)
	// Output:
	// [{Name:Platypus Order:Monotremata} {Name:Quoll Order:Dasyuromorphia}]
}

func ExampleGet() {
	val := []byte(`{"ID":1,"Name":"Reds","Colors":["Crimson","Red","Ruby","Maroon"]}`)
	fmt.Printf(Get(val, "Colors", 0).ToString())
	// Output:
	// Crimson
}

func ExampleMyKey() {
	hello := MyKey("hello")
	output, _ := Marshal(map[*MyKey]string{&hello: "world"})
	fmt.Println(string(output))
	obj := map[*MyKey]string{}
	Unmarshal(output, &obj)
	for k, v := range obj {
		fmt.Println(*k, v)
	}
	// Output:
	// {"Hello":"world"}
	// Hel world
}

type MyKey string

func (m *MyKey) MarshalText() ([]byte, error) {
	return []byte(strings.Replace(string(*m), "h", "H", -1)), nil
}

func (m *MyKey) UnmarshalText(text []byte) error {
	*m = MyKey(text[:3])
	return nil
}
