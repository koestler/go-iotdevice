package dataflow

import "context"

type ValueSubscription struct {
	ctx           context.Context
	outputChannel chan Value
	filter        ValueFilterFunc
}

func (s *ValueSubscription) Drain() <-chan Value {
	return s.outputChannel
}
