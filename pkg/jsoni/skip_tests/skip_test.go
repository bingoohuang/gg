package skip_tests

import (
	"encoding/json"
	"errors"
	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/stretchr/testify/require"
	"io"
	"reflect"
	"testing"
)

type testCase struct {
	ptr    interface{}
	inputs []string
}

var testCases []testCase

func Test_skip(t *testing.T) {
	for _, testCase := range testCases {
		valType := reflect.TypeOf(testCase.ptr).Elem()
		for _, input := range testCase.inputs {
			t.Run(input, func(t *testing.T) {
				should := require.New(t)
				ptrVal := reflect.New(valType)
				stdErr := json.Unmarshal([]byte(input), ptrVal.Interface())
				iter := jsoni.ParseString(jsoni.ConfigDefault, input)
				iter.Skip()
				iter.ReadNil() // trigger looking forward
				err := iter.Error
				if err == io.EOF {
					err = nil
				} else {
					err = errors.New("remaining bytes")
				}
				if stdErr == nil {
					should.Nil(err)
				} else {
					should.NotNil(err)
				}
			})
		}
	}
}
