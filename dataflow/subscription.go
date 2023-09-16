package dataflow

import "context"

type FilterFunc func(Value) bool

type ValueSubscription struct {
	ctx           context.Context
	outputChannel chan Value
	filter        FilterFunc
}

func (s *ValueSubscription) Drain() <-chan Value {
	return s.outputChannel
}
