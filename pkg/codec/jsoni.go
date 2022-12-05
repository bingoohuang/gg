package codec

import (
	"context"
	"fmt"
	"github.com/bingoohuang/gg/pkg/jsoni"
	"io"
	"os"
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

func ReadJsonFile(obj interface{}, file string) error {
	// 读取配置文件
	configFile, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read config file %s failed: %w", file, err)
	}

	// 反序列化到结构体中
	if err := jsoni.Unmarshal(configFile, obj); err != nil {
		return fmt.Errorf("unmarshal file data failed: %w", err)
	}

	return nil
}

func WriteJsonFile(obj interface{}, indent, file string) error {
	data, err := jsoni.MarshalIndent(obj, "", indent)
	if err != nil {
		return fmt.Errorf("MarshalIndent failed: %w", err)
	}

	if err := os.WriteFile(file, data, os.ModePerm); err != nil {
		return fmt.Errorf("writing config file %s failed: %w", file, err)
	}

	return nil
}
