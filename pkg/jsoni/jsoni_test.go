package jsoni_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/bingoohuang/gg/pkg/jsoni/extra"
	"github.com/bingoohuang/gg/pkg/randx"
	"github.com/bingoohuang/gg/pkg/reflector"
	"github.com/bingoohuang/gg/pkg/strcase"
	"github.com/modern-go/reflect2"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type MapStringBytes map[string][]byte

type ContextKey int

const (
	Converting ContextKey = iota
)

func (m *MapStringBytes) UnmarshalJSONContext(ctx context.Context, data []byte) error {
	api := ctx.Value(jsoni.ContextCfg).(jsoni.API)
	if ctx.Value(Converting) == "go" {
		var mm map[string]string
		if err := api.Unmarshal(ctx, data, &mm); err != nil {
			return err
		}

		*m = revertMap(mm)
		return nil
	}

	r := make(map[string][]byte)
	err := api.Unmarshal(ctx, data, &r)
	if err != nil {
		return err
	}
	*m = r
	return nil
}

func (m *MapStringBytes) MarshalJSONContext(ctx context.Context) ([]byte, error) {
	api := ctx.Value(jsoni.ContextCfg).(jsoni.API)
	if ctx.Value(Converting) == "go" {
		return api.Marshal(ctx, convertMap(m))
	}

	return api.Marshal(ctx, (map[string][]byte)(*m))
}

func convertMap(m *MapStringBytes) map[string]string {
	mm := make(map[string]string, len(*m))
	for k, v := range *m {
		mm[k] = string(v)
	}
	return mm
}

func revertMap(mm map[string]string) map[string][]byte {
	r := make(map[string][]byte, len(mm))
	for k, v := range mm {
		r[k] = []byte(v)
	}
	return r
}

func TestContext(t *testing.T) {
	m := make(MapStringBytes)
	m["name"] = []byte("bingoohuang")
	m["addr"] = []byte("中华人民共和国")

	jmc := reflect.TypeOf((*jsoni.MarshalerContext)(nil)).Elem()
	mt := reflect.ValueOf(m).Type()
	assert.False(t, mt.Implements(jmc))
	assert.False(t, mt.ConvertibleTo(jmc))
	pmt := reflect.PtrTo(mt)
	assert.True(t, pmt.Implements(jmc))
	assert.True(t, pmt.ConvertibleTo(jmc))

	api := jsoni.ConfigCompatibleWithStandardLibrary

	bytes, _ := api.Marshal(nil, &m)
	assert.Equal(t, `{"addr":"5Lit5Y2O5Lq65rCR5YWx5ZKM5Zu9","name":"YmluZ29vaHVhbmc="}`, string(bytes))

	ctx := context.WithValue(context.Background(), Converting, "go")
	bytes, _ = api.Marshal(ctx, &m)
	assert.Equal(t, `{"addr":"中华人民共和国","name":"bingoohuang"}`, string(bytes))

	bytes, _ = api.Marshal(nil, m)
	assert.Equal(t, `{"addr":"5Lit5Y2O5Lq65rCR5YWx5ZKM5Zu9","name":"YmluZ29vaHVhbmc="}`, string(bytes))

	bytes, _ = api.Marshal(ctx, m)
	assert.Equal(t, `{"addr":"中华人民共和国","name":"bingoohuang"}`, string(bytes))

	m = make(MapStringBytes)
	assert.Nil(t, jsoni.Unmarshal([]byte(`{"addr":"5Lit5Y2O5Lq65rCR5YWx5ZKM5Zu9","name":"YmluZ29vaHVhbmc="}`), &m))
	expect := MapStringBytes{"name": []byte("bingoohuang"), "addr": []byte("中华人民共和国")}
	assert.Equal(t, expect, m)

	assert.Nil(t, api.Unmarshal(ctx, []byte(`{"addr":"中华人民共和国","name":"bingoohuang"}`), &m))
	assert.Equal(t, expect, m)
}

func TestMarshalJSONArray(t *testing.T) {
	f := struct {
		Foo []json.Number `json:",nilasempty"`
	}{}

	s, _ := jsoni.MarshalToString(f)
	assert.Equal(t, `{"Foo":[]}`, s)
}

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
	c.RegisterExtension(&extra.NamingStrategyExtension{Translate: strcase.ToCamelLower})

	ctx := context.Background()
	c.Unmarshal(ctx, []byte(`{"Foo":"12341"}`), &f)
	assert.Equal(t, int64(12341), f.Foo)
	c.Unmarshal(ctx, []byte(`{"Foo":12342}`), &f)
	assert.Equal(t, int64(12342), f.Foo)
	c.Unmarshal(ctx, []byte(`{"foo":12343}`), &f)
	assert.Equal(t, int64(12343), f.Foo)
	c.Unmarshal(ctx, []byte(`{"foo":"12344"}`), &f)
	assert.Equal(t, int64(12344), f.Foo)

	s, _ := c.MarshalToString(ctx, f)
	assert.Equal(t, `{"foo":"12344"}`, s)
}
func TestUInt64(t *testing.T) {
	f := struct {
		Foo uint64
	}{}

	c := jsoni.Config{EscapeHTML: true, Int64AsString: true}.Froze()
	c.RegisterExtension(&extra.NamingStrategyExtension{Translate: extra.LowerCaseWithUnderscores})
	ctx := context.Background()
	c.Unmarshal(ctx, []byte(`{"Foo":"12341"}`), &f)
	assert.Equal(t, uint64(12341), f.Foo)
	c.Unmarshal(ctx, []byte(`{"Foo":12342}`), &f)
	assert.Equal(t, uint64(12342), f.Foo)
	c.Unmarshal(ctx, []byte(`{"foo":12343}`), &f)
	assert.Equal(t, uint64(12343), f.Foo)
	c.Unmarshal(ctx, []byte(`{"foo":"12344"}`), &f)
	assert.Equal(t, uint64(12344), f.Foo)

	s, _ := c.MarshalToString(ctx, f)
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
