package semaphore

import (
	"sync"
)

type Semaphore struct {
	Size   int
	Curr   int
	Finish bool
	Mu     sync.Mutex
	Cond   *sync.Cond
}

func NewSemaphore(capacity int) *Semaphore {
	sema := &Semaphore{Size: capacity, Curr: capacity, Finish: false}
	sema.Cond = sync.NewCond(&sema.Mu)
	return sema
}
func (s *Semaphore) Up() {
	s.Curr++
	// Signal another thread to start working
	s.Cond.Signal()
	s.Mu.Unlock()
}
func (s *Semaphore) Down() {
	s.Mu.Lock()
	// If thread capacity is reached, drop the lock and wait
	for s.Curr == 0 && !s.Finish {
		s.Cond.Wait()
	}
	s.Curr--
}
