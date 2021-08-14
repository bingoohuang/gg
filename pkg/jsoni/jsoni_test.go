package jsoni_test

import (
	"encoding/json"
	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/bingoohuang/gg/pkg/jsoni/extra"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMarshalJSON(t *testing.T) {
	f := struct {
		Foo json.Number
	}{}

	jsoni.Unmarshal([]byte(`{"Foo":"12345"}`), &f)
	foo, _ := f.Foo.Int64()
	assert.Equal(t, int64(12345), foo)

	s, _ := jsoni.MarshalToString(f)
	assert.Equal(t, `{"Foo":12345}`, s)
}

func TestMarshalJSONTag(t *testing.T) {
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

	c := jsoni.Config{EscapeHTML: true, Int64AsString: true}.Froze()
	c.RegisterExtension(&extra.NamingStrategyExtension{Translate: extra.LowerCaseWithUnderscores})

	c.Unmarshal([]byte(`{"Foo":"12341"}`), &f)
	assert.Equal(t, int64(12341), f.Foo)
	c.Unmarshal([]byte(`{"Foo":12342}`), &f)
	assert.Equal(t, int64(12342), f.Foo)
	c.Unmarshal([]byte(`{"foo":12343}`), &f)
	assert.Equal(t, int64(12343), f.Foo)
	c.Unmarshal([]byte(`{"foo":"12344"}`), &f)
	assert.Equal(t, int64(12344), f.Foo)

	s, _ := c.MarshalToString(f)
	assert.Equal(t, `{"foo":"12344"}`, s)
}
func TestUInt64(t *testing.T) {
	f := struct {
		Foo uint64
	}{}

	c := jsoni.Config{EscapeHTML: true, Int64AsString: true}.Froze()
	c.RegisterExtension(&extra.NamingStrategyExtension{Translate: extra.LowerCaseWithUnderscores})

	c.Unmarshal([]byte(`{"Foo":"12341"}`), &f)
	assert.Equal(t, uint64(12341), f.Foo)
	c.Unmarshal([]byte(`{"Foo":12342}`), &f)
	assert.Equal(t, uint64(12342), f.Foo)
	c.Unmarshal([]byte(`{"foo":12343}`), &f)
	assert.Equal(t, uint64(12343), f.Foo)
	c.Unmarshal([]byte(`{"foo":"12344"}`), &f)
	assert.Equal(t, uint64(12344), f.Foo)

	s, _ := c.MarshalToString(f)
	assert.Equal(t, `{"foo":"12344"}`, s)
}
