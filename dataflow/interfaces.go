package dataflow

// something which provides values (-> a source)
type Drainable interface {
	Drain() <-chan Value
	Append(fillable Fillable) Fillable
}

// something which can consume values (-> a sink)
type Fillable interface {
	Fill(value Value)
}
