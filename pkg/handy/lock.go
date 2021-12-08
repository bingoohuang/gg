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
	l.Lock()
	return l.Unlock
}

// LockUnlock lock on mutex and return unlock.
// e.g. defer LockUnlock(mutex)
func LockUnlock(m *sync.Mutex) func() {
	m.Lock()
	return m.Unlock
}

type RWLock struct {
	sync.RWMutex
}

// WithLock run code within protection of lock.
func (l *RWLock) WithLock(f func()) {
	defer l.LockDeferUnlock()()
	f()
}

// WithRLock run code within protection of lock.
func (l *RWLock) WithRLock(f func()) {
	defer l.RLockDeferRUnlock()()
	f()
}

func (l *RWLock) TryLock() bool {
	state := (*int32)(unsafe.Pointer(&l.RWMutex))
	return atomic.CompareAndSwapInt32(state, 0, mutexLocked)
}

func (l *RWLock) LockDeferUnlock() func() {
	l.Lock()
	return l.Unlock
}

func (l *RWLock) RLockDeferRUnlock() func() {
	l.RLock()
	return l.RUnlock
}

// RWLockUnlock lock on mutex and return unlock.
// e.g. defer LockUnlock(mutex)
func RWLockUnlock(m *sync.RWMutex) func() {
	m.Lock()
	return m.Unlock
}

// RWRLockRUnlock lock on mutex and return unlock.
// e.g. defer LockUnlock(mutex)
func RWRLockRUnlock(m *sync.RWMutex) func() {
	m.RLock()
	return m.RUnlock
}
