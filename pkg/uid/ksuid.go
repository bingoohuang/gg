package uid

import (
	"bytes"
	"crypto/rand"
	"database/sql/driver"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"math/bits"
	mrand "math/rand"
	"sync"
	"time"
)

const (
	// KSUID's epoch starts more recently so that the 32-bit number space gives a
	// significantly higher useful lifetime of around 136 years from March 2017.
	// This number (14e8) was picked to be easy to remember.
	epochStamp int64 = 1400000000

	// Timestamp is a uint32
	timestampBytesNum = 4

	// Payload is 16-bytes
	payloadBytesNum = 16

	// KSUIDs are 20 bytes when binary encoded
	byteLength = timestampBytesNum + payloadBytesNum

	// The length of a KSUID when string (base62) encoded
	stringEncodedLength = 27

	// A string-encoded minimum value for a KSUID
	minStringEncoded = "000000000000000000000000000"

	// A string-encoded maximum value for a KSUID
	maxStringEncoded = "aWgEPTl1tmebfsQzFP4bxwgy80V"
)

// KSUID is 20 bytes:
//  00-03 byte: uint32 BE UTC timestamp with custom epoch
//  04-19 byte: random "payload"
type KSUID [byteLength]byte

var (
	rander     = rand.Reader
	randMutex  = sync.Mutex{}
	randBuffer = [payloadBytesNum]byte{}

	errSize        = fmt.Errorf("valid KSUIDs are %v bytes", byteLength)
	errStrSize     = fmt.Errorf("valid encoded KSUIDs are %v characters", stringEncodedLength)
	errStrValue    = fmt.Errorf("valid encoded KSUIDs are bounded by %s and %s", minStringEncoded, maxStringEncoded)
	errPayloadSize = fmt.Errorf("valid KSUID payloads are %v bytes", payloadBytesNum)

	// Nil represents a completely empty (invalid) KSUID.
	Nil KSUID
	// Max represents the highest value a KSUID can have.
	Max = KSUID{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}
)

// Append appends the string representation of i to b, returning a slice to a
// potentially larger memory area.
func (i KSUID) Append(b []byte) []byte { return fastAppendEncodeBase62(b, i[:]) }

// Time is the timestamp portion of the ID as a Time object.
func (i KSUID) Time() time.Time { return correctedUTCTimestampToTime(i.Timestamp()) }

// Timestamp is the timestamp portion of the ID as a bare integer which is uncorrected
// for KSUID's special epoch.
func (i KSUID) Timestamp() uint32 { return binary.BigEndian.Uint32(i[:timestampBytesNum]) }

// Payload is the 16-byte random payload without the timestamp.
func (i KSUID) Payload() []byte { return i[timestampBytesNum:] }

// String is string-encoded representation that can be passed through Parse().
func (i KSUID) String() string { return string(i.Append(make([]byte, 0, stringEncodedLength))) }

// Bytes raw byte representation of KSUID.
func (i KSUID) Bytes() []byte { /* Safe because this is by-value*/ return i[:] }

// IsNil returns true if this is a "nil" KSUID.
func (i KSUID) IsNil() bool { return i == Nil }

// Get satisfies the flag.Getter interface, making it possible to use KSUIDs as
// part of the command line options of a program.
func (i KSUID) Get() interface{} { return i }

// Set satisfies the flag.Value interface, making it possible to use KSUIDs as
// part of the command line options of a program.
func (i *KSUID) Set(s string) error            { return i.UnmarshalText([]byte(s)) }
func (i KSUID) MarshalText() ([]byte, error)   { return []byte(i.String()), nil }
func (i KSUID) MarshalBinary() ([]byte, error) { return i.Bytes(), nil }

func (i *KSUID) UnmarshalText(b []byte) error {
	id, err := Parse(string(b))
	if err != nil {
		return err
	}
	*i = id
	return nil
}

func (i *KSUID) UnmarshalBinary(b []byte) error {
	id, err := FromBytes(b)
	if err != nil {
		return err
	}
	*i = id
	return nil
}

// Value converts the KSUID into a SQL driver value which can be used to
// directly use the KSUID as parameter to a SQL query.
func (i KSUID) Value() (driver.Value, error) {
	if i.IsNil() {
		return nil, nil
	}
	return i.String(), nil
}

// Scan implements the sql.Scanner interface. It supports converting from
// string, []byte, or nil into a KSUID value. Attempting to convert from
// another type will return an error.
func (i *KSUID) Scan(src interface{}) error {
	switch v := src.(type) {
	case nil:
		return i.scan(nil)
	case []byte:
		return i.scan(v)
	case string:
		return i.scan([]byte(v))
	default:
		return fmt.Errorf("Scan: unable to scan type %T into KSUID", v)
	}
}

func (i *KSUID) scan(b []byte) error {
	switch len(b) {
	case 0:
		*i = Nil
		return nil
	case byteLength:
		return i.UnmarshalBinary(b)
	case stringEncodedLength:
		return i.UnmarshalText(b)
	default:
		return errSize
	}
}

// Parse decodes a string-encoded representation of a KSUID object
func Parse(s string) (KSUID, error) {
	if len(s) != stringEncodedLength {
		return Nil, errStrSize
	}

	src := [stringEncodedLength]byte{}
	dst := [byteLength]byte{}

	copy(src[:], s[:])

	if err := fastDecodeBase62(dst[:], src[:]); err != nil {
		return Nil, errStrValue
	}

	return FromBytes(dst[:])
}

func timeToCorrectedUTCTimestamp(t time.Time) uint32  { return uint32(t.Unix() - epochStamp) }
func correctedUTCTimestampToTime(ts uint32) time.Time { return time.Unix(int64(ts)+epochStamp, 0) }

// New generates a new KSUID. In the strange case that random bytes
// can't be read, it will panic.
func New() KSUID {
	ksuid, err := NewRandom()
	if err != nil {
		panic(fmt.Sprintf("Couldn't generate KSUID, inconceivable! error: %v", err))
	}
	return ksuid
}

// NewRandom generates a new KSUID
func NewRandom() (ksuid KSUID, err error) { return NewRandomWithTime(time.Now()) }

func NewRandomWithTime(t time.Time) (ksuid KSUID, err error) {
	// Go's default random number generators are not safe for concurrent use by
	// multiple goroutines, the use of the rander and randBuffer are explicitly
	// synchronized here.
	randMutex.Lock()

	_, err = io.ReadAtLeast(rander, randBuffer[:], len(randBuffer))
	copy(ksuid[timestampBytesNum:], randBuffer[:])

	randMutex.Unlock()

	if err != nil {
		ksuid = Nil // don't leak random bytes on error
		return
	}

	ts := timeToCorrectedUTCTimestamp(t)
	binary.BigEndian.PutUint32(ksuid[:timestampBytesNum], ts)
	return
}

// FromParts constructs a KSUID from constituent parts.
func FromParts(t time.Time, payload []byte) (KSUID, error) {
	if len(payload) != payloadBytesNum {
		return Nil, errPayloadSize
	}

	var ksuid KSUID

	ts := timeToCorrectedUTCTimestamp(t)
	binary.BigEndian.PutUint32(ksuid[:timestampBytesNum], ts)

	copy(ksuid[timestampBytesNum:], payload)

	return ksuid, nil
}

// FromBytes constructs a KSUID from a 20-byte binary representation
func FromBytes(b []byte) (KSUID, error) {
	var ksuid KSUID

	if len(b) != byteLength {
		return Nil, errSize
	}

	copy(ksuid[:], b)
	return ksuid, nil
}

// SetRand Sets the global source of random bytes for KSUID generation. This
// should probably only be set once globally. While this is technically
// thread-safe as in it won't cause corruption, there's no guarantee
// on ordering.
func SetRand(r io.Reader) {
	if r == nil {
		rander = rand.Reader
		return
	}
	rander = r
}

// Compare implements comparison for KSUID type.
func Compare(a, b KSUID) int {
	return bytes.Compare(a[:], b[:])
}

// Sort sorts the given slice of KSUIDs.
func Sort(ids []KSUID) { quickSort(ids, 0, len(ids)-1) }

// IsSorted checks whether a slice of KSUIDs is sorted
func IsSorted(ids []KSUID) bool {
	if len(ids) != 0 {
		min := ids[0]
		for _, id := range ids[1:] {
			if bytes.Compare(min[:], id[:]) > 0 {
				return false
			}
			min = id
		}
	}
	return true
}

func quickSort(a []KSUID, lo int, hi int) {
	if lo < hi {
		pivot := a[hi]
		i := lo - 1

		for j, n := lo, hi; j != n; j++ {
			if bytes.Compare(a[j][:], pivot[:]) < 0 {
				i++
				a[i], a[j] = a[j], a[i]
			}
		}

		i++
		if bytes.Compare(a[hi][:], a[i][:]) < 0 {
			a[i], a[hi] = a[hi], a[i]
		}

		quickSort(a, lo, i-1)
		quickSort(a, i+1, hi)
	}
}

// Next returns the next KSUID after id.
func (i KSUID) Next() KSUID {
	zero := makeUint128(0, 0)

	t := i.Timestamp()
	u := uint128Payload(i)
	v := add128(u, makeUint128(0, 1))

	if v == zero { // overflow
		t++
	}

	return v.ksuid(t)
}

// Prev returns the previous KSUID before id.
func (i KSUID) Prev() KSUID {
	max := makeUint128(math.MaxUint64, math.MaxUint64)

	t := i.Timestamp()
	u := uint128Payload(i)
	v := sub128(u, makeUint128(0, 1))

	if v == max { // overflow
		t--
	}

	return v.ksuid(t)
}

// uint128 represents an unsigned 128 bits little endian integer.
type uint128 [2]uint64

func uint128Payload(ksuid KSUID) uint128 {
	return makeUint128FromPayload(ksuid[timestampBytesNum:])
}

func makeUint128(high uint64, low uint64) uint128 { return uint128{low, high} }

func makeUint128FromPayload(payload []byte) uint128 {
	return uint128{
		binary.BigEndian.Uint64(payload[8:]), // low
		binary.BigEndian.Uint64(payload[:8]), // high
	}
}

func (v uint128) ksuid(timestamp uint32) (out KSUID) {
	binary.BigEndian.PutUint32(out[:4], timestamp) // time
	binary.BigEndian.PutUint64(out[4:12], v[1])    // high
	binary.BigEndian.PutUint64(out[12:], v[0])     // low
	return
}

func (v uint128) bytes() (out [16]byte) {
	binary.BigEndian.PutUint64(out[:8], v[1])
	binary.BigEndian.PutUint64(out[8:], v[0])
	return
}

func (v uint128) String() string { return fmt.Sprintf("0x%016X%016X", v[0], v[1]) }

func cmp128(x, y uint128) int {
	if x[1] < y[1] {
		return -1
	}
	if x[1] > y[1] {
		return 1
	}
	if x[0] < y[0] {
		return -1
	}
	if x[0] > y[0] {
		return 1
	}
	return 0
}

func add128(x, y uint128) (z uint128) {
	var c uint64
	z[0], c = bits.Add64(x[0], y[0], 0)
	z[1], _ = bits.Add64(x[1], y[1], c)
	return
}

func sub128(x, y uint128) (z uint128) {
	var b uint64
	z[0], b = bits.Sub64(x[0], y[0], 0)
	z[1], _ = bits.Sub64(x[1], y[1], b)
	return
}

func incr128(x uint128) (z uint128) {
	var c uint64
	z[0], c = bits.Add64(x[0], 1, 0)
	z[1] = x[1] + c
	return
}

// Sequence is a KSUID generator which produces a sequence of ordered KSUIDs
// from a seed.
//
// Up to 65536 KSUIDs can be generated by for a single seed.
//
// A typical usage of a Sequence looks like this:
//
//	seq := ksuid.Sequence{
//		Seed: ksuid.New(),
//	}
//	id, err := seq.Next()
//
// Sequence values are not safe to use concurrently from multiple goroutines.
type Sequence struct {
	// The seed is used as base for the KSUID generator, all generated KSUIDs
	// share the same leading 18 bytes of the seed.
	Seed  KSUID
	count uint32 // uint32 for overflow, only 2 bytes are used
}

// Next produces the next KSUID in the sequence, or returns an error if the
// sequence has been exhausted.
func (seq *Sequence) Next() (KSUID, error) {
	id := seq.Seed // copy
	count := seq.count
	if count > math.MaxUint16 {
		return Nil, errors.New("too many IDs were generated")
	}
	seq.count++
	return withSequenceNumber(id, uint16(count)), nil
}

// Bounds returns the inclusive min and max bounds of the KSUIDs that may be
// generated by the sequence. If all ids have been generated already then the
// returned min value is equal to the max.
func (seq *Sequence) Bounds() (min KSUID, max KSUID) {
	count := seq.count
	if count > math.MaxUint16 {
		count = math.MaxUint16
	}
	return withSequenceNumber(seq.Seed, uint16(count)), withSequenceNumber(seq.Seed, math.MaxUint16)
}

func withSequenceNumber(id KSUID, n uint16) KSUID {
	binary.BigEndian.PutUint16(id[len(id)-2:], n)
	return id
}

const (
	// lexographic ordering (based on Unicode table) is 0-9A-Za-z
	base62Characters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	zeroString       = "000000000000000000000000000"
	offsetUppercase  = 10
	offsetLowercase  = 36
)

var (
	errShortBuffer = errors.New("the output buffer is too small to hold to decoded value")
)

// Converts a base 62 byte into the number value that it represents.
func base62Value(digit byte) byte {
	switch {
	case digit >= '0' && digit <= '9':
		return digit - '0'
	case digit >= 'A' && digit <= 'Z':
		return offsetUppercase + (digit - 'A')
	default:
		return offsetLowercase + (digit - 'a')
	}
}

// This function encodes the base 62 representation of the src KSUID in binary
// form into dst.
//
// In order to support a couple of optimizations the function assumes that src
// is 20 bytes long and dst is 27 bytes long.
//
// Any unused bytes in dst will be set to the padding '0' byte.
func fastEncodeBase62(dst []byte, src []byte) {
	const srcBase = 4294967296
	const dstBase = 62

	// Split src into 5 4-byte words, this is where most of the efficiency comes
	// from because this is a O(N^2) algorithm, and we make N = N / 4 by working
	// on 32 bits at a time.
	parts := [5]uint32{
		binary.BigEndian.Uint32(src[0:4]),
		binary.BigEndian.Uint32(src[4:8]),
		binary.BigEndian.Uint32(src[8:12]),
		binary.BigEndian.Uint32(src[12:16]),
		binary.BigEndian.Uint32(src[16:20]),
	}

	n := len(dst)
	bp := parts[:]
	bq := [5]uint32{}

	for len(bp) != 0 {
		quotient := bq[:0]
		remainder := uint64(0)

		for _, c := range bp {
			value := uint64(c) + uint64(remainder)*srcBase
			digit := value / dstBase
			remainder = value % dstBase

			if len(quotient) != 0 || digit != 0 {
				quotient = append(quotient, uint32(digit))
			}
		}

		// Writes at the end of the destination buffer because we computed the
		// lowest bits first.
		n--
		dst[n] = base62Characters[remainder]
		bp = quotient
	}

	// Add padding at the head of the destination buffer for all bytes that were
	// not set.
	copy(dst[:n], zeroString)
}

// This function appends the base 62 representation of the KSUID in src to dst,
// and returns the extended byte slice.
// The result is left-padded with '0' bytes to always append 27 bytes to the
// destination buffer.
func fastAppendEncodeBase62(dst []byte, src []byte) []byte {
	dst = reserve(dst, stringEncodedLength)
	n := len(dst)
	fastEncodeBase62(dst[n:n+stringEncodedLength], src)
	return dst[:n+stringEncodedLength]
}

// This function decodes the base 62 representation of the src KSUID to the
// binary form into dst.
//
// In order to support a couple of optimizations the function assumes that src
// is 27 bytes long and dst is 20 bytes long.
//
// Any unused bytes in dst will be set to zero.
func fastDecodeBase62(dst []byte, src []byte) error {
	const srcBase = 62
	const dstBase = 4294967296

	// This line helps BCE (Bounds Check Elimination).
	// It may be safely removed.
	_ = src[26]

	parts := [27]byte{
		base62Value(src[0]),
		base62Value(src[1]),
		base62Value(src[2]),
		base62Value(src[3]),
		base62Value(src[4]),
		base62Value(src[5]),
		base62Value(src[6]),
		base62Value(src[7]),
		base62Value(src[8]),
		base62Value(src[9]),

		base62Value(src[10]),
		base62Value(src[11]),
		base62Value(src[12]),
		base62Value(src[13]),
		base62Value(src[14]),
		base62Value(src[15]),
		base62Value(src[16]),
		base62Value(src[17]),
		base62Value(src[18]),
		base62Value(src[19]),

		base62Value(src[20]),
		base62Value(src[21]),
		base62Value(src[22]),
		base62Value(src[23]),
		base62Value(src[24]),
		base62Value(src[25]),
		base62Value(src[26]),
	}

	n := len(dst)
	bp := parts[:]
	bq := [stringEncodedLength]byte{}

	for len(bp) > 0 {
		quotient := bq[:0]
		remainder := uint64(0)

		for _, c := range bp {
			value := uint64(c) + uint64(remainder)*srcBase
			digit := value / dstBase
			remainder = value % dstBase

			if len(quotient) != 0 || digit != 0 {
				quotient = append(quotient, byte(digit))
			}
		}

		if n < 4 {
			return errShortBuffer
		}

		dst[n-4] = byte(remainder >> 24)
		dst[n-3] = byte(remainder >> 16)
		dst[n-2] = byte(remainder >> 8)
		dst[n-1] = byte(remainder)
		n -= 4
		bp = quotient
	}

	var zero [20]byte
	copy(dst[:n], zero[:])
	return nil
}

// This function appends the base 62 decoded version of src into dst.
func fastAppendDecodeBase62(dst []byte, src []byte) []byte {
	dst = reserve(dst, byteLength)
	n := len(dst)
	fastDecodeBase62(dst[n:n+byteLength], src)
	return dst[:n+byteLength]
}

// Ensures that at least nbytes are available in the remaining capacity of the
// destination slice, if not, a new copy is made and returned by the function.
func reserve(dst []byte, nbytes int) []byte {
	c := cap(dst)
	n := len(dst)

	if avail := c - n; avail < nbytes {
		c *= 2
		if (c - n) < nbytes {
			c = n + nbytes
		}
		b := make([]byte, n, c)
		copy(b, dst)
		dst = b
	}

	return dst
}

// CompressedSet is an immutable data type which stores a set of KSUIDs.
type CompressedSet []byte

// Iter returns an iterator that produces all KSUIDs in the set.
func (set CompressedSet) Iter() CompressedSetIter {
	return CompressedSetIter{
		content: set,
	}
}

// String satisfies the fmt.Stringer interface, returns a human-readable string
// representation of the set.
func (set CompressedSet) String() string {
	b := bytes.Buffer{}
	b.WriteByte('[')
	set.writeTo(&b)
	b.WriteByte(']')
	return b.String()
}

// GoString satisfies the fmt.GoStringer interface, returns a Go representation of
// the set.
func (set CompressedSet) GoString() string {
	b := bytes.Buffer{}
	b.WriteString("ksuid.CompressedSet{")
	set.writeTo(&b)
	b.WriteByte('}')
	return b.String()
}

func (set CompressedSet) writeTo(b *bytes.Buffer) {
	a := [27]byte{}

	for i, it := 0, set.Iter(); it.Next(); i++ {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteByte('"')
		it.KSUID.Append(a[:0])
		b.Write(a[:])
		b.WriteByte('"')
	}
}

// Compress creates and returns a compressed set of KSUIDs from the list given
// as arguments.
func Compress(ids ...KSUID) CompressedSet {
	c := 1 + byteLength + (len(ids) / 5)
	b := make([]byte, 0, c)
	return AppendCompressed(b, ids...)
}

// AppendCompressed uses the given byte slice as pre-allocated storage space to
// build a KSUID set.
//
// Note that the set uses a compression technique to store the KSUIDs, so the
// resulting length is not 20 x len(ids). The rule of thumb here is for the given
// byte slice to reserve the amount of memory that the application would be OK
// to waste.
func AppendCompressed(set []byte, ids ...KSUID) CompressedSet {
	if len(ids) != 0 {
		if !IsSorted(ids) {
			Sort(ids)
		}
		one := makeUint128(0, 1)

		// The first KSUID is always written to the set, this is the starting
		// point for all deltas.
		set = append(set, byte(rawKSUID))
		set = append(set, ids[0][:]...)

		timestamp := ids[0].Timestamp()
		lastKSUID := ids[0]
		lastPayload := uint128Payload(ids[0])

		for i := 1; i != len(ids); i++ {
			id := ids[i]

			if id == lastKSUID {
				continue
			}

			t := id.Timestamp()
			p := uint128Payload(id)

			if t != timestamp {
				d := t - timestamp
				n := varintLength32(d)

				set = append(set, timeDelta|byte(n))
				set = appendVarint32(set, d, n)
				set = append(set, id[timestampBytesNum:]...)

				timestamp = t
			} else {
				d := sub128(p, lastPayload)

				if d != one {
					n := varintLength128(d)

					set = append(set, payloadDelta|byte(n))
					set = appendVarint128(set, d, n)
				} else {
					l, c := rangeLength(ids[i+1:], t, id, p)
					m := uint64(l + 1)
					n := varintLength64(m)

					set = append(set, payloadRange|byte(n))
					set = appendVarint64(set, m, n)

					i += c
					id = ids[i]
					p = uint128Payload(id)
				}
			}

			lastKSUID = id
			lastPayload = p
		}
	}
	return set
}

func rangeLength(ids []KSUID, timestamp uint32, lastKSUID KSUID, lastValue uint128) (length int, count int) {
	one := makeUint128(0, 1)

	for i := range ids {
		id := ids[i]

		if id == lastKSUID {
			continue
		}

		if id.Timestamp() != timestamp {
			count = i
			return
		}

		v := uint128Payload(id)

		if sub128(v, lastValue) != one {
			count = i
			return
		}

		lastKSUID = id
		lastValue = v
		length++
	}

	count = len(ids)
	return
}

func appendVarint128(b []byte, v uint128, n int) []byte {
	c := v.bytes()
	return append(b, c[len(c)-n:]...)
}

func appendVarint64(b []byte, v uint64, n int) []byte {
	c := [8]byte{}
	binary.BigEndian.PutUint64(c[:], v)
	return append(b, c[len(c)-n:]...)
}

func appendVarint32(b []byte, v uint32, n int) []byte {
	c := [4]byte{}
	binary.BigEndian.PutUint32(c[:], v)
	return append(b, c[len(c)-n:]...)
}

func varint128(b []byte) uint128 {
	a := [16]byte{}
	copy(a[16-len(b):], b)
	return makeUint128FromPayload(a[:])
}

func varint64(b []byte) uint64 {
	a := [8]byte{}
	copy(a[8-len(b):], b)
	return binary.BigEndian.Uint64(a[:])
}

func varint32(b []byte) uint32 {
	a := [4]byte{}
	copy(a[4-len(b):], b)
	return binary.BigEndian.Uint32(a[:])
}

func varintLength128(v uint128) int {
	if v[1] != 0 {
		return 8 + varintLength64(v[1])
	}
	return varintLength64(v[0])
}

func varintLength64(v uint64) int {
	switch {
	case (v & 0xFFFFFFFFFFFFFF00) == 0:
		return 1
	case (v & 0xFFFFFFFFFFFF0000) == 0:
		return 2
	case (v & 0xFFFFFFFFFF000000) == 0:
		return 3
	case (v & 0xFFFFFFFF00000000) == 0:
		return 4
	case (v & 0xFFFFFF0000000000) == 0:
		return 5
	case (v & 0xFFFF000000000000) == 0:
		return 6
	case (v & 0xFF00000000000000) == 0:
		return 7
	default:
		return 8
	}
}

func varintLength32(v uint32) int {
	switch {
	case (v & 0xFFFFFF00) == 0:
		return 1
	case (v & 0xFFFF0000) == 0:
		return 2
	case (v & 0xFF000000) == 0:
		return 3
	default:
		return 4
	}
}

const (
	rawKSUID     = 0
	timeDelta    = 1 << 6
	payloadDelta = 1 << 7
	payloadRange = (1 << 6) | (1 << 7)
)

// CompressedSetIter is an iterator type returned by Set.Iter to produce the
// list of KSUIDs stored in a set.
//
// Here's is how the iterator type is commonly used:
//
//	for it := set.Iter(); it.Next(); {
//		id := it.KSUID
//		// ...
//	}
//
// CompressedSetIter values are not safe to use concurrently from multiple
// goroutines.
type CompressedSetIter struct {
	// KSUID is modified by calls to the Next method to hold the KSUID loaded
	// by the iterator.
	KSUID

	content []byte
	offset  int

	seqLength uint64
	timestamp uint32
	lastValue uint128
}

// Next moves the iterator forward, returning true if there is a KSUID was found,
// or false if the iterator as reached the end of the set it was created from.
func (it *CompressedSetIter) Next() bool {
	if it.seqLength != 0 {
		value := incr128(it.lastValue)
		it.KSUID = value.ksuid(it.timestamp)
		it.seqLength--
		it.lastValue = value
		return true
	}

	if it.offset == len(it.content) {
		return false
	}

	b := it.content[it.offset]
	it.offset++

	const mask = rawKSUID | timeDelta | payloadDelta | payloadRange
	tag := int(b) & mask
	cnt := int(b) & ^mask

	switch tag {
	case rawKSUID:
		off0 := it.offset
		off1 := off0 + byteLength

		copy(it.KSUID[:], it.content[off0:off1])

		it.offset = off1
		it.timestamp = it.KSUID.Timestamp()
		it.lastValue = uint128Payload(it.KSUID)

	case timeDelta:
		off0 := it.offset
		off1 := off0 + cnt
		off2 := off1 + payloadBytesNum

		it.timestamp += varint32(it.content[off0:off1])

		binary.BigEndian.PutUint32(it.KSUID[:timestampBytesNum], it.timestamp)
		copy(it.KSUID[timestampBytesNum:], it.content[off1:off2])

		it.offset = off2
		it.lastValue = uint128Payload(it.KSUID)

	case payloadDelta:
		off0 := it.offset
		off1 := off0 + cnt

		delta := varint128(it.content[off0:off1])
		value := add128(it.lastValue, delta)

		it.KSUID = value.ksuid(it.timestamp)
		it.offset = off1
		it.lastValue = value

	case payloadRange:
		off0 := it.offset
		off1 := off0 + cnt

		value := incr128(it.lastValue)
		it.KSUID = value.ksuid(it.timestamp)
		it.seqLength = varint64(it.content[off0:off1])
		it.offset = off1
		it.seqLength--
		it.lastValue = value

	default:
		panic("KSUID set iterator is reading malformed data")
	}

	return true
}

// FastRander is an io.Reader that uses math/rand and is optimized for
// generating 16 bytes KSUID payloads. It is intended to be used as a
// performance improvements for programs that have no need for
// cryptographically secure KSUIDs and are generating a lot of them.
var FastRander = newRBG()

func newRBG() io.Reader {
	r, err := newRandomBitsGenerator()
	if err != nil {
		panic(err)
	}
	return r
}

func newRandomBitsGenerator() (r io.Reader, err error) {
	var seed int64

	if seed, err = readCryptoRandomSeed(); err != nil {
		return
	}

	r = &randSourceReader{source: mrand.NewSource(seed).(mrand.Source64)}
	return
}

func readCryptoRandomSeed() (seed int64, err error) {
	var b [8]byte

	if _, err = io.ReadFull(rand.Reader, b[:]); err != nil {
		return
	}

	seed = int64(binary.LittleEndian.Uint64(b[:]))
	return
}

type randSourceReader struct {
	source mrand.Source64
}

func (r *randSourceReader) Read(b []byte) (int, error) {
	// optimized for generating 16 bytes payloads
	binary.LittleEndian.PutUint64(b[:8], r.source.Uint64())
	binary.LittleEndian.PutUint64(b[8:], r.source.Uint64())
	return 16, nil
}
