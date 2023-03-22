package timex

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormat(t *testing.T) {
	s := FormatTime(time.Now(), "06yyyy-MM-dd")
	assert.True(t, strings.HasPrefix(s, "06"))
}
