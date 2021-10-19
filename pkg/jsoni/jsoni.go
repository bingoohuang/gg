package jsoni

import (
	"context"
	"github.com/bingoohuang/gg/pkg/jsoni/extra"
	"github.com/bingoohuang/gg/pkg/strcase"
	"io"
)

// MarshalerContext is the interface implemented by types that
// can marshal themselves into valid JSON with context.Context.
type MarshalerContext interface {
	MarshalJSONContext(context.Context) ([]byte, error)
}

// UnmarshalerContext is the interface implemented by types
// that can unmarshal with context.Context a JSON description of themselves.
type UnmarshalerContext interface {
	UnmarshalJSONContext(context.Context, []byte) error
}

type contextKey int

const (
	ContextCfg contextKey = iota
)

// JsoniConfig tries to be 100% compatible with standard library behavior
var JsoniConfig = Config{
	EscapeHTML: true,
}.Froze()

func init() {
	JsoniConfig.RegisterExtension(&extra.NamingStrategyExtension{Translate: strcase.ToCamelLower})
}

func Json(v interface{}) []byte {
	vv, _ := JsoniConfig.Marshal(context.Background(), v)
	return vv
}

func Jsonify(v interface{}) ([]byte, error) {
	return JsoniConfig.Marshal(context.Background(), v)
}

func ParseJson(data []byte, v interface{}) error {
	return JsoniConfig.Unmarshal(context.Background(), data, v)
}

func Encode(w io.Writer, v interface{}) error {
	return JsoniConfig.NewEncoder(w).Encode(context.Background(), v)
}

func Decode(r io.Reader, obj interface{}) error {
	return JsoniConfig.NewDecoder(r).Decode(context.Background(), obj)
}
