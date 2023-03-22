package setstruct_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bingoohuang/gg/pkg/setstruct"
	"github.com/stretchr/testify/assert"
)

type MyParser string

func (r *MyParser) ValueOf(s string) error {
	*r = MyParser(strings.ToLower(s))
	return nil
}

func TestSetStruct(t *testing.T) {
	type StructA struct {
		Name           string
		Age            int
		Female         bool
		Weight         float32
		GopherDuration time.Duration
		Budgets        []string
		Props          map[string]string
		My             MyParser
		pri            string
	}

	a := StructA{}

	setstruct.Set(&a, func(f reflect.StructField) string {
		return map[string]string{
			"Name":           "bingoohuang",
			"Age":            "100",
			"Female":         "yes",
			"Weight":         "63.5",
			"GopherDuration": "23h",
			"Budgets":        "aa,bb,cc",
			"Props":          "city=beijing,floor=15,alone",
			"My":             "HELLO",
			"pri":            "pri",
		}[f.Name]
	})

	assert.Equal(t, StructA{
		Name:           "bingoohuang",
		Age:            100,
		Female:         true,
		Weight:         63.5,
		GopherDuration: 23 * time.Hour,
		Budgets:        []string{"aa", "bb", "cc"},
		Props:          map[string]string{"city": "beijing", "floor": "15", "alone": ""},
		My:             "hello",
	}, a)
}
