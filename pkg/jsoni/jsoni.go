package jsoni

import (
	"context"
	"time"
	"unsafe"
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

// CreateTimeEncodeFn creates a time.Time encoding function on the specified layout.
func CreateTimeEncodeFn(layout string) EncoderFunc {
	return func(_ context.Context, p unsafe.Pointer, stream *Stream) {
		stream.WriteString((*((*time.Time)(p))).Format(layout))
	}
}
