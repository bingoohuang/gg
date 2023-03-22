package extra

import (
	"context"
	"time"
	"unsafe"

	"github.com/bingoohuang/gg/pkg/jsoni"
)

// RegisterTimeAsInt64Codec encode/decode time since number of unit since epoch. the precision is the unit.
func RegisterTimeAsInt64Codec(precision time.Duration) {
	jsoni.RegisterTypeEncoder("time.Time", &timeAsInt64Codec{precision})
	jsoni.RegisterTypeDecoder("time.Time", &timeAsInt64Codec{precision})
}

type timeAsInt64Codec struct {
	precision time.Duration
}

func (codec *timeAsInt64Codec) Decode(_ context.Context, ptr unsafe.Pointer, iter *jsoni.Iterator) {
	nanoseconds := iter.ReadInt64() * codec.precision.Nanoseconds()
	*((*time.Time)(ptr)) = time.Unix(0, nanoseconds)
}

func (codec *timeAsInt64Codec) IsEmpty(_ context.Context, ptr unsafe.Pointer, _ bool) bool {
	ts := *((*time.Time)(ptr))
	return ts.UnixNano() == 0
}

func (codec *timeAsInt64Codec) Encode(_ context.Context, ptr unsafe.Pointer, stream *jsoni.Stream) {
	ts := *((*time.Time)(ptr))
	stream.WriteInt64(ts.UnixNano() / codec.precision.Nanoseconds())
}
