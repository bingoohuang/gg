package snow

import "errors"

const (
	encodeBase58Map = "123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"
)

// nolint gochecknoglobals
var (
	decodeBase58Map [256]byte

	// ErrInvalidBase58 is returned by ParseBase58 when given an invalid []byte
	ErrInvalidBase58 = errors.New("invalid base58")
)

// Create maps for decoding Base58/Base32.
// This speeds up the process tremendously.
// nolint gochecknoinits
func init() {
	for i := 0; i < len(encodeBase58Map); i++ {
		decodeBase58Map[i] = 0xFF
	}

	for i := 0; i < len(encodeBase58Map); i++ {
		decodeBase58Map[encodeBase58Map[i]] = byte(i)
	}
}

// Base58 returns a base58 string of the snowflake ID
// nolint gomnd
func (f ID) Base58() string {
	if f < 58 {
		return string(encodeBase58Map[f])
	}

	b := make([]byte, 0, 11)
	for f >= 58 {
		b = append(b, encodeBase58Map[f%58])
		f /= 58
	}

	b = append(b, encodeBase58Map[f])

	for x, y := 0, len(b)-1; x < y; x, y = x+1, y-1 {
		b[x], b[y] = b[y], b[x]
	}

	return string(b)
}

// ParseBase58 parses a base58 []byte into a snowflake ID
func ParseBase58(b []byte) (ID, error) {
	var id int64

	for i := range b {
		if decodeBase58Map[b[i]] == 0xFF { // nolint gomnd
			return -1, ErrInvalidBase58
		}

		id = id*58 + int64(decodeBase58Map[b[i]]) // nolint gomnd
	}

	return ID(id), nil
}
