// nolint gomnd
package snow

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNodeLocalIP(t *testing.T) {
	that := assert.New(t)

	node, err := NewNode(WithNodeIDLocalIP(0, "192.168.0.10"))

	that.Nil(err)
	that.Equal(int64(10), node.GetNodeID())
	that.Equal(int64(10), node.NodeIDOf(node.Next()))

	node, err = NewNode(WithNodeIDLocalIP(1, "192.168.0.10"))

	that.Nil(err)
	that.Equal(int64(1<<8|10), node.GetNodeID())
	that.Equal(int64(1<<8|10), node.NodeIDOf(node.Next()))

	node, err = NewNode(WithEpoch(12345678))

	that.Nil(err)
	that.True(node.TimeOf(node.Next()) > node.GetEpoch())
}

//******************************************************************************
// General Test funcs

func TestNewNode(t *testing.T) {
	_, err := NewNode(WithNodeID(0))
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	_, err = NewNode(WithNodeID(5000))
	if err == nil {
		t.Fatalf("no error creating NewNode, %s", err)
	}
}

// lazy check if Next will create duplicate IDs
// would be good to later enhance this with more smarts
func TestGenerateDuplicateID(t *testing.T) {
	node, _ := NewNode(WithNodeID(1))

	var x, y ID
	for i := 0; i < 1000000; i++ {
		y = node.Next()
		if x == y {
			t.Errorf("x(%d) & y(%d) are the same", x, y)
		}

		x = y
	}
}

// I feel like there's probably a better way
func TestRace(t *testing.T) {
	node, _ := NewNode(WithNodeID(1))

	go func() {
		for i := 0; i < 1000000000; i++ {
			NewNode(WithNodeID(1))
		}
	}()

	for i := 0; i < 4000; i++ {
		node.Next()
	}
}

//******************************************************************************
// Converters/Parsers Test funcs
// We should have funcs here to test conversion both ways for everything

func TestPrintAll(t *testing.T) {
	node, err := NewNode(WithNodeID(0))
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	id := node.Next()

	t.Logf("Int32    : %d", id.Int32())
	t.Logf("Uint32   : %d", id.Uint32())
	t.Logf("Int64    : %d", id.Int64())
	t.Logf("Uint64   : %d", id.Uint64())
	t.Logf("String   : %s", id.String())
	t.Logf("Base2    : %s", id.Base2())
	t.Logf("Base32   : %s", id.Base32())
	t.Logf("Base36   : %s", id.Base36())
	t.Logf("Base58   : %s", id.Base58())
	t.Logf("Base64   : %s", id.Base64())
	t.Logf("Bytes    : %v", id.Bytes())
	t.Logf("IntBytes : %v", id.IntBytes())
}

func TestInt64(t *testing.T) {
	node, err := NewNode(WithNodeID(0))
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	oID := node.Next()
	i := oID.Int64()

	pID := ParseInt64(i)
	if pID != oID {
		t.Fatalf("pID %v != oID %v", pID, oID)
	}

	mi := int64(1116766490855473152)
	pID = ParseInt64(mi)

	if pID.Int64() != mi {
		t.Fatalf("pID %v != mi %v", pID.Int64(), mi)
	}
}

func TestString(t *testing.T) {
	node, err := NewNode(WithNodeID(0))
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	oID := node.Next()
	si := oID.String()

	pID, err := ParseString(si)
	if err != nil {
		t.Fatalf("error parsing, %s", err)
	}

	if pID != oID {
		t.Fatalf("pID %v != oID %v", pID, oID)
	}

	ms := `1116766490855473152`
	_, err = ParseString(ms)

	if err != nil {
		t.Fatalf("error parsing, %s", err)
	}

	ms = `1112316766490855473152`

	_, err = ParseString(ms)

	if err == nil {
		t.Fatalf("no error parsing %s", ms)
	}
}

func TestBase2(t *testing.T) {
	node, err := NewNode(WithNodeID(0))
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	oID := node.Next()
	i := oID.Base2()

	pID, err := ParseBase2(i)

	if err != nil {
		t.Fatalf("error parsing, %s", err)
	}

	if pID != oID {
		t.Fatalf("pID %v != oID %v", pID, oID)
	}

	ms := `111101111111101110110101100101001000000000000000000000000000`
	_, err = ParseBase2(ms)

	if err != nil {
		t.Fatalf("error parsing, %s", err)
	}

	ms = `1112316766490855473152`
	_, err = ParseBase2(ms)

	if err == nil {
		t.Fatalf("no error parsing %s", ms)
	}
}

func TestBase32(t *testing.T) {
	node, err := NewNode(WithNodeID(0))
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	for i := 0; i < 100; i++ {
		sf := node.Next()
		b32i := sf.Base32()
		psf, err := ParseBase32([]byte(b32i))

		if err != nil {
			t.Fatal(err)
		}

		if sf != psf {
			t.Fatal("Parsed does not match String.")
		}
	}
}

func TestBase36(t *testing.T) {
	node, err := NewNode(WithNodeID(0))
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	oID := node.Next()
	i := oID.Base36()

	pID, err := ParseBase36(i)
	if err != nil {
		t.Fatalf("error parsing, %s", err)
	}

	if pID != oID {
		t.Fatalf("pID %v != oID %v", pID, oID)
	}

	ms := `8hgmw4blvlkw`
	_, err = ParseBase36(ms)

	if err != nil {
		t.Fatalf("error parsing, %s", err)
	}

	ms = `68h5gmw443blv2lk1w`
	_, err = ParseBase36(ms)

	if err == nil {
		t.Fatalf("no error parsing, %s", err)
	}
}

func TestBase58(t *testing.T) {
	node, err := NewNode(WithNodeID(0))
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	for i := 0; i < 10; i++ {
		sf := node.Next()
		b58 := sf.Base58()
		psf, err := ParseBase58([]byte(b58))

		if err != nil {
			t.Fatal(err)
		}

		if sf != psf {
			t.Fatal("Parsed does not match String.")
		}
	}
}

func TestBase64(t *testing.T) {
	node, err := NewNode(WithNodeID(0))
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	oID := node.Next()
	i := oID.Base64()

	pID, err := ParseBase64(i)
	if err != nil {
		t.Fatalf("error parsing, %s", err)
	}

	if pID != oID {
		t.Fatalf("pID %v != oID %v", pID, oID)
	}

	ms := `MTExNjgxOTQ5NDY2MDk5NzEyMA==`
	_, err = ParseBase64(ms)

	if err != nil {
		t.Fatalf("error parsing, %s", err)
	}

	ms = `MTExNjgxOTQ5NDY2MDk5NzEyMA`
	_, err = ParseBase64(ms)

	if err == nil {
		t.Fatalf("no error parsing, %s", err)
	}
}

func TestBytes(t *testing.T) {
	node, err := NewNode(WithNodeID(0))
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	oID := node.Next()
	i := oID.Bytes()

	pID, err := ParseBytes(i)
	if err != nil {
		t.Fatalf("error parsing, %s", err)
	}

	if pID != oID {
		t.Fatalf("pID %v != oID %v", pID, oID)
	}

	ms := []byte{0x31, 0x31, 0x31, 0x36, 0x38, 0x32, 0x31, 0x36, 0x37, 0x39, 0x35, 0x37, 0x30, 0x34, 0x31, 0x39, 0x37, 0x31, 0x32}
	_, err = ParseBytes(ms)

	if err != nil {
		t.Fatalf("error parsing, %#v", err)
	}

	ms = []byte{0xFF, 0xFF, 0xFF, 0x31, 0x31, 0x31, 0x36, 0x38, 0x32, 0x31, 0x36, 0x37, 0x39, 0x35, 0x37, 0x30, 0x34, 0x31, 0x39, 0x37, 0x31, 0x32}
	_, err = ParseBytes(ms)

	if err == nil {
		t.Fatalf("no error parsing, %#v", err)
	}
}

func TestIntBytes(t *testing.T) {
	node, err := NewNode(WithNodeID(0))
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	oID := node.Next()
	i := oID.IntBytes()

	pID := ParseIntBytes(i)
	if pID != oID {
		t.Fatalf("pID %v != oID %v", pID, oID)
	}

	ms := [8]uint8{0xf, 0x7f, 0xc0, 0xfc, 0x2f, 0x80, 0x0, 0x0}
	mi := int64(1116823421972381696)
	pID = ParseIntBytes(ms)

	if pID.Int64() != mi {
		t.Fatalf("pID %v != mi %v", pID.Int64(), mi)
	}
}

//******************************************************************************
// Marshall Test Methods

func TestMarshalJSON(t *testing.T) {
	id := ID(13587)
	expected := "\"13587\""

	bytes, err := id.MarshalJSON()
	if err != nil {
		t.Fatalf("Unexpected error during MarshalJSON")
	}

	if string(bytes) != expected {
		t.Fatalf("Got %s, expected %s", string(bytes), expected)
	}
}

func TestMarshalsIntBytes(t *testing.T) {
	id := ID(13587).IntBytes()
	expected := []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x35, 0x13}

	if !bytes.Equal(id[:], expected) {
		t.Fatalf("Expected ID to be encoded as %v, got %v", expected, id)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	tt := []struct {
		json        string
		expectedID  ID
		expectedErr error
	}{
		{`"13587"`, 13587, nil},
		{`1`, 0, JSONSyntaxError{[]byte(`1`)}},
		{`"invalid`, 0, JSONSyntaxError{[]byte(`"invalid`)}},
	}

	for _, tc := range tt {
		var id ID
		err := id.UnmarshalJSON([]byte(tc.json))

		if !reflect.DeepEqual(err, tc.expectedErr) {
			t.Fatalf("Expected to get error '%s' decoding JSON, but got '%s'", tc.expectedErr, err)
		}

		if id != tc.expectedID {
			t.Fatalf("Expected to get ID '%s' decoding JSON, but got '%s'", tc.expectedID, id)
		}
	}
}

// ****************************************************************************
// Benchmark Methods

// nolint gomnd
func BenchmarkParseBase32(b *testing.B) {
	node, _ := NewNode(WithNodeID(1))
	sf := node.Next()
	b32i := sf.Base32()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		ParseBase32([]byte(b32i))
	}
}

func BenchmarkBase32(b *testing.B) {
	node, _ := NewNode(WithNodeID(1))
	sf := node.Next()

	b.ReportAllocs()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		sf.Base32()
	}
}
func BenchmarkParseBase58(b *testing.B) {
	node, _ := NewNode(WithNodeID(1))
	sf := node.Next()
	b58 := sf.Base58()

	b.ReportAllocs()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		ParseBase58([]byte(b58))
	}
}

func BenchmarkBase58(b *testing.B) {
	node, _ := NewNode(WithNodeID(1))
	sf := node.Next()

	b.ReportAllocs()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		sf.Base58()
	}
}
func BenchmarkGenerate(b *testing.B) {
	node, _ := NewNode(WithNodeID(1))

	b.ReportAllocs()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = node.Next()
	}
}

func BenchmarkGenerateMaxSequence(b *testing.B) {
	node, _ := NewNode(WithNodeID(1), WithNodeBits(1), WithStepBits(21))

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = node.Next()
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	// Next the ID to unmarshal
	node, _ := NewNode(WithNodeID(1))
	id := node.Next()
	bytes, _ := id.MarshalJSON()

	var id2 ID

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = id2.UnmarshalJSON(bytes)
	}
}

func BenchmarkMarshal(b *testing.B) {
	// Next the ID to marshal
	node, _ := NewNode(WithNodeID(1))
	id := node.Next()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_, _ = id.MarshalJSON()
	}
}
