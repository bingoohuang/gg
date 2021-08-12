package snow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNext(t *testing.T) {
	that := assert.New(t)
	that.True(Next() > 0)
}
