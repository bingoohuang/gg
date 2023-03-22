package test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/stretchr/testify/require"
)

func Test_marshal_indent(t *testing.T) {
	should := require.New(t)
	obj := struct {
		F1 int
		F2 []int
	}{1, []int{2, 3, 4}}
	output, err := json.MarshalIndent(obj, "", "  ")
	should.Nil(err)
	should.Equal("{\n  \"F1\": 1,\n  \"F2\": [\n    2,\n    3,\n    4\n  ]\n}", string(output))
	output, err = jsoni.MarshalIndent(obj, "", "  ")
	should.Nil(err)
	should.Equal("{\n  \"F1\": 1,\n  \"F2\": [\n    2,\n    3,\n    4\n  ]\n}", string(output))
}

func Test_marshal_indent_map(t *testing.T) {
	should := require.New(t)
	obj := map[int]int{1: 2}
	output, err := json.MarshalIndent(obj, "", "  ")
	should.Nil(err)
	should.Equal("{\n  \"1\": 2\n}", string(output))
	output, err = jsoni.MarshalIndent(obj, "", "  ")
	should.Nil(err)
	should.Equal("{\n  \"1\": 2\n}", string(output))
	output, err = jsoni.ConfigCompatibleWithStandardLibrary.MarshalIndent(context.Background(), obj, "", "  ")
	should.Nil(err)
	should.Equal("{\n  \"1\": 2\n}", string(output))
}
