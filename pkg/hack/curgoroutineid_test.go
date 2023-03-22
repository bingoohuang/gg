package hack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGid(t *testing.T) {
	assert.NotEmpty(t, CurGoroutineID())
}
