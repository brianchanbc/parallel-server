package queue

import (
	"sync/atomic"
	"unsafe"
)

type Request struct {
	Command   string  `json:"command"`
	Id        int64   `json:"id"`
	Body      string  `json:"body"`
	Timestamp float64 `json:"timestamp"`
}

type node struct {
	req  *Request
	next unsafe.Pointer
}

// Golden source of pseudocode: https://www.cs.rochester.edu/research/synchronization/pseudocode/queues.html
// Go implementation: https://github.com/golang-design/lockfree/blob/master/queue.go

// LockfreeQueue represents a FIFO structure with operations to enqueue
// and dequeue tasks represented as Request
type LockFreeQueue struct {
	head   unsafe.Pointer // Use unsafe pointer so that go atomic operations can be used
	tail   unsafe.Pointer
	length uint64
}

// NewQueue creates and initializes a LockFreeQueue
func NewLockFreeQueue() *LockFreeQueue {
	head := node{req: nil, next: nil}
	lockFreeQueue := LockFreeQueue{head: unsafe.Pointer(&head), tail: unsafe.Pointer(&head)}
	return &lockFreeQueue
}

// Enqueue adds a series of Request to the queue
func (queue *LockFreeQueue) Enqueue(task *Request) {
	// Create new node
	newNode := &node{req: task, next: nil}
	var last, lastNext *node
	for {
		// Load the tail and next node after tail
		last = atomicLoad(&queue.tail)
		lastNext = atomicLoad(&last.next)
		// Check if tail is the still the last node
		// If not, continue to check again
		if atomicLoad(&queue.tail) == last {
			if lastNext == nil {
				// If tail is really the last node
				// Try to link the next node after last node be newNode
				if cas(&last.next, lastNext, newNode) {
					// If this succeeded, swing the new tail to the newNode
					cas(&queue.tail, last, newNode)
					// Atomically add 1 to the length
					atomic.AddUint64(&queue.length, 1)
					// Enqueue completes and return
					return
				}
			} else {
				// If next node after tail has something (non-nil),
				// Some other thread(s) enqueued before this thread
				// Try to reposition by swinging the tail to the next node
				cas(&queue.tail, last, lastNext)
			}
		}
	}
}

// Dequeue removes a Request from the queue
func (queue *LockFreeQueue) Dequeue() *Request {
	var first, firstNext, last *node
	for {
		// Atomically load first node, last node, node after first node
		first = atomicLoad(&queue.head)
		last = atomicLoad(&queue.tail)
		firstNext = atomicLoad(&first.next)
		// Check if first node is the head
		if first == atomicLoad(&queue.head) {
			// If first node is the head, check if first node is the last node
			if first == last {
				// If first node == last node and first next is nil, the queue
				// is at the same state as it was initiated (i.e. empty queue)
				// Return nil
				if firstNext == nil {
					return nil
				}
				// Now first == last but firstNext has something
				// that means tail is failing behind, spin to advance
				// tail to the next node (i.e. firstNext)
				cas(&queue.tail, last, firstNext)
			} else {
				// Read the value before cas in case another thread
				// dequeue free the next node
				req := firstNext.req
				// Try to swing the head to the next node
				// If succeed, we have successfully dequeued
				if cas(&queue.head, first, firstNext) {
					// Deduct 1 to the queue length
					atomic.AddUint64(&queue.length, ^uint64(0))
					// Return request
					return req
				}
			}
		}
	}
}

// Atomically load the length of the queue
func (queue *LockFreeQueue) Length() uint64 {
	return atomic.LoadUint64(&queue.length)
}

// Atomically load the pointer
func atomicLoad(ptr *unsafe.Pointer) *node {
	return (*node)(atomic.LoadPointer(ptr))
}

// Compare-and-swap operation with pointer
func cas(ptr *unsafe.Pointer, old, new *node) bool {
	return atomic.CompareAndSwapPointer(ptr, unsafe.Pointer(old), unsafe.Pointer(new))
}
