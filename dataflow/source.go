package dataflow

type Source struct {
	outputChain chan Value
}

func CreateSource(output chan Value) *Source {
	return &Source{
		outputChain: output,
	}
}

func (source *Source) Drain() <-chan Value {
	return source.outputChain
}

func (source *Source) Append(fillable Fillable) Fillable {
	fillable.Fill(source.Drain())
	return fillable
}
