package jsoni

import (
	"encoding/json"
	"github.com/modern-go/reflect2"
	"unsafe"
)

var jsonRawMessageType = PtrElem((*json.RawMessage)(nil))
var jsoniterRawMessageType = PtrElem((*RawMessage)(nil))

func createEncoderOfJsonRawMessage(_ *ctx, typ reflect2.Type) ValEncoder {
	if typ == jsonRawMessageType {
		return &jsonRawMessageCodec{}
	}
	if typ == jsoniterRawMessageType {
		return &jsoniterRawMessageCodec{}
	}
	return nil
}

func createDecoderOfJsonRawMessage(_ *ctx, typ reflect2.Type) ValDecoder {
	if typ == jsonRawMessageType {
		return &jsonRawMessageCodec{}
	}
	if typ == jsoniterRawMessageType {
		return &jsoniterRawMessageCodec{}
	}
	return nil
}

type jsonRawMessageCodec struct{}

func (c *jsonRawMessageCodec) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if iter.ReadNil() {
		*((*json.RawMessage)(ptr)) = nil
	} else {
		*((*json.RawMessage)(ptr)) = iter.SkipAndReturnBytes()
	}
}

func (c *jsonRawMessageCodec) Encode(ptr unsafe.Pointer, stream *Stream) {
	if *((*json.RawMessage)(ptr)) == nil {
		stream.WriteNil()
	} else {
		stream.WriteRaw(string(*((*json.RawMessage)(ptr))))
	}
}

func (c *jsonRawMessageCodec) IsEmpty(ptr unsafe.Pointer) bool {
	return len(*((*json.RawMessage)(ptr))) == 0
}

type jsoniterRawMessageCodec struct{}

func (c *jsoniterRawMessageCodec) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if iter.ReadNil() {
		*((*RawMessage)(ptr)) = nil
	} else {
		*((*RawMessage)(ptr)) = iter.SkipAndReturnBytes()
	}
}

func (c *jsoniterRawMessageCodec) Encode(ptr unsafe.Pointer, stream *Stream) {
	if *((*RawMessage)(ptr)) == nil {
		stream.WriteNil()
	} else {
		stream.WriteRaw(string(*((*RawMessage)(ptr))))
	}
}

func (c *jsoniterRawMessageCodec) IsEmpty(ptr unsafe.Pointer) bool {
	return len(*((*RawMessage)(ptr))) == 0
}
