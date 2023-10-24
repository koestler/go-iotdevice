package dataflow

type Fillable interface {
	Fill(value Value)
}

type ValueFilterFunc func(Value) bool
type RegisterFilterFunc func(Register) bool

type RegisterFilterConf interface {
	IncludeRegisters() []string
	SkipRegisters() []string
	IncludeCategories() []string
	SkipCategories() []string
	DefaultInclude() bool
}
