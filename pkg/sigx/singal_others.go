//go:build !windows
// +build !windows

package sigx

import (
	"os"
	"syscall"
)

var defaultSignals = []os.Signal{syscall.SIGUSR1}
