package hack

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGid(t *testing.T) {
	assert.NotEmpty(t, CurGoroutineID())
}
