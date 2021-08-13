package any_tests

import (
	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_wrap_map(t *testing.T) {
	should := require.New(t)
	any := jsoni.Wrap(map[string]string{"Field1": "hello"})
	should.Equal("hello", any.Get("Field1").ToString())
	any = jsoni.Wrap(map[string]string{"Field1": "hello"})
	should.Equal(1, any.Size())
}

func Test_map_wrapper_any_get_all(t *testing.T) {
	should := require.New(t)
	any := jsoni.Wrap(map[string][]int{"Field1": {1, 2}})
	should.Equal(`{"Field1":1}`, any.Get('*', 0).ToString())
	should.Contains(any.Keys(), "Field1")

	// map write to
	stream := jsoni.NewStream(jsoni.ConfigDefault, nil, 0)
	any.WriteTo(stream)
	// TODO cannot pass
	//should.Equal(string(stream.buf), "")
}
