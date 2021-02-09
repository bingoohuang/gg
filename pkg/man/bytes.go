package man

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

// These utilities convert a size in bytes to a human-readable string in either SI (decimal) or IEC (binary) format.
// https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
// |     Input     | ByteCountSI | ByteCountIEC |
// |---------------|-------------|--------------|
// |      999      |   "999B"   |   "999B"    |
// |     1000      |  "1kB"   |   "1000B"   |
// |     1023      |  "1kB"   |   "1023B"   |
// |     1024      |  "1kB"   |  "1KiB"   |
// |  987,654,321  | "987.7MB"  | "941.9MiB"  |
// | math.MaxInt64 |  "9.2EB"   |  "8EiB"   |

// ByteCount produces a human readable representation of an SI size（copy go version).
func ByteCount(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}

	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	f := strings.TrimSuffix(fmt.Sprintf("%.1f", float64(b)/float64(div)), ".0")
	return fmt.Sprintf("%s%cB", f, "kMGTPE"[exp])
}

// IByteCount produces a human readable representation of an IEC size（copy go version).
func IByteCount(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}

	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	f := strings.TrimSuffix(fmt.Sprintf("%.1f", float64(b)/float64(div)), ".0")
	return fmt.Sprintf("%s%ciB", f, "KMGTPE"[exp])
}

// IEC Sizes.
// kibis of bits
const (
	Byte = 1 << (iota * 10)
	KiByte
	MiByte
	GiByte
	TiByte
	PiByte
	EiByte
)

// SI Sizes.
const (
	IByte = 1
	KByte = IByte * 1000
	MByte = KByte * 1000
	GByte = MByte * 1000
	TByte = GByte * 1000
	PByte = TByte * 1000
	EByte = PByte * 1000
)

var bytesSizeTable = map[string]uint64{
	"b": Byte, "kib": KiByte, "kb": KByte, "mib": MiByte, "mb": MByte, "gib": GiByte, "gb": GByte,
	"tib": TiByte, "tb": TByte, "pib": PiByte, "pb": PByte, "eib": EiByte, "eb": EByte,
	// Without suffix
	"": Byte, "ki": KiByte, "k": KByte, "mi": MiByte, "m": MByte, "gi": GiByte, "g": GByte,
	"ti": TiByte, "t": TByte, "pi": PiByte, "p": PByte, "ei": EiByte, "e": EByte,
}

func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}

func humanateBytes(s uint64, base float64, sizes []string) string {
	if s < 10 {
		return fmt.Sprintf("%dB", s)
	}
	e := math.Floor(logn(float64(s), base))
	suffix := sizes[int(e)]
	val := math.Floor(float64(s)/math.Pow(base, e)*10+0.5) / 10
	f := strings.TrimSuffix(fmt.Sprintf("%.1f", val), ".0")
	return f + suffix
}

// Bytes produces a human readable representation of an SI size.
//
// See also: ParseBytes.
//
// Bytes(82854982) -> 83MB
func Bytes(s uint64) string {
	sizes := []string{"B", "kB", "MB", "GB", "TB", "PB", "EB"}
	return humanateBytes(s, 1000, sizes)
}

// IBytes produces a human readable representation of an IEC size.
//
// See also: ParseBytes.
//
// IBytes(82854982) -> 79MiB
func IBytes(s uint64) string {
	sizes := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
	return humanateBytes(s, 1024, sizes)
}

// ParseBytes parses a string representation of bytes into the number
// of bytes it represents.
//
// See Also: Bytes, IBytes.
//
// ParseBytes("42 MB") -> 42000000, nil
// ParseBytes("42 mib") -> 44040192, nil
func ParseBytes(s string) (uint64, error) {
	lastDigit := 0
	hasComma := false
	for _, r := range s {
		if !(unicode.IsDigit(r) || r == '.' || r == ',') {
			break
		}
		if r == ',' {
			hasComma = true
		}
		lastDigit++
	}

	num := s[:lastDigit]
	if hasComma {
		num = strings.Replace(num, ",", "", -1)
	}

	f, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return 0, err
	}

	extra := strings.ToLower(strings.TrimSpace(s[lastDigit:]))
	if m, ok := bytesSizeTable[extra]; ok {
		f *= float64(m)
		if f >= math.MaxUint64 {
			return 0, fmt.Errorf("too large: %v", s)
		}
		return uint64(f), nil
	}

	return 0, fmt.Errorf("unhandled size name: %v", extra)
}
