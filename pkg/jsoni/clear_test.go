package jsoni

import (
	"encoding/json"
	"fmt"
)

func jsonify(data interface{}) string {
	d, _ := json.Marshal(data)
	return string(d)
}

func ExampleSprintf0() {
	type Val struct {
		Val string
	}

	s := jsonify(Val{Val: jsonify(Val{Val: jsonify(Val{Val: "bingoo"})})})
	fmt.Printf("value: %s\n", s)
	fmt.Printf("value: %j\n", AsClearJSON(s))
	// Output:
	// value: {"Val":"{\"Val\":\"{\\\"Val\\\":\\\"bingoo\\\"}\"}"}
	// value: {"Val":{"Val":{"Val":"bingoo"}}}
}
