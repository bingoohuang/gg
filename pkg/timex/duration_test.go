package timex

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDuration(t *testing.T) {
	d, err := ParseDuration("10w")
	assert.Nil(t, err)
	assert.Equal(t, 10*7*24*time.Hour, d)

	d, err = ParseDuration("10M")
	assert.Nil(t, err)
	assert.Equal(t, 10*30*24*time.Hour, d)
}
