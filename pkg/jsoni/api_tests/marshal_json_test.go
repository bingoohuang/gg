package test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/stretchr/testify/require"
)

type Foo struct {
	Bar interface{}
}

func (f Foo) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(f.Bar)
	return buf.Bytes(), err
}

// Standard Encoder has trailing newline.
func TestEncodeMarshalJSON(t *testing.T) {
	foo := Foo{
		Bar: 123,
	}
	should := require.New(t)
	var buf, stdbuf bytes.Buffer
	enc := jsoni.ConfigCompatibleWithStandardLibrary.NewEncoder(&buf)
	enc.Encode(context.Background(), foo)
	stdenc := json.NewEncoder(&stdbuf)
	stdenc.Encode(foo)
	should.Equal(stdbuf.Bytes(), buf.Bytes())
}
