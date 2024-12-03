package genset

import "fmt"

func (i Inputs) String() string {
	return fmt.Sprintf(
		"Inputs{Time: %v, CommandSwitch: %v, ResetSwitch: %v, "+
			"IOAvailable: %v, ArmSwitch: %v, FireDetected: %v, EngineTemp: %v, AuxTemp0: %v, AuxTemp1: %v, "+
			"OutputAvailable: %v, U1: %v, U2: %v, U3: %v, L1: %v, L2: %v, L3: %v, F: %v}",
		i.Time, i.CommandSwitch, i.ResetSwitch,
		i.IOAvailable, i.ArmSwitch, i.FireDetected, i.EngineTemp, i.AuxTemp0, i.AuxTemp1,
		i.OutputAvailable, i.U1, i.U2, i.U3, i.L1, i.L2, i.L3, i.F,
	)
}

func (s State) String() string {
	return fmt.Sprintf("State{Node:\t%v,\tChanged: %v}", s.Node, s.Changed)
}

func (o Outputs) String() string {
	return fmt.Sprintf(
		"Outputs{Ignition: %v, Starter: %v, Fan: %v, Pump: %v, Load: %v "+
			"TimeInState: %v, IoCheck: %v, OutputCheck: %v}",
		o.Ignition, o.Starter, o.Fan, o.Pump, o.Load,
		o.TimeInState, o.IoCheck, o.OutputCheck,
	)
}
