package queue

import (
	"errors"
	"sync"
)

var ErrFull = errors.New("fifo is full, value discarded")

// Fifo is a concurrency safe queue backed by a fixed-size ring buffer. Enqueue never blocks.
// Values are delivered in order via the channel returned by Dequeue.
type Fifo[T any] struct {
	lock sync.Mutex
	buf  []T // fixed size ring buffer
	head int // index of the oldest element
	size int // number of stored elements

	notify    chan struct{} // signals the routine that a value was added
	dequeue   chan T
	done      chan struct{}
	closeOnce sync.Once
}

// NewFifo creates a Fifo and starts its internal go routine.
// The Caller must Close it.
func NewFifo[T any](maxLength int) *Fifo[T] {
	q := &Fifo[T]{
		buf:     make([]T, maxLength),
		notify:  make(chan struct{}, 1),
		dequeue: make(chan T),
		done:    make(chan struct{}),
	}
	go q.run()
	return q
}

func (q *Fifo[T]) Close() {
	q.closeOnce.Do(func() { close(q.done) })
}

// Enqueue appends value or returns ErrFull. It never blocks.
func (q *Fifo[T]) Enqueue(value T) error {
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

// Dequeue returns a channel delivering values in order. It is closed by Close.
func (q *Fifo[T]) Dequeue() <-chan T {
	return q.dequeue
}

func (q *Fifo[T]) run() {
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
		select {
		case q.dequeue <- value:
			q.lock.Lock()
			var zero T
			q.buf[q.head] = zero // release reference for GC
			q.head = (q.head + 1) % len(q.buf)
			q.size--
			q.lock.Unlock()
		case <-q.done:
			return
		}
	}
}
