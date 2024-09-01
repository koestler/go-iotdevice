package generatorDevice

type State int

const (
	Failed State = iota
	Reset
	Off
	Ready
	Cranking
	WarmUp
	Producing
	EngineCoolDown
	EnclosureCoolDown
)
