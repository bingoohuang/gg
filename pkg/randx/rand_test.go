package randx

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUint64(t *testing.T) {
	assert.True(t, Uint64() >= 0)

	for i := 0; i < 10; i++ {
		fmt.Print(Bool(), " ")
	}
	fmt.Println()
}
