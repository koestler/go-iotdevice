package queue

import (
	"container/list"
)

type FifoQueue[T any] struct {
	maxLength     int
	containerList list.List
}

func NewFifoQueue[T any](maxLength int) FifoQueue[T] {
	return FifoQueue[T]{maxLength: maxLength}
}

func (q *FifoQueue[T]) Len() int {
	return q.containerList.Len()
}

func (q *FifoQueue[T]) Enqueue(value T) {
	q.containerList.PushBack(value)

	// when list gets to long; truncate first element
	if q.containerList.Len() > q.maxLength {
		elem := q.containerList.Front()
		if elem != nil {
			q.containerList.Remove(elem)
		}
	}
}

func (q *FifoQueue[T]) Dequeue() (value T, ok bool) {
	elem := q.containerList.Front()
	if elem == nil {
		var empty T
		return empty, false
	}

	q.containerList.Remove(elem)
	return elem.Value.(T), true
}
