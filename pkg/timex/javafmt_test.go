package timex

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func TestFormat(t *testing.T) {
	s := Format("06yyyy-MM-dd", time.Now())
	assert.True(t, strings.HasPrefix(s, "06"))
}
