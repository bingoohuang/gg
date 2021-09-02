package yaml

import (
	"github.com/bingoohuang/gg/pkg/yaml/ast"
	"golang.org/x/xerrors"
	"io"
	"reflect"
)

// DecodeOption functional option type for Decoder
type DecodeOption func(d *Decoder) error

var ErrContinue = xerrors.New("continue processing")

// DecoderFn is a prototype func for customized decoding of yaml values.
// If err returns ErrContinue, the decoding will continued after the func invoking.
type DecoderFn func(node ast.Node, typ reflect.Type) (reflect.Value, error)

// TypeDecoder registers a decoder associated with a struct field type.
func TypeDecoder(typ reflect.Type, f DecoderFn) DecodeOption {
	return func(d *Decoder) error {
		d.typeDecoderMap[typ] = f
		return nil
	}
}

// LabelDecoder registers a decoder associated with a label defined in the struct tag like 'yaml:"name,label=xyz".
func LabelDecoder(label string, f DecoderFn) DecodeOption {
	return func(d *Decoder) error {
		d.labelDecoderMap[label] = f
		return nil
	}
}

// ReferenceReaders pass to Decoder that reference to anchor defined by passed readers
func ReferenceReaders(readers ...io.Reader) DecodeOption {
	return func(d *Decoder) error {
		d.referenceReaders = append(d.referenceReaders, readers...)
		return nil
	}
}

// ReferenceFiles pass to Decoder that reference to anchor defined by passed files
func ReferenceFiles(files ...string) DecodeOption {
	return func(d *Decoder) error {
		d.referenceFiles = files
		return nil
	}
}

// ReferenceDirs pass to Decoder that reference to anchor defined by files under the passed dirs
func ReferenceDirs(dirs ...string) DecodeOption {
	return func(d *Decoder) error {
		d.referenceDirs = dirs
		return nil
	}
}

// RecursiveDir search yaml file recursively from passed dirs by ReferenceDirs option
func RecursiveDir(isRecursive bool) DecodeOption {
	return func(d *Decoder) error {
		d.isRecursiveDir = isRecursive
		return nil
	}
}

// Validator set StructValidator instance to Decoder
func Validator(v StructValidator) DecodeOption {
	return func(d *Decoder) error {
		d.validator = v
		return nil
	}
}

// Strict enable DisallowUnknownField and DisallowDuplicateKey
func Strict() DecodeOption {
	return func(d *Decoder) error {
		d.disallowUnknownField = true
		d.disallowDuplicateKey = true
		return nil
	}
}

// DisallowUnknownField causes the Decoder to return an error when the destination
// is a struct and the input contains object keys which do not match any
// non-ignored, exported fields in the destination.
func DisallowUnknownField() DecodeOption {
	return func(d *Decoder) error {
		d.disallowUnknownField = true
		return nil
	}
}

// DisallowDuplicateKey causes an error when mapping keys that are duplicates
func DisallowDuplicateKey() DecodeOption {
	return func(d *Decoder) error {
		d.disallowDuplicateKey = true
		return nil
	}
}

// UseOrderedMap can be interpreted as a map,
// and uses MapSlice ( ordered map ) aggressively if there is no type specification
func UseOrderedMap() DecodeOption {
	return func(d *Decoder) error {
		d.useOrderedMap = true
		return nil
	}
}

// UseJSONUnmarshaler if neither `BytesUnmarshaler` nor `InterfaceUnmarshaler` is implemented
// and `UnmashalJSON([]byte)error` is implemented, convert the argument from `YAML` to `JSON` and then call it.
func UseJSONUnmarshaler() DecodeOption {
	return func(d *Decoder) error {
		d.useJSONUnmarshaler = true
		return nil
	}
}

// EncodeOption functional option type for Encoder
type EncodeOption func(e *Encoder) error

// KeyNaming changes the yaml keys' naming when encoding structs.
func KeyNaming(v KeyNamingStrategy) EncodeOption {
	return func(e *Encoder) error {
		e.KeyNaming = v
		return nil
	}
}

// Indent change indent number
func Indent(spaces int) EncodeOption {
	return func(e *Encoder) error {
		e.indent = spaces
		return nil
	}
}

// IndentSequence causes sequence values to be indented the same value as Indent
func IndentSequence(indent bool) EncodeOption {
	return func(e *Encoder) error {
		e.indentSequence = indent
		return nil
	}
}

// Flow encoding by flow style
func Flow(isFlowStyle bool) EncodeOption {
	return func(e *Encoder) error {
		e.isFlowStyle = isFlowStyle
		return nil
	}
}

// UseLiteralStyleIfMultiline causes encoding multiline strings with a literal syntax,
// no matter what characters they include
func UseLiteralStyleIfMultiline(useLiteralStyleIfMultiline bool) EncodeOption {
	return func(e *Encoder) error {
		e.useLiteralStyleIfMultiline = useLiteralStyleIfMultiline
		return nil
	}
}

// JSON encode in JSON format
func JSON() EncodeOption {
	return func(e *Encoder) error {
		e.isJSONStyle = true
		e.isFlowStyle = true
		return nil
	}
}

// MarshalAnchor call back if encoder find an anchor during encoding
func MarshalAnchor(callback func(*ast.AnchorNode, interface{}) error) EncodeOption {
	return func(e *Encoder) error {
		e.anchorCallback = callback
		return nil
	}
}

// UseJSONMarshaler if neither `BytesMarshaler` nor `InterfaceMarshaler`
// nor `encoding.TextMarshaler` is implemented and `MarshalJSON()([]byte, error)` is implemented,
// call `MarshalJSON` to convert the returned `JSON` to `YAML` for processing.
func UseJSONMarshaler() EncodeOption {
	return func(e *Encoder) error {
		e.useJSONMarshaler = true
		return nil
	}
}
