package ctx

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// RegisterSignals registers signal handlers.
func RegisterSignals(c context.Context, signals ...os.Signal) context.Context {
	if c == nil {
		c = context.Background()
	}
	cc, cancel := context.WithCancel(c)
	sig := make(chan os.Signal, 1)
	if len(signals) == 0 {
		// syscall.SIGINT: ctl + c, syscall.SIGTERM: kill pid
		signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}
	signal.Notify(sig, signals...)
	go func() {
		<-sig
		cancel()
	}()

	return cc
}
