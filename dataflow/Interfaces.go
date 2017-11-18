package dataflow

// something which provides values (-> a source)
type Drainable interface {
	Drain() <-chan Value
	Append(fillable Fillable) (Fillable)
}

// something which can consume values (-> a sink)
type Fillable interface {
	Fill(input <-chan Value)
}

// simething which does both (-> a pipeline stage)
type Pipelineable interface {
	Drainable
	Fillable
}
