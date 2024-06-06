// Package lock provides an implementation of a read-write lock
// that uses condition variables and mutexes.
package lock

import (
	"sync"
)

type RWMutex struct {
	mutex       sync.Mutex
	cond        *sync.Cond
	readerCount int
	maxReader   int
	writer      bool
}

func NewRWMutex(reader int) *RWMutex {
	// Create new RWMutex lock with maximum readers
	rw := &RWMutex{maxReader: reader}
	rw.cond = sync.NewCond(&rw.mutex)
	return rw
}

func (rw *RWMutex) Lock() {
	// Writer attempt to lock
	rw.mutex.Lock()
	// If there are readers reading or writer writing, wait and put down the lock
	for rw.readerCount > 0 || rw.writer {
		rw.cond.Wait()
	}
	// This writer's turn, set writer to true to ensure other threads know
	rw.writer = true
	// Writer unlock
	rw.mutex.Unlock()
}

func (rw *RWMutex) Unlock() {
	// Writer attempt to lock
	rw.mutex.Lock()
	// Writer finish, set writer to false
	rw.writer = false
	// Tell other threads waiting they can get process now
	rw.cond.Broadcast()
	// Writer unlock
	rw.mutex.Unlock()
}

func (rw *RWMutex) RLock() {
	// Reader attempt to lock
	rw.mutex.Lock()
	// If there is writer or reader count reached limit, wait and put down the lock
	for rw.writer || rw.readerCount == rw.maxReader {
		rw.cond.Wait()
	}
	// Reader gets in, add reader count
	rw.readerCount++
	// Unlock
	rw.mutex.Unlock()
}

func (rw *RWMutex) RUnlock() {
	// Reader attempt to lock
	rw.mutex.Lock()
	// Reader finish, decrement reader count
	rw.readerCount--
	// Release space for other threads
	rw.cond.Broadcast()
	// Reader unlock
	rw.mutex.Unlock()
}
