package handy

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

const mutexLocked = 1 << iota // mutex is locked

type Lock struct {
	sync.Mutex
}

// WithLock run code within protection of lock.
func (l *Lock) WithLock(f func()) {
	defer l.LockDeferUnlock()()
	f()
}

func (l *Lock) TryLock() bool {
	state := (*int32)(unsafe.Pointer(&l.Mutex))
	return atomic.CompareAndSwapInt32(state, 0, mutexLocked)
}

func (l *Lock) LockDeferUnlock() func() {
	l.Mutex.Lock()
	return l.Mutex.Unlock
}
