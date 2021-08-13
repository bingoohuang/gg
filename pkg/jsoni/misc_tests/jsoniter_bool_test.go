package misc_tests

import (
	"bytes"
	"testing"

	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/stretchr/testify/require"
)

func Test_true(t *testing.T) {
	should := require.New(t)
	iter := jsoni.ParseString(jsoni.ConfigDefault, `true`)
	should.True(iter.ReadBool())
	iter = jsoni.ParseString(jsoni.ConfigDefault, `true`)
	should.Equal(true, iter.Read())
}

func Test_false(t *testing.T) {
	should := require.New(t)
	iter := jsoni.ParseString(jsoni.ConfigDefault, `false`)
	should.False(iter.ReadBool())
}

func Test_write_true_false(t *testing.T) {
	should := require.New(t)
	buf := &bytes.Buffer{}
	stream := jsoni.NewStream(jsoni.ConfigDefault, buf, 4096)
	stream.WriteTrue()
	stream.WriteFalse()
	stream.WriteBool(false)
	stream.Flush()
	should.Nil(stream.Error)
	should.Equal("truefalsefalse", buf.String())
}

func Test_write_val_bool(t *testing.T) {
	should := require.New(t)
	buf := &bytes.Buffer{}
	stream := jsoni.NewStream(jsoni.ConfigDefault, buf, 4096)
	stream.WriteVal(true)
	should.Equal(stream.Buffered(), 4)
	stream.Flush()
	should.Equal(stream.Buffered(), 0)
	should.Nil(stream.Error)
	should.Equal("true", buf.String())
}
