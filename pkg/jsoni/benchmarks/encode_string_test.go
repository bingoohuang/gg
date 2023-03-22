package test

import (
	"bytes"
	"context"
	"testing"

	"github.com/bingoohuang/gg/pkg/jsoni"
)

func Benchmark_encode_string_with_SetEscapeHTML(b *testing.B) {
	type V struct {
		S string
		B bool
		I int
	}
	json := jsoni.ConfigCompatibleWithStandardLibrary
	b.ReportAllocs()
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(true)
		if err := enc.Encode(ctx, V{S: "s", B: true, I: 233}); err != nil {
			b.Fatal(err)
		}
	}
}
