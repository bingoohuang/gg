package b64

import (
	"bytes"
	"encoding/base64"
	"io"
	"strings"
)

type EncodeFlags uint8

const (
	Std EncodeFlags = 1 << iota
	URL
	Raw
)

// EncodeBytes encodes src into base64 []byte.
func EncodeBytes(src []byte, flags ...EncodeFlags) ([]byte, error) {
	var b bytes.Buffer
	if _, err := Encode(&b, bytes.NewReader(src), flags...); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// EncodeString encodes src into base64 string.
func EncodeString(src string, flags ...EncodeFlags) (string, error) {
	var b bytes.Buffer
	if _, err := Encode(&b, strings.NewReader(src), flags...); err != nil {
		return "", err
	}
	return b.String(), nil
}

type rawStdEncodingReader struct{ io.Reader }

// StdEncoding：RFC 4648 定义的标准 BASE64 编码字符集，结果填充=，使字节数为4的倍数
// URLEncoding：RFC 4648 定义的另一 BASE64 编码字符集，用 - 和 _ 替换了 + 和 /，用于URL和文件名，结果填充=
// RawStdEncoding：同 StdEncoding，但结果不填充=
// RawURLEncoding：同 URLEncoding，但结果不填充=
func (f *rawStdEncodingReader) Read(p []byte) (int, error) {
	n, err := f.Reader.Read(p)
	for i := 0; i < n; i++ {
		switch p[i] {
		case '-':
			p[i] = '+'
		case '_':
			p[i] = '/'
		case '=':
			n = i
			break
		}
	}

	return n, err
}

// DecodeBytes decode bytes which is in base64 format ( any one of StdEncoding/URLEncoding/RawStdEncoding/RawURLEncoding).
func DecodeBytes(src []byte) ([]byte, error) {
	var b bytes.Buffer
	if _, err := Decode(&b, bytes.NewReader(src)); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// DecodeString decode string which is in base64 format ( any one of StdEncoding/URLEncoding/RawStdEncoding/RawURLEncoding).
func DecodeString(src string) (string, error) {
	var b bytes.Buffer
	if _, err := Decode(&b, strings.NewReader(src)); err != nil {
		return "", err
	}
	return b.String(), nil
}

// Decode copies io.Reader which is in base64 format ( any one of StdEncoding/URLEncoding/RawStdEncoding/RawURLEncoding).
func Decode(dst io.Writer, src io.Reader) (int64, error) {
	return io.Copy(dst, base64.NewDecoder(base64.RawStdEncoding, &rawStdEncodingReader{Reader: src}))
}

// Encode copies io.Reader to io.Writer which is in base64 format ( any one of StdEncoding/URLEncoding/RawStdEncoding/RawURLEncoding).
func Encode(dst io.Writer, src io.Reader, flags ...EncodeFlags) (int64, error) {
	enc := base64.StdEncoding
	var flag EncodeFlags
	for _, f := range flags {
		flag |= f
	}

	if (flag&URL) == URL && (flag&Raw) == Raw {
		enc = base64.RawURLEncoding
	} else if (flag&URL) == Std && (flag&Raw) == Raw {
		enc = base64.RawStdEncoding
	} else if (flag & URL) == URL {
		enc = base64.URLEncoding
	} else {
		enc = base64.StdEncoding
	}

	return io.Copy(base64.NewEncoder(enc, dst), src)
}
