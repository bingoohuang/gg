package jsoni

import (
	"encoding/json"
	"github.com/modern-go/reflect2"
	"unsafe"
)

var jsonRawMessageType = PtrElem((*json.RawMessage)(nil))
var jsoniRawMessageType = PtrElem((*RawMessage)(nil))

func createEncoderOfJsonRawMessage(_ *ctx, typ reflect2.Type) ValEncoder {
	switch typ {
	case jsonRawMessageType:
		return &jsonRawMessageCodec{}
	case jsoniRawMessageType:
		return &jsoniRawMessageCodec{}
	default:
		return nil
	}
}

func createDecoderOfJsonRawMessage(_ *ctx, typ reflect2.Type) ValDecoder {
	switch typ {
	case jsonRawMessageType:
		return &jsonRawMessageCodec{}
	case jsoniRawMessageType:
		return &jsoniRawMessageCodec{}
	default:
		return nil
	}
}

type jsonRawMessageCodec struct{}

func (c *jsonRawMessageCodec) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if r := (*json.RawMessage)(ptr); iter.ReadNil() {
		*r = nil
	} else {
		*r = iter.SkipAndReturnBytes()
	}
}

func (c *jsonRawMessageCodec) Encode(ptr unsafe.Pointer, stream *Stream) {
	if r := *((*json.RawMessage)(ptr)); r == nil {
		stream.WriteNil()
	} else {
		stream.WriteRaw(string(r))
	}
}

func (c *jsonRawMessageCodec) IsEmpty(p unsafe.Pointer) bool {
	return len(*((*json.RawMessage)(p))) == 0
}

type jsoniRawMessageCodec struct{}

func (c *jsoniRawMessageCodec) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if r := (*RawMessage)(ptr); iter.ReadNil() {
		*r = nil
	} else {
		*r = iter.SkipAndReturnBytes()
	}
}

func (c *jsoniRawMessageCodec) Encode(ptr unsafe.Pointer, stream *Stream) {
	if r := *((*RawMessage)(ptr)); r == nil {
		stream.WriteNil()
	} else {
		stream.WriteRaw(string(r))
	}
}

func (c *jsoniRawMessageCodec) IsEmpty(p unsafe.Pointer) bool { return len(*((*RawMessage)(p))) == 0 }
