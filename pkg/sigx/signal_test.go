package sigx

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRegisterSignals(t *testing.T) {
	c, _ := RegisterSignals(nil)
	proc, _ := os.FindProcess(os.Getpid())
	if err := proc.Signal(os.Interrupt); err != nil {
		t.Failed()
	}

	done := false
	select {
	case <-c.Done():
		done = true
	case <-time.After(1 * time.Millisecond):
	}

	assert.True(t, done)
}
