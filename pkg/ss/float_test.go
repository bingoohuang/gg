package ss

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFormatFloatOption(t *testing.T) {
	assert.Equal(t, "1", FormatFloatOption(1.23, WithPrec(0)))
	assert.Equal(t, "1", FormatFloatOption(1.01, WithPrec(1), WithRemoveTrailingZeros(true)))
	assert.Equal(t, "1.01", FormatFloatOption(1.0123, WithPrec(2), WithRemoveTrailingZeros(true)))
	assert.Equal(t, "1.01", FormatFloatOption(1.012001, WithPrec(2), WithRemoveTrailingZeros(true)))
}
