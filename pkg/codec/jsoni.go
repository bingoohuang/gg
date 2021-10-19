package codec

import (
	"context"
	"github.com/bingoohuang/gg/pkg/jsoni"
	"io"
)

func Json(v interface{}) []byte {
	vv, _ := jsoni.Marshal(v)
	return vv
}

func Jsonify(v interface{}) ([]byte, error) {
	return jsoni.Marshal(v)
}

func ParseJson(data []byte, v interface{}) error {
	return jsoni.Unmarshal(data, v)
}

func EncodeJson(w io.Writer, v interface{}) error {
	return jsoni.NewEncoder(w).Encode(context.Background(), v)
}

func DecodeJson(r io.Reader, obj interface{}) error {
	return jsoni.NewDecoder(r).Decode(context.Background(), obj)
}
