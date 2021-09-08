//go:build go1.8
// +build go1.8

package test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"unicode/utf8"

	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/stretchr/testify/require"
)

func Test_new_encoder(t *testing.T) {
	should := require.New(t)
	buf1 := &bytes.Buffer{}
	encoder1 := json.NewEncoder(buf1)
	encoder1.SetEscapeHTML(false)
	encoder1.Encode([]int{1})
	should.Equal("[1]\n", buf1.String())
	buf2 := &bytes.Buffer{}
	encoder2 := jsoni.NewEncoder(buf2)
	encoder2.SetEscapeHTML(false)
	encoder2.Encode(context.Background(), []int{1})
	should.Equal("[1]\n", buf2.String())
}

func Test_string_encode_with_std_without_html_escape(t *testing.T) {
	api := jsoni.Config{EscapeHTML: false}.Froze()
	should := require.New(t)
	ctx := context.Background()
	for i := 0; i < utf8.RuneSelf; i++ {
		input := string([]byte{byte(i)})
		buf := &bytes.Buffer{}
		encoder := json.NewEncoder(buf)
		encoder.SetEscapeHTML(false)
		err := encoder.Encode(input)
		should.Nil(err)
		stdOutput := buf.String()
		stdOutput = stdOutput[:len(stdOutput)-1]
		jsoniterOutputBytes, err := api.Marshal(ctx, input)
		should.Nil(err)
		jsoniterOutput := string(jsoniterOutputBytes)
		should.Equal(stdOutput, jsoniterOutput)
	}
}
