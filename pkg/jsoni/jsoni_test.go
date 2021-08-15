package jsoni_test

import (
	"encoding/json"
	"fmt"
	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/bingoohuang/gg/pkg/jsoni/extra"
	"github.com/bingoohuang/gg/pkg/randx"
	"github.com/bingoohuang/gg/pkg/reflector"
	"github.com/modern-go/reflect2"
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

type structFieldDecoder struct {
	field        reflect2.StructField
	fieldDecoder jsoni.ValDecoder
}

var hashes []int64
var hashMap = make(map[int64]*structFieldDecoder)
var switchMap = &tenFieldsStructDecoder{}

func init() {
	obj := reflector.New(switchMap)
	for i := 1; i <= 10; i++ {
		hash := randx.Int64()
		hashes = append(hashes, hash)
		d := &structFieldDecoder{}
		hashMap[hash] = d
		obj.Field(fmt.Sprintf("H%d", i)).Set(hash)
		obj.Field(fmt.Sprintf("D%d", i)).Set(d)

	}
}

/*
🕙[2021-08-15 09:51:09.043] ❯ go test -bench=.
goos: darwin
goarch: amd64
pkg: github.com/bingoohuang/gg/pkg/jsoni
cpu: Intel(R) Core(TM) i7-8850H CPU @ 2.60GHz
BenchmarkMapHash-12       623845              1875 ns/op             560 B/op         38 allocs/op
BenchmarkSwitch-12        661324              1760 ns/op             560 B/op         38 allocs/op
*/

func BenchmarkMapHash(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		randx.Shuffle(hashes)
		for _, h := range hashes {
			if _, ok := hashMap[h]; !ok {
				b.Errorf("hash not found")
			}
		}
	}
}

func BenchmarkSwitch(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		randx.Shuffle(hashes)
		for _, h := range hashes {
			if d := switchMap.Switch(h); d == nil {
				b.Errorf("hash not found")
			}
		}
	}
}

type tenFieldsStructDecoder struct {
	H1, H2, H3, H4, H5, H6, H7, H8, H9, H10 int64
	D1, D2, D3, D4, D5, D6, D7, D8, D9, D10 *structFieldDecoder
}

func (d *tenFieldsStructDecoder) Switch(fieldHash int64) *structFieldDecoder {
	switch fieldHash {
	case d.H1:
		return d.D1
	case d.H2:
		return d.D2
	case d.H3:
		return d.D3
	case d.H4:
		return d.D4
	case d.H5:
		return d.D5
	case d.H6:
		return d.D6
	case d.H7:
		return d.D7
	case d.H8:
		return d.D8
	case d.H9:
		return d.D9
	case d.H10:
		return d.D10
	default:
		return nil
	}
}
