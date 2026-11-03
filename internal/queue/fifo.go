package queue

import (
	"github.com/koestler/go-list"
	"sync"
)

type Fifo[T any] struct {
	maxLength int
	list      list.List[T]
	lock      sync.Mutex
}

func NewFifo[T any](maxLength int) *Fifo[T] {
	return &Fifo[T]{maxLength: maxLength}
}

func (q *Fifo[T]) Enqueue(value T) {
	if q.maxLength < 1 {
		// when maxLength is zero or lower, store nothing
		return
	}

	q.lock.Lock()
	defer q.lock.Unlock()

	q.list.PushBack(value)

	// when list gets too long, truncate the first element
	if q.list.Len() > q.maxLength {
		q.list.Remove(q.list.Front())
	}
}

func (q *Fifo[T]) Dequeue() (value T, ok bool) {
	q.lock.Lock()
	defer q.lock.Unlock()

	elem := q.list.Front()
	if elem == nil {
		return value, false
	}

	v := q.list.Remove(elem)
	return v, true
}
