package bytex

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToInt64(t *testing.T) {
	b := FromUint64(100)
	assert.Equal(t, uint64(100), ToUint64(b))
}
