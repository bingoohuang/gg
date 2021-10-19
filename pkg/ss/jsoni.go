package ss

import (
	"context"
	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/bingoohuang/gg/pkg/jsoni/extra"
	"github.com/bingoohuang/gg/pkg/strcase"
	"io"
)

// JsoniConfig tries to be 100% compatible with standard library behavior
var JsoniConfig = jsoni.Config{
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

func EncodeJson(w io.Writer, v interface{}) error {
	return JsoniConfig.NewEncoder(w).Encode(context.Background(), v)
}

func DecodeJson(r io.Reader, obj interface{}) error {
	return JsoniConfig.NewDecoder(r).Decode(context.Background(), obj)
}
