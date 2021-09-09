package ss

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToLowerKebab(t *testing.T) {
	assert.Equal(t, "abc", ToLowerKebab("ABC"))
	assert.Equal(t, "hello-world", ToLowerKebab("hello-world"))
	assert.Equal(t, "hello-world", ToLowerKebab("HelloWorld"))
	assert.Equal(t, "hello-url", ToLowerKebab("HelloURL"))
	assert.Equal(t, "hello-url-addr", ToLowerKebab("HelloURLAddr"))
}

func TestSplit(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, Split(",a,b,"))
	assert.Len(t, Split(""), 0)
}
