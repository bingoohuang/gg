package ss

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"unicode"
)

func TestStripSpaces(t *testing.T) {
	assert.Equal(t, "abc", StripSpaces("a b\rc"))
	assert.Equal(t, "abc", Strip("\ta\u0020b\u3000c", unicode.IsSpace, Not(unicode.IsPrint)))
}
