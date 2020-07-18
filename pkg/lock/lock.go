package lock

import "sync"

type RWLock struct {
	Mutex sync.RWMutex
}

func (rw *RWLock) Lock() func() {
	rw.Mutex.Lock()

	return func() { rw.Mutex.Unlock() }
}

func (rw *RWLock) RLock() func() {
	rw.Mutex.RLock()

	return func() { rw.Mutex.RUnlock() }
}
