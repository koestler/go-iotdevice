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

func NewFifo[T any](maxLength int) Fifo[T] {
	return Fifo[T]{maxLength: maxLength}
}

func (q *Fifo[T]) Enqueue(value T) {
	if q.maxLength < 1 {
		// when maxLength is zero or lower, store nothing
		return
	}

	q.lock.Lock()
	defer q.lock.Unlock()

	q.list.PushBack(value)

	// when list gets to long; truncate first element
	if q.list.Len() > q.maxLength {
		if elem := q.list.Front(); elem != nil {
			q.list.Remove(elem)
		}
	}
}

func (q *Fifo[T]) Dequeue() (value T, ok bool) {
	q.lock.Lock()
	defer q.lock.Unlock()

	elem := q.list.Front()
	if elem == nil {
		var empty T
		return empty, false
	}

	v := q.list.Remove(elem)
	return v, true
}
