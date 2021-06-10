package bytex

import "encoding/binary"

func FromUint64(i uint64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], i)
	return buf[:]
}

func ToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}
