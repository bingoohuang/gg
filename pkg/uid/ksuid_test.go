package uid

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestConstructionTimestamp(t *testing.T) {
	x := New()
	nowTime := time.Now().Round(1 * time.Minute)
	xTime := x.Time().Round(1 * time.Minute)

	if xTime != nowTime {
		t.Fatal(xTime, "!=", nowTime)
	}
}

func TestNil(t *testing.T) {
	if !Nil.IsNil() {
		t.Fatal("Nil should be Nil!")
	}

	x, _ := FromBytes(make([]byte, byteLength))
	if !x.IsNil() {
		t.Fatal("Zero-byte array should be Nil!")
	}
}

func TestEncoding(t *testing.T) {
	x, _ := FromBytes(make([]byte, byteLength))
	if !x.IsNil() {
		t.Fatal("Zero-byte array should be Nil!")
	}

	encoded := x.String()
	expected := strings.Repeat("0", stringEncodedLength)

	if encoded != expected {
		t.Fatal("expected", expected, "encoded", encoded)
	}
}

func TestPadding(t *testing.T) {
	b := make([]byte, byteLength)
	for i := 0; i < byteLength; i++ {
		b[i] = 255
	}

	x, _ := FromBytes(b)
	xEncoded := x.String()
	nilEncoded := Nil.String()

	if len(xEncoded) != len(nilEncoded) {
		t.Fatal("Encoding should produce equal-length strings for zero and max case")
	}
}

func TestParse(t *testing.T) {
	_, err := Parse("123")
	if err != errStrSize {
		t.Fatal("Expected Parsing a 3-char string to return an error")
	}

	parsed, err := Parse(strings.Repeat("0", stringEncodedLength))
	if err != nil {
		t.Fatal("Unexpected error", err)
	}

	if Compare(parsed, Nil) != 0 {
		t.Fatal("Parsing all-zeroes string should equal Nil value",
			"expected:", Nil,
			"actual:", parsed)
	}

	maxBytes := make([]byte, byteLength)
	for i := 0; i < byteLength; i++ {
		maxBytes[i] = 255
	}
	maxBytesKSUID, err := FromBytes(maxBytes)
	if err != nil {
		t.Fatal("Unexpected error", err)
	}

	maxParseKSUID, err := Parse(maxStringEncoded)
	if err != nil {
		t.Fatal("Unexpected error", err)
	}

	if Compare(maxBytesKSUID, maxParseKSUID) != 0 {
		t.Fatal("String decoder broke for max string")
	}
}

func TestIssue25(t *testing.T) {
	// https://github.com/segmentio/ksuid/issues/25
	for _, s := range []string{
		"aaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"aWgEPTl1tmebfsQzFP4bxwgy80!",
	} {
		_, err := Parse(s)
		if err != errStrValue {
			t.Error("invalid KSUID representations cannot be successfully parsed, got err =", err)
		}
	}
}

func TestEncodeAndDecode(t *testing.T) {
	x := New()
	builtFromEncodedString, err := Parse(x.String())
	if err != nil {
		t.Fatal("Unexpected error", err)
	}

	if Compare(x, builtFromEncodedString) != 0 {
		t.Fatal("Parse(X).String() != X")
	}
}

func TestMarshalText(t *testing.T) {
	var id1 = New()
	var id2 KSUID

	if err := id2.UnmarshalText([]byte(id1.String())); err != nil {
		t.Fatal(err)
	}

	if id1 != id2 {
		t.Fatal(id1, "!=", id2)
	}

	if b, err := id2.MarshalText(); err != nil {
		t.Fatal(err)
	} else if s := string(b); s != id1.String() {
		t.Fatal(s)
	}
}

func TestMarshalBinary(t *testing.T) {
	var id1 = New()
	var id2 KSUID

	if err := id2.UnmarshalBinary(id1.Bytes()); err != nil {
		t.Fatal(err)
	}

	if id1 != id2 {
		t.Fatal(id1, "!=", id2)
	}

	if b, err := id2.MarshalBinary(); err != nil {
		t.Fatal(err)
	} else if bytes.Compare(b, id1.Bytes()) != 0 {
		t.Fatal("bad binary form:", id2)
	}
}

func TestMarshalJSON(t *testing.T) {
	var id1 = New()
	var id2 KSUID

	if b, err := json.Marshal(id1); err != nil {
		t.Fatal(err)
	} else if err := json.Unmarshal(b, &id2); err != nil {
		t.Fatal(err)
	} else if id1 != id2 {
		t.Error(id1, "!=", id2)
	}
}

func TestFlag(t *testing.T) {
	var id1 = New()
	var id2 KSUID

	fset := flag.NewFlagSet("test", flag.ContinueOnError)
	fset.Var(&id2, "id", "the KSUID")

	if err := fset.Parse([]string{"-id", id1.String()}); err != nil {
		t.Fatal(err)
	}

	if id1 != id2 {
		t.Error(id1, "!=", id2)
	}
}

func TestSqlValuer(t *testing.T) {
	id, _ := Parse(maxStringEncoded)

	if v, err := id.Value(); err != nil {
		t.Error(err)
	} else if s, ok := v.(string); !ok {
		t.Error("not a string value")
	} else if s != maxStringEncoded {
		t.Error("bad string value::", s)
	}
}

func TestSqlValuerNilValue(t *testing.T) {
	if v, err := Nil.Value(); err != nil {
		t.Error(err)
	} else if v != nil {
		t.Errorf("bad nil value: %v", v)
	}
}

func TestSqlScanner(t *testing.T) {
	id1 := New()
	id2 := New()

	tests := []struct {
		ksuid KSUID
		value interface{}
	}{
		{Nil, nil},
		{id1, id1.String()},
		{id2, id2.Bytes()},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%T", test.value), func(t *testing.T) {
			var id KSUID

			if err := id.Scan(test.value); err != nil {
				t.Error(err)
			}

			if id != test.ksuid {
				t.Error("bad KSUID:")
				t.Logf("expected %v", test.ksuid)
				t.Logf("found    %v", id)
			}
		})
	}
}

func TestAppend(t *testing.T) {
	for _, repr := range []string{"0pN1Own7255s7jwpwy495bAZeEa", "aWgEPTl1tmebfsQzFP4bxwgy80V"} {
		k, _ := Parse(repr)
		a := make([]byte, 0, stringEncodedLength)

		a = append(a, "?: "...)
		a = k.Append(a)

		if s := string(a); s != "?: "+repr {
			t.Error(s)
		}
	}
}

func TestSort(t *testing.T) {
	ids1 := [11]KSUID{}
	ids2 := [11]KSUID{}

	for i := range ids1 {
		ids1[i] = New()
	}

	ids2 = ids1
	sort.Slice(ids2[:], func(i, j int) bool {
		return Compare(ids2[i], ids2[j]) < 0
	})

	Sort(ids1[:])

	if !IsSorted(ids1[:]) {
		t.Error("not sorted")
	}

	if ids1 != ids2 {
		t.Error("bad order:")
		t.Log(ids1)
		t.Log(ids2)
	}
}

func TestPrevNext(t *testing.T) {
	tests := []struct {
		id   KSUID
		prev KSUID
		next KSUID
	}{
		{
			id:   Nil,
			prev: Max,
			next: KSUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		},
		{
			id:   Max,
			prev: KSUID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe},
			next: Nil,
		},
	}

	for _, test := range tests {
		t.Run(test.id.String(), func(t *testing.T) {
			testPrevNext(t, test.id, test.prev, test.next)
		})
	}
}

func TestGetTimestamp(t *testing.T) {
	nowTime := time.Now()
	x, _ := NewRandomWithTime(nowTime)
	xTime := int64(x.Timestamp())
	unix := nowTime.Unix()
	if xTime != unix-epochStamp {
		t.Fatal(xTime, "!=", unix)
	}
}

func testPrevNext(t *testing.T, id, prev, next KSUID) {
	id1 := id.Prev()
	id2 := id.Next()

	if id1 != prev {
		t.Error("previous id of the nil KSUID is wrong:", id1, "!=", prev)
	}

	if id2 != next {
		t.Error("next id of the nil KSUID is wrong:", id2, "!=", next)
	}
}

func BenchmarkAppend(b *testing.B) {
	a := make([]byte, 0, stringEncodedLength)
	k := New()

	for i := 0; i != b.N; i++ {
		k.Append(a)
	}
}

func BenchmarkString(b *testing.B) {
	k := New()

	for i := 0; i != b.N; i++ {
		_ = k.String()
	}
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i != b.N; i++ {
		Parse(maxStringEncoded)
	}
}

func BenchmarkCompare(b *testing.B) {
	k1 := New()
	k2 := New()

	for i := 0; i != b.N; i++ {
		Compare(k1, k2)
	}
}

func BenchmarkSort(b *testing.B) {
	ids1 := [101]KSUID{}
	ids2 := [101]KSUID{}

	for i := range ids1 {
		ids1[i] = New()
	}

	for i := 0; i != b.N; i++ {
		ids2 = ids1
		Sort(ids2[:])
	}
}

func BenchmarkNew(b *testing.B) {
	b.Run("with crypto rand", func(b *testing.B) {
		SetRand(nil)
		for i := 0; i != b.N; i++ {
			New()
		}
	})
	b.Run("with math rand", func(b *testing.B) {
		SetRand(FastRander)
		for i := 0; i != b.N; i++ {
			New()
		}
	})
}

func TestCmp128(t *testing.T) {
	tests := []struct {
		x uint128
		y uint128
		k int
	}{
		{
			x: makeUint128(0, 0),
			y: makeUint128(0, 0),
			k: 0,
		},
		{
			x: makeUint128(0, 1),
			y: makeUint128(0, 0),
			k: +1,
		},
		{
			x: makeUint128(0, 0),
			y: makeUint128(0, 1),
			k: -1,
		},
		{
			x: makeUint128(1, 0),
			y: makeUint128(0, 1),
			k: +1,
		},
		{
			x: makeUint128(0, 1),
			y: makeUint128(1, 0),
			k: -1,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("cmp128(%s,%s)", test.x, test.y), func(t *testing.T) {
			if k := cmp128(test.x, test.y); k != test.k {
				t.Error(k, "!=", test.k)
			}
		})
	}
}

func TestAdd128(t *testing.T) {
	tests := []struct {
		x uint128
		y uint128
		z uint128
	}{
		{
			x: makeUint128(0, 0),
			y: makeUint128(0, 0),
			z: makeUint128(0, 0),
		},
		{
			x: makeUint128(0, 1),
			y: makeUint128(0, 0),
			z: makeUint128(0, 1),
		},
		{
			x: makeUint128(0, 0),
			y: makeUint128(0, 1),
			z: makeUint128(0, 1),
		},
		{
			x: makeUint128(1, 0),
			y: makeUint128(0, 1),
			z: makeUint128(1, 1),
		},
		{
			x: makeUint128(0, 1),
			y: makeUint128(1, 0),
			z: makeUint128(1, 1),
		},
		{
			x: makeUint128(0, 0xFFFFFFFFFFFFFFFF),
			y: makeUint128(0, 1),
			z: makeUint128(1, 0),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("add128(%s,%s)", test.x, test.y), func(t *testing.T) {
			if z := add128(test.x, test.y); z != test.z {
				t.Error(z, "!=", test.z)
			}
		})
	}
}

func TestSub128(t *testing.T) {
	tests := []struct {
		x uint128
		y uint128
		z uint128
	}{
		{
			x: makeUint128(0, 0),
			y: makeUint128(0, 0),
			z: makeUint128(0, 0),
		},
		{
			x: makeUint128(0, 1),
			y: makeUint128(0, 0),
			z: makeUint128(0, 1),
		},
		{
			x: makeUint128(0, 0),
			y: makeUint128(0, 1),
			z: makeUint128(0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF),
		},
		{
			x: makeUint128(1, 0),
			y: makeUint128(0, 1),
			z: makeUint128(0, 0xFFFFFFFFFFFFFFFF),
		},
		{
			x: makeUint128(0, 1),
			y: makeUint128(1, 0),
			z: makeUint128(0xFFFFFFFFFFFFFFFF, 1),
		},
		{
			x: makeUint128(0, 0xFFFFFFFFFFFFFFFF),
			y: makeUint128(0, 1),
			z: makeUint128(0, 0xFFFFFFFFFFFFFFFE),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("sub128(%s,%s)", test.x, test.y), func(t *testing.T) {
			if z := sub128(test.x, test.y); z != test.z {
				t.Error(z, "!=", test.z)
			}
		})
	}
}

func BenchmarkCmp128(b *testing.B) {
	x := makeUint128(0, 0)
	y := makeUint128(0, 0)

	for i := 0; i != b.N; i++ {
		cmp128(x, y)
	}
}

func BenchmarkAdd128(b *testing.B) {
	x := makeUint128(0, 0)
	y := makeUint128(0, 0)

	for i := 0; i != b.N; i++ {
		add128(x, y)
	}
}

func BenchmarkSub128(b *testing.B) {
	x := makeUint128(0, 0)
	y := makeUint128(0, 0)

	for i := 0; i != b.N; i++ {
		sub128(x, y)
	}
}

func TestSequence(t *testing.T) {
	seq := Sequence{Seed: New()}

	if min, max := seq.Bounds(); min == max {
		t.Error("min and max of KSUID range must differ when no ids have been generated")
	}

	for i := 0; i <= math.MaxUint16; i++ {
		id, err := seq.Next()
		if err != nil {
			t.Fatal(err)
		}
		if j := int(binary.BigEndian.Uint16(id[len(id)-2:])); j != i {
			t.Fatalf("expected %d but got %d in %s", i, j, id)
		}
	}

	if _, err := seq.Next(); err == nil {
		t.Fatal("no error returned after exhausting the id generator")
	}

	if min, max := seq.Bounds(); min != max {
		t.Error("after all KSUIDs were generated the min and max must be equal")
	}
}

func TestBase10ToBase62AndBack(t *testing.T) {
	number := []byte{1, 2, 3, 4}
	encoded := base2base(number, 10, 62)
	decoded := base2base(encoded, 62, 10)

	if bytes.Compare(number, decoded) != 0 {
		t.Fatal(number, " != ", decoded)
	}
}

func TestBase256ToBase62AndBack(t *testing.T) {
	number := []byte{255, 254, 253, 251}
	encoded := base2base(number, 256, 62)
	decoded := base2base(encoded, 62, 256)

	if bytes.Compare(number, decoded) != 0 {
		t.Fatal(number, " != ", decoded)
	}
}

func TestEncodeAndDecodeBase62(t *testing.T) {
	helloWorld := []byte("hello world")
	encoded := encodeBase62(helloWorld)
	decoded := decodeBase62(encoded)

	if len(encoded) < len(helloWorld) {
		t.Fatal("length of encoded base62 string", encoded, "should be >= than raw bytes!")

	}

	if bytes.Compare(helloWorld, decoded) != 0 {
		t.Fatal(decoded, " != ", helloWorld)
	}
}

func TestLexographicOrdering(t *testing.T) {
	unsortedStrings := make([]string, 256)
	for i := 0; i < 256; i++ {
		s := string(encodeBase62([]byte{0, byte(i)}))
		unsortedStrings[i] = strings.Repeat("0", 2-len(s)) + s
	}

	if !sort.StringsAreSorted(unsortedStrings) {
		sortedStrings := make([]string, len(unsortedStrings))
		for i, s := range unsortedStrings {
			sortedStrings[i] = s
		}
		sort.Strings(sortedStrings)

		t.Fatal("base62 encoder does not produce lexographically sorted output.",
			"expected:", sortedStrings,
			"actual:", unsortedStrings)
	}
}

func TestBase62Value(t *testing.T) {
	s := base62Characters

	for i := range s {
		v := int(base62Value(s[i]))

		if v != i {
			t.Error("bad value:")
			t.Log("<<<", i)
			t.Log(">>>", v)
		}
	}
}

func TestFastAppendEncodeBase62(t *testing.T) {
	for i := 0; i != 1000; i++ {
		id := New()

		b0 := id[:]
		b1 := appendEncodeBase62(nil, b0)
		b2 := fastAppendEncodeBase62(nil, b0)

		s1 := string(leftpad(b1, '0', stringEncodedLength))
		s2 := string(b2)

		if s1 != s2 {
			t.Error("bad base62 representation of", id)
			t.Log("<<<", s1, len(s1))
			t.Log(">>>", s2, len(s2))
		}
	}
}

func TestFastAppendDecodeBase62(t *testing.T) {
	for i := 0; i != 1000; i++ {
		id := New()
		b0 := leftpad(encodeBase62(id[:]), '0', stringEncodedLength)

		b1 := appendDecodeBase62(nil, []byte(string(b0))) // because it modifies the input buffer
		b2 := fastAppendDecodeBase62(nil, b0)

		if !bytes.Equal(leftpad(b1, 0, byteLength), b2) {
			t.Error("bad binary representation of", string(b0))
			t.Log("<<<", b1)
			t.Log(">>>", b2)
		}
	}
}

func BenchmarkAppendEncodeBase62(b *testing.B) {
	a := [stringEncodedLength]byte{}
	id := New()

	for i := 0; i != b.N; i++ {
		appendEncodeBase62(a[:0], id[:])
	}
}

func BenchmarkAppendFastEncodeBase62(b *testing.B) {
	a := [stringEncodedLength]byte{}
	id := New()

	for i := 0; i != b.N; i++ {
		fastAppendEncodeBase62(a[:0], id[:])
	}
}

func BenchmarkAppendDecodeBase62(b *testing.B) {
	a := [byteLength]byte{}
	id := []byte(New().String())

	for i := 0; i != b.N; i++ {
		b := [stringEncodedLength]byte{}
		copy(b[:], id)
		appendDecodeBase62(a[:0], b[:])
	}
}

func BenchmarkAppendFastDecodeBase62(b *testing.B) {
	a := [byteLength]byte{}
	id := []byte(New().String())

	for i := 0; i != b.N; i++ {
		fastAppendDecodeBase62(a[:0], id)
	}
}

// The functions bellow were the initial implementation of the base conversion
// algorithms, they were replaced by optimized versions later on. We keep them
// in the test files as a reference to ensure compatibility between the generic
// and optimized implementations.
func appendBase2Base(dst []byte, src []byte, inBase int, outBase int) []byte {
	off := len(dst)
	bs := src[:]
	bq := [stringEncodedLength]byte{}

	for len(bs) > 0 {
		length := len(bs)
		quotient := bq[:0]
		remainder := 0

		for i := 0; i != length; i++ {
			acc := int(bs[i]) + remainder*inBase
			d := acc/outBase | 0
			remainder = acc % outBase

			if len(quotient) > 0 || d > 0 {
				quotient = append(quotient, byte(d))
			}
		}

		// Appends in reverse order, the byte slice gets reversed before it's
		// returned by the function.
		dst = append(dst, byte(remainder))
		bs = quotient
	}

	reverse(dst[off:])
	return dst
}

func base2base(src []byte, inBase int, outBase int) []byte {
	return appendBase2Base(nil, src, inBase, outBase)
}

func appendEncodeBase62(dst []byte, src []byte) []byte {
	off := len(dst)
	dst = appendBase2Base(dst, src, 256, 62)
	for i, c := range dst[off:] {
		dst[off+i] = base62Characters[c]
	}
	return dst
}

func encodeBase62(in []byte) []byte {
	return appendEncodeBase62(nil, in)
}

func appendDecodeBase62(dst []byte, src []byte) []byte {
	// Kind of intrusive, we modify the input buffer... it's OK here, it saves
	// a memory allocation in Parse.
	for i, b := range src {
		// O(1)... technically. Has better real-world perf than a map
		src[i] = byte(strings.IndexByte(base62Characters, b))
	}
	return appendBase2Base(dst, src, 62, 256)
}

func decodeBase62(src []byte) []byte {
	return appendDecodeBase62(
		make([]byte, 0, len(src)*2),
		append(make([]byte, 0, len(src)), src...),
	)
}

func reverse(b []byte) {
	i := 0
	j := len(b) - 1

	for i < j {
		b[i], b[j] = b[j], b[i]
		i++
		j--
	}
}

func leftpad(b []byte, c byte, n int) []byte {
	if n -= len(b); n > 0 {
		for i := 0; i != n; i++ {
			b = append(b, c)
		}

		copy(b[n:], b)

		for i := 0; i != n; i++ {
			b[i] = c
		}
	}
	return b
}

func TestCompressedSet(t *testing.T) {
	tests := []struct {
		scenario string
		function func(*testing.T)
	}{
		{
			scenario: "String",
			function: testCompressedSetString,
		},
		{
			scenario: "GoString",
			function: testCompressedSetGoString,
		},
		{
			scenario: "sparse",
			function: testCompressedSetSparse,
		},
		{
			scenario: "packed",
			function: testCompressedSetPacked,
		},
		{
			scenario: "mixed",
			function: testCompressedSetMixed,
		},
		{
			scenario: "iterating over a nil compressed set returns no ids",
			function: testCompressedSetNil,
		},
		{
			scenario: "concatenating multiple compressed sets is supported",
			function: testCompressedSetConcat,
		},
		{
			scenario: "duplicate ids are appear only once in the compressed set",
			function: testCompressedSetDuplicates,
		},
		{
			scenario: "building a compressed set with a single id repeated multiple times produces the id only once",
			function: testCompressedSetSingle,
		},
		{
			scenario: "iterating over a compressed sequence returns the full sequence",
			function: testCompressedSetSequence,
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, test.function)
	}
}

func testCompressedSetString(t *testing.T) {
	id1, _ := Parse("0uHjRkQoL2JKAQIULPdqqb5fOkk")
	id2, _ := Parse("0uHjRvkOG5CbtoXW5oCEp3L2xBu")
	id3, _ := Parse("0uHjSJ4Pe5606kT2XWixK6dirlo")

	set := Compress(id1, id2, id3)

	if s := set.String(); s != `["0uHjRkQoL2JKAQIULPdqqb5fOkk", "0uHjRvkOG5CbtoXW5oCEp3L2xBu", "0uHjSJ4Pe5606kT2XWixK6dirlo"]` {
		t.Error(s)
	}
}

func testCompressedSetGoString(t *testing.T) {
	id1, _ := Parse("0uHjRkQoL2JKAQIULPdqqb5fOkk")
	id2, _ := Parse("0uHjRvkOG5CbtoXW5oCEp3L2xBu")
	id3, _ := Parse("0uHjSJ4Pe5606kT2XWixK6dirlo")

	set := Compress(id1, id2, id3)

	if s := set.GoString(); s != `ksuid.CompressedSet{"0uHjRkQoL2JKAQIULPdqqb5fOkk", "0uHjRvkOG5CbtoXW5oCEp3L2xBu", "0uHjSJ4Pe5606kT2XWixK6dirlo"}` {
		t.Error(s)
	}
}

func testCompressedSetSparse(t *testing.T) {
	now := time.Now()

	times := [100]time.Time{}
	for i := range times {
		times[i] = now.Add(time.Duration(i) * 2 * time.Second)
	}

	ksuids := [1000]KSUID{}
	for i := range ksuids {
		ksuids[i], _ = NewRandomWithTime(times[i%len(times)])
	}

	set := Compress(ksuids[:]...)

	for i, it := 0, set.Iter(); it.Next(); {
		if i >= len(ksuids) {
			t.Error("too many KSUIDs were produced by the set iterator")
			break
		}
		if ksuids[i] != it.KSUID {
			t.Errorf("bad KSUID at index %d: expected %s but found %s", i, ksuids[i], it.KSUID)
		}
		i++
	}

	reportCompressionRatio(t, ksuids[:], set)
}

func testCompressedSetPacked(t *testing.T) {
	sequences := [10]Sequence{}
	for i := range sequences {
		sequences[i] = Sequence{Seed: New()}
	}

	ksuids := [1000]KSUID{}
	for i := range ksuids {
		ksuids[i], _ = sequences[i%len(sequences)].Next()
	}

	set := Compress(ksuids[:]...)

	for i, it := 0, set.Iter(); it.Next(); {
		if i >= len(ksuids) {
			t.Error("too many KSUIDs were produced by the set iterator")
			break
		}
		if ksuids[i] != it.KSUID {
			t.Errorf("bad KSUID at index %d: expected %s but found %s", i, ksuids[i], it.KSUID)
		}
		i++
	}

	reportCompressionRatio(t, ksuids[:], set)
}

func testCompressedSetMixed(t *testing.T) {
	now := time.Now()

	times := [20]time.Time{}
	for i := range times {
		times[i] = now.Add(time.Duration(i) * 2 * time.Second)
	}

	sequences := [200]Sequence{}
	for i := range sequences {
		seed, _ := NewRandomWithTime(times[i%len(times)])
		sequences[i] = Sequence{Seed: seed}
	}

	ksuids := [1000]KSUID{}
	for i := range ksuids {
		ksuids[i], _ = sequences[i%len(sequences)].Next()
	}

	set := Compress(ksuids[:]...)

	for i, it := 0, set.Iter(); it.Next(); {
		if i >= len(ksuids) {
			t.Error("too many KSUIDs were produced by the set iterator")
			break
		}
		if ksuids[i] != it.KSUID {
			t.Errorf("bad KSUID at index %d: expected %s but found %s", i, ksuids[i], it.KSUID)
		}
		i++
	}

	reportCompressionRatio(t, ksuids[:], set)
}

func testCompressedSetDuplicates(t *testing.T) {
	sequence := Sequence{Seed: New()}

	ksuids := [1000]KSUID{}
	for i := range ksuids[:10] {
		ksuids[i], _ = sequence.Next() // exercise dedupe on the id range code path
	}
	for i := range ksuids[10:] {
		ksuids[i+10] = New()
	}
	for i := 1; i < len(ksuids); i += 4 {
		ksuids[i] = ksuids[i-1] // generate many dupes
	}

	miss := make(map[KSUID]struct{})
	uniq := make(map[KSUID]struct{})

	for _, id := range ksuids {
		miss[id] = struct{}{}
	}

	set := Compress(ksuids[:]...)

	for it := set.Iter(); it.Next(); {
		if _, dupe := uniq[it.KSUID]; dupe {
			t.Errorf("duplicate id found in compressed set: %s", it.KSUID)
		}
		uniq[it.KSUID] = struct{}{}
		delete(miss, it.KSUID)
	}

	if len(miss) != 0 {
		t.Error("some ids were not found in the compressed set:")
		for id := range miss {
			t.Log(id)
		}
	}
}

func testCompressedSetSingle(t *testing.T) {
	id := New()

	set := Compress(
		id, id, id, id, id, id, id, id, id, id,
		id, id, id, id, id, id, id, id, id, id,
		id, id, id, id, id, id, id, id, id, id,
		id, id, id, id, id, id, id, id, id, id,
	)

	n := 0

	for it := set.Iter(); it.Next(); {
		if n != 0 {
			t.Errorf("too many ids found in the compressed set: %s", it.KSUID)
		} else if id != it.KSUID {
			t.Errorf("invalid id found in the compressed set: %s != %s", it.KSUID, id)
		}
		n++
	}

	if n == 0 {
		t.Error("no ids were produced by the compressed set")
	}
}

func testCompressedSetSequence(t *testing.T) {
	seq := Sequence{Seed: New()}

	ids := make([]KSUID, 5)

	for i := 0; i < 5; i++ {
		ids[i], _ = seq.Next()
	}

	iter := Compress(ids...).Iter()

	index := 0
	for iter.Next() {
		if iter.KSUID != ids[index] {
			t.Errorf("mismatched id at index %d: %s != %s", index, iter.KSUID, ids[index])
		}
		index++
	}

	if index != 5 {
		t.Errorf("Expected 5 ids, got %d", index)
	}
}

func testCompressedSetNil(t *testing.T) {
	set := CompressedSet(nil)

	for it := set.Iter(); it.Next(); {
		t.Errorf("too many ids returned by the iterator of a nil compressed set: %s", it.KSUID)
	}
}

func testCompressedSetConcat(t *testing.T) {
	ksuids := [100]KSUID{}

	for i := range ksuids {
		ksuids[i] = New()
	}

	set := CompressedSet(nil)
	set = AppendCompressed(set, ksuids[:42]...)
	set = AppendCompressed(set, ksuids[42:64]...)
	set = AppendCompressed(set, ksuids[64:]...)

	for i, it := 0, set.Iter(); it.Next(); i++ {
		if ksuids[i] != it.KSUID {
			t.Errorf("invalid ID at index %d: %s != %s", i, ksuids[i], it.KSUID)
		}
	}
}

func reportCompressionRatio(t *testing.T, ksuids []KSUID, set CompressedSet) {
	len1 := byteLength * len(ksuids)
	len2 := len(set)
	t.Logf("original %d B, compressed %d B (%.4g%%)", len1, len2, 100*(1-(float64(len2)/float64(len1))))
}

func BenchmarkCompressedSet(b *testing.B) {
	ksuids1 := [1000]KSUID{}
	ksuids2 := [1000]KSUID{}

	for i := range ksuids1 {
		ksuids1[i] = New()
	}

	ksuids2 = ksuids1
	buf := make([]byte, 0, 1024)
	set := Compress(ksuids2[:]...)

	b.Run("write", func(b *testing.B) {
		n := 0
		for i := 0; i != b.N; i++ {
			ksuids2 = ksuids1
			buf = AppendCompressed(buf[:0], ksuids2[:]...)
			n = len(buf)
		}
		b.SetBytes(int64(n + len(ksuids2)))
	})

	b.Run("read", func(b *testing.B) {
		n := 0
		for i := 0; i != b.N; i++ {
			n = 0
			for it := set.Iter(); true; {
				if !it.Next() {
					n++
					break
				}
			}
		}
		b.SetBytes(int64((n * byteLength) + len(set)))
	})
}
