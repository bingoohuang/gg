package extra

import (
	"testing"

	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/stretchr/testify/require"
)

func init() {
	jsoni.RegisterExtension(&BinaryAsStringExtension{})
}

func TestBinaryAsStringCodec(t *testing.T) {
	t.Run("safe set", func(t *testing.T) {
		should := require.New(t)
		output, err := jsoni.Marshal([]byte("hello"))
		should.NoError(err)
		should.Equal(`"hello"`, string(output))
		var val []byte
		should.NoError(jsoni.Unmarshal(output, &val))
		should.Equal(`hello`, string(val))
	})
	t.Run("non safe set", func(t *testing.T) {
		should := require.New(t)
		output, err := jsoni.Marshal([]byte{1, 2, 3, 23})
		should.NoError(err)
		should.Equal(`"\\x01\\x02\\x03\\x17"`, string(output))
		var val []byte
		should.NoError(jsoni.Unmarshal(output, &val))
		should.Equal([]byte{1, 2, 3, 23}, val)
	})
}
