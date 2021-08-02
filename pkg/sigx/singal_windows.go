package sigx

import (
	"os"
	"syscall"
)

var defaultSignals = []os.Signal{syscall.SIGALRM}
