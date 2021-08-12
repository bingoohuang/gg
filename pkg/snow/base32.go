package snow

import "errors"

const (
	encodeBase32Map = "ybndrfg8ejkmcpqxot1uwisza345h769"
)

// nolint gochecknoglobals
var (
	decodeBase32Map [256]byte

	// ErrInvalidBase32 is returned by ParseBase32 when given an invalid []byte
	ErrInvalidBase32 = errors.New("invalid base32")
)

// Create maps for decoding Base58/Base32.
// This speeds up the process tremendously.
// nolint gochecknoinits
func init() {
	for i := 0; i < len(encodeBase32Map); i++ {
		decodeBase32Map[i] = 0xFF
	}

	for i := 0; i < len(encodeBase32Map); i++ {
		decodeBase32Map[encodeBase32Map[i]] = byte(i)
	}
}

// ParseBase32 parses a base32 []byte into a snowflake ID
// NOTE: There are many different base32 implementations so becareful when
// doing any interoperation.
func ParseBase32(b []byte) (ID, error) {
	var id int64

	for i := range b {
		if decodeBase32Map[b[i]] == 0xFF { // nolint gomnd
			return -1, ErrInvalidBase32
		}

		id = id*32 + int64(decodeBase32Map[b[i]]) // nolint gomnd
	}

	return ID(id), nil
}

// Base32 uses the z-base-32 character set but encodes and decodes similar
// to base58, allowing it to create an even smaller result string.
// NOTE: There are many different base32 implementations so becareful when
// doing any interoperation.
// nolint gomnd
func (f ID) Base32() string {
	if f < 32 {
		return string(encodeBase32Map[f])
	}

	b := make([]byte, 0, 12)
	for f >= 32 {
		b = append(b, encodeBase32Map[f%32])
		f /= 32
	}

	b = append(b, encodeBase32Map[f])

	for x, y := 0, len(b)-1; x < y; x, y = x+1, y-1 { // nolint gomnd
		b[x], b[y] = b[y], b[x]
	}

	return string(b)
}
