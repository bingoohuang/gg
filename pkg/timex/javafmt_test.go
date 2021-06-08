package timex

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func TestFormat(t *testing.T) {
	s := FormatTime(time.Now(), "06yyyy-MM-dd")
	assert.True(t, strings.HasPrefix(s, "06"))
}
