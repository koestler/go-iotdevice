package queue

import (
	"container/list"
	"sync"
)

type FifoQueue[T any] struct {
	maxLength     int
	containerList list.List
	lock          sync.Mutex
}

func NewFifoQueue[T any](maxLength int) FifoQueue[T] {
	return FifoQueue[T]{maxLength: maxLength}
}

func (q *FifoQueue[T]) Enqueue(value T) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.containerList.PushBack(value)

	// when list gets to long; truncate first element
	if q.containerList.Len() > q.maxLength {
		if elem := q.containerList.Front(); elem != nil {
			q.containerList.Remove(elem)
		}
	}
}

func (q *FifoQueue[T]) Dequeue() (value T, ok bool) {
	q.lock.Lock()
	defer q.lock.Unlock()

	elem := q.containerList.Front()
	if elem == nil {
		var empty T
		return empty, false
	}

	q.containerList.Remove(elem)
	return elem.Value.(T), true
}
