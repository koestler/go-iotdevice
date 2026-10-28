package genset

type StateNode int

const (
	Error StateNode = iota
	Reset
	Off
	Ready
	Priming
	Cranking
	Stabilizing
	WarmUp
	Producing
	EngineCoolDown
	EnclosureCoolDown
)

func StateNodeMap() map[StateNode]string {
	return map[StateNode]string{
		Error:             "Error",
		Reset:             "Reset",
		Off:               "Off",
		Ready:             "Ready",
		Priming:           "Priming",
		Cranking:          "Cranking",
		Stabilizing:       "Stabilizing",
		WarmUp:            "WarmUp",
		Producing:         "Producing",
		EngineCoolDown:    "EngineCoolDown",
		EnclosureCoolDown: "EnclosureCoolDown",
	}
}

func (s StateNode) String() string {
	if v, ok := StateNodeMap()[s]; ok {
		return v
	}
	return "Unknown"
}
