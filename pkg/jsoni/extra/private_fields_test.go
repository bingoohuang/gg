package extra

import (
	"testing"

	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/stretchr/testify/require"
)

func Test_private_fields(t *testing.T) {
	type TestObject struct {
		field1 string
	}
	SupportPrivateFields()
	should := require.New(t)
	obj := TestObject{}
	should.Nil(jsoni.UnmarshalFromString(`{"field1":"Hello"}`, &obj))
	should.Equal("Hello", obj.field1)
}
