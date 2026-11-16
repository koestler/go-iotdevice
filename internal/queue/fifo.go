package queue

import "sync"

// FIFO is a concurrency safe queue backed by a fixed-size ring buffer. Enqueue never blocks.
// Values are delivered in order via the channel returned by Dequeue.
// A value is then either removed with Next(true) or kept in place with Next(false).
// An element that is being processed can still be replaced by Enqueue when the buffer is full;
// in that case Next has no effect on the queue.
type FIFO[T any] struct {
	lock sync.Mutex
	buf  []T    // fixed size ring buffer
	head int    // index of the oldest element
	size int    // number of stored elements
	pops uint64 // number of elements ever removed from the front

	notify  chan struct{} // signals the goroutine that a value was added
	dequeue chan Element[T]
	done    chan struct{}
}

// Element is a dequeued value. The FIFO delivers nothing else until Next is called.
type Element[T any] struct {
	Value   T
	result  chan bool     // closed by Next so a second call panics
	settled chan struct{} // closed by run once a removal has been applied or the fifo is closed
}

// NewFIFO creates a FIFO and starts its internal go routine.
// maxLength must be >= 1 (it panics otherwise). A full buffer replaces its oldest element on Enqueue.
// The caller must Close it.
func NewFIFO[T any](maxLength int) *FIFO[T] {
	if maxLength < 1 {
		panic("queue: maxLength must be >= 1")
	}
	q := &FIFO[T]{
		buf:     make([]T, maxLength),
		notify:  make(chan struct{}, 1), // buffered (1) to avoid ever blocking Enqueue
		dequeue: make(chan Element[T]),
		done:    make(chan struct{}),
	}
	go q.run()
	return q
}

// Close stops the internal go routine and closes the Dequeue channel.
// Close must be called exactly once; a second call panics. It does not wait for the goroutine to exit.
// A pending or later Next call returns without effect. Enqueue is not rejected after Close,
// but enqueued values are never delivered.
func (q *FIFO[T]) Close() {
	close(q.done)
}

// Enqueue appends value. When the buffer is full, the oldest element is replaced and true is returned.
// It returns false if there was capacity left. It never blocks.
func (q *FIFO[T]) Enqueue(value T) (replaced bool) {
	q.lock.Lock()
	replaced = q.size >= len(q.buf)
	if replaced {
		// the oldest slot is the next free one: overwrite it, even if it is currently being processed
		q.buf[q.head] = value
		q.head = (q.head + 1) % len(q.buf)
		q.pops++
	} else {
		q.buf[(q.head+q.size)%len(q.buf)] = value
		q.size++
	}
	q.lock.Unlock()

	select {
	case q.notify <- struct{}{}:
	default: // a wake-up is already pending
	}
	return replaced
}

// Dequeue returns a channel delivering elements in order. It is closed by Close.
func (q *FIFO[T]) Dequeue() <-chan Element[T] {
	return q.dequeue
}

// Next must be called exactly once per element; a second call panics.
// Until it is called, the FIFO delivers no further element; only Close releases a blocked queue.
// With remove=true the element is deleted from the queue (processing succeeded).
// With remove=false the element stays at the front of the queue (processing failed) and is delivered again.
// If Enqueue replaced the element in the meantime, neither has an effect on the queue.
func (e Element[T]) Next(remove bool) {
	defer close(e.result) // a second call panics on close(e.result) or on the send to the closed channel
	select {
	case e.result <- remove:
	case <-e.settled:
		return
	}
	<-e.settled
}

func (q *FIFO[T]) run() {
	defer close(q.dequeue)

	for {
		q.lock.Lock()
		empty := q.size == 0
		var value T
		var pops uint64
		if !empty {
			value = q.buf[q.head]
			pops = q.pops
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
		case <-q.notify:
			continue // the front element may have been replaced: reload it
		case <-q.done:
			return
		}

		// wait for the receiver to decide
		select {
		case remove := <-element.result:
			if remove {
				q.lock.Lock()
				if q.pops == pops { // otherwise the element was already replaced by Enqueue
					var zero T
					q.buf[q.head] = zero // release reference for GC
					q.head = (q.head + 1) % len(q.buf)
					q.size--
					q.pops++
				}
				q.lock.Unlock()
			}
			close(element.settled)
		case <-q.done:
			close(element.settled) // release a pending Next
			return
		}
	}
}
