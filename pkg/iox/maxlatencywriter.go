package iox

import (
	"io"
	"log"
	"sync"
	"time"
)

// from https://golang.org/src/net/http/httputil/reverseproxy.go

type WriteFlusher interface {
	io.Writer
	Flush() error
}

type MaxLatencyWriter struct {
	Dst     WriteFlusher
	Latency time.Duration // non-zero; negative means to flush immediately

	mu           sync.Mutex // protects t, flushPending, and dst.Flush
	t            *time.Timer
	flushPending bool
}

func (m *MaxLatencyWriter) Write(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, err = m.Dst.Write(p)
	if m.Latency < 0 {
		if err := m.Dst.Flush(); err != nil {
			return 0, err
		}
		return
	}
	if m.flushPending {
		return
	}
	if m.t == nil {
		m.t = time.AfterFunc(m.Latency, m.delayedFlush)
	} else {
		m.t.Reset(m.Latency)
	}
	m.flushPending = true
	return
}

func (m *MaxLatencyWriter) delayedFlush() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.flushPending { // if stop was called but AfterFunc already started this goroutine
		return
	}
	if err := m.Dst.Flush(); err != nil {
		log.Printf("flush failed: %v", err)
	}
	m.flushPending = false
}

func (m *MaxLatencyWriter) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.flushPending = false
	if m.t != nil {
		m.t.Stop()
	}
}
