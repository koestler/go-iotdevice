package dataflow

import "context"

type FilterFunc func(Value) bool

type Subscription struct {
	ctx           context.Context
	outputChannel chan Value
	filter        FilterFunc
}

func (s *Subscription) Drain() <-chan Value {
	return s.outputChannel
}
