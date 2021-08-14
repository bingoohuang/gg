package jsoni_test

import (
	"encoding/json"
	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMarshalJSON(t *testing.T) {
	f := struct {
		Foo json.Number `json:"foo"`
	}{}

	jsoni.Unmarshal([]byte(`{"foo":"12345"}`), &f)
	foo, _ := f.Foo.Int64()
	assert.Equal(t, int64(12345), foo)

	s, _ := jsoni.MarshalToString(f)
	assert.Equal(t, `{"foo":12345}`, s)
}

func TestInt64(t *testing.T) {
	f := struct {
		Foo int64
	}{}

	jsoni.Unmarshal([]byte(`{"Foo":12345}`), &f)
	assert.Equal(t, int64(12345), f.Foo)

	s, _ := jsoni.MarshalToString(f)
	assert.Equal(t, `{"Foo":12345}`, s)
}
