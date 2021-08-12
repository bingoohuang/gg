package snow

import (
	"encoding/base64"
	"encoding/binary"
	"strconv"
)

// Base64 returns a base64 string of the snowflake ID
func (f ID) Base64() string { return base64.StdEncoding.EncodeToString(f.Bytes()) }

// ParseBase64 converts a base64 string into a snowflake ID
func ParseBase64(id string) (ID, error) {
	b, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return -1, err
	}

	return ParseBytes(b)
}

// Base2 returns a string base2 of the snowflake ID
func (f ID) Base2() string { return strconv.FormatInt(int64(f), 2) }

// ParseBase2 converts a Base2 string into a snowflake ID
func ParseBase2(id string) (ID, error) {
	i, err := strconv.ParseInt(id, 2, 64)
	return ID(i), err
}

// Base36 returns a base36 string of the snowflake ID
func (f ID) Base36() string { return strconv.FormatInt(int64(f), 36) }

// ParseBase36 converts a Base36 string into a snowflake ID
func ParseBase36(id string) (ID, error) {
	i, err := strconv.ParseInt(id, 36, 64)
	return ID(i), err
}

// Int64 returns an int64 of the snowflake ID
func (f ID) Int64() int64 { return int64(f) }

// ParseInt64 converts an int64 into a snowflake ID
func ParseInt64(id int64) ID { return ID(id) }

// String returns a string of the snowflake ID
func (f ID) String() string { return strconv.FormatInt(int64(f), 10) }

// ParseString converts a string into a snowflake ID
func ParseString(id string) (ID, error) {
	i, err := strconv.ParseInt(id, 10, 64)
	return ID(i), err
}

// Bytes returns a byte slice of the snowflake ID
func (f ID) Bytes() []byte { return []byte(f.String()) }

// ParseBytes converts a byte slice into a snowflake ID
func ParseBytes(id []byte) (ID, error) {
	i, err := strconv.ParseInt(string(id), 10, 64)
	return ID(i), err
}

// IntBytes returns an array of bytes of the snowflake ID, encoded as a
// big endian integer.
func (f ID) IntBytes() [8]byte {
	var b [8]byte

	binary.BigEndian.PutUint64(b[:], uint64(f))

	return b
}

// ParseIntBytes converts an array of bytes encoded as big endian integer as
// a snowflake ID
func ParseIntBytes(id [8]byte) ID {
	return ID(binary.BigEndian.Uint64(id[:]))
}
