package jsoni

import (
	"bytes"
	"io"
)

// RawMessage to make replace json with jsoniter
type RawMessage []byte

// Unmarshal adapts to json/encoding Unmarshal API
//
// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v.
// Refer to https://godoc.org/encoding/json#Unmarshal for more information
func Unmarshal(data []byte, v interface{}) error { return ConfigDefault.Unmarshal(data, v) }

// UnmarshalFromString is a convenient method to read from string instead of []byte
func UnmarshalFromString(str string, v interface{}) error {
	return ConfigDefault.UnmarshalFromString(str, v)
}

// Get quick method to get value from deeply nested JSON structure
func Get(data []byte, path ...interface{}) Any {
	return ConfigDefault.Get(data, path...)
}

// Marshal adapts to json/encoding Marshal API
//
// Marshal returns the JSON encoding of v, adapts to json/encoding Marshal API
// Refer to https://godoc.org/encoding/json#Marshal for more information
func Marshal(v interface{}) ([]byte, error) {
	return ConfigDefault.Marshal(v)
}

// MarshalIndent same as json.MarshalIndent. Prefix is not supported.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return ConfigDefault.MarshalIndent(v, prefix, indent)
}

// MarshalToString convenient method to write as string instead of []byte
func MarshalToString(v interface{}) (string, error) {
	return ConfigDefault.MarshalToString(v)
}

// NewDecoder adapts to json/stream NewDecoder API.
//
// NewDecoder returns a new decoder that reads from r.
//
// Instead of a json/encoding Decoder, an Decoder is returned
// Refer to https://godoc.org/encoding/json#NewDecoder for more information
func NewDecoder(reader io.Reader) *Decoder {
	return ConfigDefault.NewDecoder(reader)
}

// Decoder reads and decodes JSON values from an input stream.
// Decoder provides identical APIs with json/stream Decoder (Token() and UseNumber() are in progress)
type Decoder struct {
	iter *Iterator
}

// Decode decode JSON into interface{}
func (a *Decoder) Decode(obj interface{}) error {
	if a.iter.head == a.iter.tail && a.iter.reader != nil {
		if !a.iter.loadMore() {
			return io.EOF
		}
	}
	a.iter.ReadVal(obj)
	err := a.iter.Error
	if err == io.EOF {
		return nil
	}
	return a.iter.Error
}

// More is there more?
func (a *Decoder) More() bool {
	iter := a.iter
	if iter.Error != nil {
		return false
	}
	c := iter.nextToken()
	if c == 0 {
		return false
	}
	iter.unreadByte()
	return c != ']' && c != '}'
}

// Buffered remaining buffer
func (a *Decoder) Buffered() io.Reader {
	remaining := a.iter.buf[a.iter.head:a.iter.tail]
	return bytes.NewReader(remaining)
}

// UseNumber causes the Decoder to unmarshal a number into an interface{} as a
// Number instead of as a float64.
func (a *Decoder) UseNumber() {
	cfg := a.iter.cfg.configBeforeFrozen
	cfg.UseNumber = true
	a.iter.cfg = cfg.frozeWithCacheReuse(a.iter.cfg.extraExtensions)
}

// DisallowUnknownFields causes the Decoder to return an error when the destination
// is a struct and the input contains object keys which do not match any
// non-ignored, exported fields in the destination.
func (a *Decoder) DisallowUnknownFields() {
	cfg := a.iter.cfg.configBeforeFrozen
	cfg.DisallowUnknownFields = true
	a.iter.cfg = cfg.frozeWithCacheReuse(a.iter.cfg.extraExtensions)
}

// NewEncoder same as json.NewEncoder
func NewEncoder(writer io.Writer) *Encoder {
	return ConfigDefault.NewEncoder(writer)
}

// Encoder same as json.Encoder
type Encoder struct {
	stream *Stream
}

// Encode encode interface{} as JSON to io.Writer
func (a *Encoder) Encode(val interface{}) error {
	a.stream.WriteVal(val)
	a.stream.WriteRaw("\n")
	a.stream.Flush()
	return a.stream.Error
}

// SetIndent set the indention. Prefix is not supported
func (a *Encoder) SetIndent(prefix, indent string) {
	config := a.stream.cfg.configBeforeFrozen
	config.IndentionStep = len(indent)
	a.stream.cfg = config.frozeWithCacheReuse(a.stream.cfg.extraExtensions)
}

// SetEscapeHTML escape html by default, set to false to disable
func (a *Encoder) SetEscapeHTML(escapeHTML bool) {
	config := a.stream.cfg.configBeforeFrozen
	config.EscapeHTML = escapeHTML
	a.stream.cfg = config.frozeWithCacheReuse(a.stream.cfg.extraExtensions)
}

// Valid reports whether data is a valid JSON encoding.
func Valid(data []byte) bool {
	return ConfigDefault.Valid(data)
}
