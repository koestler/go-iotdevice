package queue

import (
	"errors"
	"sync"
)

var ErrFull = errors.New("fifo is full, value discarded")

// FIFO is a concurrency safe queue backed by a fixed-size ring buffer. Enqueue never blocks.
// Values are delivered in order via the channel returned by Dequeue.
// A Value is then removed Next(true) or kept in place Next(false).
// Next must be called exactly once per value and must never be called twice (it panics).
type FIFO[T any] struct {
	lock sync.Mutex
	buf  []T // fixed size ring buffer
	head int // index of the oldest element
	size int // number of stored elements

	notify    chan struct{} // signals the routine that a value was added
	dequeue   chan Element[T]
	done      chan struct{}
	closeOnce sync.Once
}

// Element is a dequeued value. The FIFO delivers nothing else until Next is called.
type Element[T any] struct {
	Value   T
	result  chan bool     // closed by Next so a second call panics
	settled chan struct{} // closed by run once a removal has been applied or the fifo is closed
}

// NewFIFO creates a FIFO and starts its internal go routine.
// The Caller must Close it.
func NewFIFO[T any](maxLength int) *FIFO[T] {
	q := &FIFO[T]{
		buf:     make([]T, maxLength),
		notify:  make(chan struct{}, 1), // buffered (1) to avoid ever blocking Enqueue
		dequeue: make(chan Element[T]),
		done:    make(chan struct{}),
	}
	go q.run()
	return q
}

func (q *FIFO[T]) Close() {
	q.closeOnce.Do(func() { close(q.done) })
}

// Enqueue appends value or returns ErrFull. It never blocks.
func (q *FIFO[T]) Enqueue(value T) error {
	q.lock.Lock()
	if q.size >= len(q.buf) {
		q.lock.Unlock()
		return ErrFull
	}
	q.buf[(q.head+q.size)%len(q.buf)] = value
	q.size++
	q.lock.Unlock()

	select {
	case q.notify <- struct{}{}:
	default: // a wake-up is already pending
	}
	return nil
}

// Dequeue returns a channel delivering elements in order. It is closed by Close.
func (q *FIFO[T]) Dequeue() <-chan Element[T] {
	return q.dequeue
}

// Next must be called exactly once per element and must never be called twice (it panics).
// With remove=true the element is deleted from the queue (processing succeeded) and Next returns
// once the slot is free again. With remove=false the element stays at the front of the queue
// (processing failed) and is delivered again.
func (e Element[T]) Next(remove bool) {
	defer close(e.result) // a second call panics here or on the send
	select {
	case e.result <- remove:
	case <-e.settled: // fifo was closed, nobody is listening anymore
		return
	}
	if remove {
		<-e.settled
	}
}

func (q *FIFO[T]) run() {
	defer close(q.dequeue)

	for {
		q.lock.Lock()
		empty := q.size == 0
		var value T
		if !empty {
			value = q.buf[q.head]
		}
		q.lock.Unlock()

		if empty {
			// queue is empty, wait for notify or done
			select {
			case <-q.notify:
				continue
			case <-q.done:
				return
			}
		}

		// value can be sent, wait for dequeue or done
		element := Element[T]{Value: value, result: make(chan bool), settled: make(chan struct{})}
		select {
		case q.dequeue <- element:
		case <-q.done:
			return
		}

		// wait for the receiver to decide
		select {
		case remove := <-element.result:
			if remove {
				q.lock.Lock()
				var zero T
				q.buf[q.head] = zero // release reference for GC
				q.head = (q.head + 1) % len(q.buf)
				q.size--
				q.lock.Unlock()
				close(element.settled)
			}
		case <-q.done:
			close(element.settled) // release a pending Remove
			return
		}
	}
}
