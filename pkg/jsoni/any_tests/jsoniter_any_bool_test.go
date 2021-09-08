package any_tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/stretchr/testify/require"
)

var boolConvertMap = map[string]bool{
	"null":  false,
	"true":  true,
	"false": false,

	`"true"`:  true,
	`"false"`: true,

	"123":   true,
	`"123"`: true,
	"0":     false,
	`"0"`:   false,
	"-1":    true,
	`"-1"`:  true,

	"1.1":       true,
	"0.0":       false,
	"-1.1":      true,
	`""`:        false,
	"[1,2]":     true,
	"[]":        false,
	"{}":        true,
	`{"abc":1}`: true,
}

func Test_read_bool_as_any(t *testing.T) {
	should := require.New(t)

	var any jsoni.Any
	for k, v := range boolConvertMap {
		any = jsoni.Get([]byte(k))
		if v {
			should.True(any.ToBool(), fmt.Sprintf("origin val is %v", k))
		} else {
			should.False(any.ToBool(), fmt.Sprintf("origin val is %v", k))
		}
	}

}

func Test_write_bool_to_stream(t *testing.T) {
	ctx := context.Background()
	should := require.New(t)
	any := jsoni.Get([]byte("true"))
	stream := jsoni.NewStream(jsoni.ConfigDefault, nil, 32)
	any.WriteTo(ctx, stream)
	should.Equal("true", string(stream.Buffer()))
	should.Equal(any.ValueType(), jsoni.BoolValue)

	any = jsoni.Get([]byte("false"))
	stream = jsoni.NewStream(jsoni.ConfigDefault, nil, 32)
	any.WriteTo(ctx, stream)
	should.Equal("false", string(stream.Buffer()))

	should.Equal(any.ValueType(), jsoni.BoolValue)
}
