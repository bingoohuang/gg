package misc_tests

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/stretchr/testify/require"
)

func Test_read_null(t *testing.T) {
	should := require.New(t)
	iter := jsoni.ParseString(jsoni.ConfigDefault, `null`)
	should.True(iter.ReadNil())
	iter = jsoni.ParseString(jsoni.ConfigDefault, `null`)
	ctx := context.Background()
	should.Nil(iter.Read(ctx))
	iter = jsoni.ParseString(jsoni.ConfigDefault, `navy`)
	iter.Read(ctx)
	should.True(iter.Error != nil && iter.Error != io.EOF)
	iter = jsoni.ParseString(jsoni.ConfigDefault, `navy`)
	iter.ReadNil()
	should.True(iter.Error != nil && iter.Error != io.EOF)
}

func Test_write_null(t *testing.T) {
	should := require.New(t)
	buf := &bytes.Buffer{}
	stream := jsoni.NewStream(jsoni.ConfigDefault, buf, 4096)
	stream.WriteNil()
	stream.Flush()
	should.Nil(stream.Error)
	should.Equal("null", buf.String())
}

func Test_decode_null_object_field(t *testing.T) {
	should := require.New(t)
	iter := jsoni.ParseString(jsoni.ConfigDefault, `[null,"a"]`)
	iter.ReadArray()
	if iter.ReadObject() != "" {
		t.FailNow()
	}
	iter.ReadArray()
	if iter.ReadString() != "a" {
		t.FailNow()
	}
	type TestObject struct {
		Field string
	}
	objs := []TestObject{}
	should.Nil(jsoni.UnmarshalFromString("[null]", &objs))
	should.Len(objs, 1)
}

func Test_decode_null_array_element(t *testing.T) {
	should := require.New(t)
	iter := jsoni.ParseString(jsoni.ConfigDefault, `[null,"a"]`)
	should.True(iter.ReadArray())
	should.True(iter.ReadNil())
	should.True(iter.ReadArray())
	should.Equal("a", iter.ReadString())
}

func Test_decode_null_string(t *testing.T) {
	should := require.New(t)
	iter := jsoni.ParseString(jsoni.ConfigDefault, `[null,"a"]`)
	should.True(iter.ReadArray())
	should.Equal("", iter.ReadString())
	should.True(iter.ReadArray())
	should.Equal("a", iter.ReadString())
}

func Test_decode_null_skip(t *testing.T) {
	iter := jsoni.ParseString(jsoni.ConfigDefault, `[null,"a"]`)
	iter.ReadArray()
	iter.Skip()
	iter.ReadArray()
	if iter.ReadString() != "a" {
		t.FailNow()
	}
}
