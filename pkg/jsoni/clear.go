package jsoni

import (
	"context"
	"fmt"
)

type ClearJSON struct {
	Val interface{}
}

func AsClearJSON(v interface{}) *ClearJSON {
	return &ClearJSON{Val: v}
}

var clearJSONConfig = Config{
	SortMapKeys: true,
	ClearQuotes: true,
}.Froze()

// Format implements fmt.Formatter
func (val ClearJSON) Format(state fmt.State, verb rune) {
	d := val.Val
	c := context.Background()
	switch verb {
	case 'j':
		switch v := d.(type) {
		case string:
			d = clearJSON(c, []byte(v))
		case []byte:
			d = clearJSON(c, v)
		default:
			d = clearJSONInterface(c, val.Val)
		}
	}

	fmt.Fprint(state, d)
}

func clearMapStringVal(ctx context.Context, m map[string]interface{}) {
	for k, v := range m {
		switch t := v.(type) {
		case string:
			var vm map[string]interface{}
			if err := clearJSONConfig.Unmarshal(ctx, []byte(t), &vm); err == nil {
				clearMapStringVal(ctx, vm)
				m[k] = vm
			}
		case []byte:
			var vm map[string]interface{}
			if err := clearJSONConfig.Unmarshal(ctx, t, &vm); err == nil {
				clearMapStringVal(ctx, vm)
				m[k] = vm
			}
		}
	}
}

func clearJSON(ctx context.Context, data []byte) interface{} {
	var m map[string]interface{}
	if err := clearJSONConfig.Unmarshal(ctx, data, &m); err != nil {
		return data
	}

	clearMapStringVal(ctx, m)

	s, err := clearJSONConfig.MarshalToString(ctx, m)
	if err != nil {
		return data
	}

	return s
}

func clearJSONInterface(ctx context.Context, data interface{}) interface{} {
	s, err := clearJSONConfig.MarshalToString(ctx, data)
	if err != nil {
		return data
	}

	return s
}
