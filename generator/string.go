package generator

import "fmt"

func (i Inputs) String() string {
	return fmt.Sprintf(
		`Inputs{CommandSwitch: %v, ResetSwitch: %v,
     IOAvailable: %v, ArmSwitch: %v, FireDetected: %v, EngineTemp: %v, AirIntakeTemp: %v, AirExhaustTemp: %v,
     OutputAvailable: %v, U0: %v, U1: %v, U2: %v, L0: %v, L1: %v, L2: %v, F: %v} `,
		i.CommandSwitch, i.ResetSwitch,
		i.IOAvailable, i.ArmSwitch, i.FireDetected, i.EngineTemp, i.AirIntakeTemp, i.AirExhaustTemp,
		i.OutputAvailable, i.U0, i.U1, i.U2, i.L0, i.L1, i.L2, i.F,
	)
}

func (di DerivedInputs) String() string {
	return fmt.Sprintf(
		"DerivedInputs{MasterSwitch: %v, IOCheck: %v, OutputCheck: %v, TimeInState: %v}",
		di.MasterSwitch, di.IOCheck, di.OutputCheck, di.TimeInState,
	)
}

func (s State) String() string {
	switch s {
	case Error:
		return "Error"
	case Reset:
		return "Reset"
	case Off:
		return "Off"
	case Ready:
		return "Ready"
	case Priming:
		return "Priming"
	case Cranking:
		return "Cranking"
	case WarmUp:
		return "WarmUp"
	case Producing:
		return "Producing"
	case EngineCoolDown:
		return "EngineCoolDown"
	case EnclosureCoolDown:
		return "EnclosureCoolDown"
	}
	return "Unknown"
}

func (o Outputs) String() string {
	return fmt.Sprintf(
		"Outputs{Ignition: %v, Starter: %v, Fan: %v, Pump: %v, Load: %v}",
		o.Ignition, o.Starter, o.Fan, o.Pump, o.Load,
	)
}
