package flagparse

import (
	"bytes"
	"fmt"
	"github.com/bingoohuang/gg/pkg/man"
	"github.com/bingoohuang/go-yaml"
	"github.com/bingoohuang/go-yaml/ast"
	"os"
	"reflect"
	"time"
)

var durationType = reflect.TypeOf((*time.Duration)(nil)).Elem()

func decodeDuration(node ast.Node, typ reflect.Type) (reflect.Value, error) {
	if typ != durationType {
		return reflect.Value{}, yaml.ErrContinue
	}

	if s, ok := node.(*ast.StringNode); ok {
		if d, err := time.ParseDuration(s.Value); err != nil {
			return reflect.Value{}, err
		} else {
			return reflect.ValueOf(d), nil
		}
	}
	return reflect.Value{}, yaml.ErrContinue
}

func decodeSize(node ast.Node, typ reflect.Type) (reflect.Value, error) {
	if s, ok := node.(*ast.StringNode); ok {
		if v, err := man.ParseBytes(s.Value); err != nil {
			return reflect.Value{}, err
		} else {
			return yaml.CastUint64(v, typ)
		}
	}
	return reflect.Value{}, yaml.ErrContinue
}

func LoadConfFile(confFile, defaultConfFile string, app interface{}) error {
	if confFile == "" {
		if s, err := os.Stat(defaultConfFile); err != nil || s.IsDir() {
			return nil // not exists
		}
		confFile = defaultConfFile
	}

	data, err := os.ReadFile(confFile)
	if err != nil {
		return fmt.Errorf("read conf file %s error: %q", confFile, err)
	}

	sizeLabel := yaml.LabelDecoder("size", decodeSize)
	durationLabel := yaml.TypeDecoder(durationType, decodeDuration)
	decoder := yaml.NewDecoder(bytes.NewReader(data), sizeLabel, durationLabel)

	if err := decoder.Decode(app); err != nil {
		return fmt.Errorf("decode conf file %s error:%q", confFile, err)
	}

	return nil
}
