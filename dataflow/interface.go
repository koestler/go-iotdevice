package dataflow

type Fillable interface {
	Fill(value Value)
}

type FilterFunc func(Value) bool
