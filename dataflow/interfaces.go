package dataflow

type Fillable interface {
	Fill(value Value)
}
