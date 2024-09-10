package generatorDevice

import (
	"time"
)

const inStateUpdateInterval = 100 * time.Millisecond

type State int

const (
	Error State = iota
	Reset
	Off
	Ready
	Priming
	Cranking
	WarmUp
	Producing
	EngineCoolDown
	EnclosureCoolDown
)

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

type Configuration struct {
	PrimingTimeout           time.Duration
	CrankingTmeout           time.Duration
	WarmUpTimeout            time.Duration
	WarmUpTemp               float64
	EngineCoolDownTimeout    time.Duration
	EngineCoolDownTemp       float64
	EnclosureCoolDownTimeout time.Duration
	EnclosureCoolDownTemp    float64
	EngineTempMin            float64
	EngineTempMax            float64
	AirIntakeTempMin         float64
	AirIntakeTempMax         float64
	AirExhaustTempMin        float64
	AirExhaustTempMax        float64
	UMin                     float64
	UMax                     float64
	FMin                     float64
	FMax                     float64
	PMax                     float64
	PTotMax                  float64
}

type Inputs struct {
	// Switches
	CommandSwitch bool
	ResetSwitch   bool

	// I/O controller inputs
	IOAvailable    bool
	ArmSwitch      bool
	FireDetected   bool
	EngineTemp     float64
	AirIntakeTemp  float64
	AirExhaustTemp float64

	// Output measurement inputs
	MessurementAvailable bool
	U0                   float64
	U1                   float64
	U2                   float64
	L0                   float64
	L1                   float64
	L2                   float64
	F                    float64

	// Virtual inputs
	TimeInState time.Duration
}

type Outputs struct {
	Ignition bool
	Starter  bool
	Fan      bool
	Pump     bool
	Load     bool
}

type Controller struct {
	config  Configuration
	state   State
	inputs  Inputs
	outputs Outputs

	lastStateChange chan time.Time

	ChangeInput    chan func(Inputs) Inputs
	StateChanged   chan State
	InputsChanged  chan Inputs
	OutputsChanged chan Outputs
}

func NewController(config Configuration) *Controller {
	return &Controller{
		config:          config,
		state:           Off,
		lastStateChange: make(chan time.Time),
		ChangeInput:     make(chan func(Inputs) Inputs),
		StateChanged:    make(chan State),
		InputsChanged:   make(chan Inputs),
		OutputsChanged:  make(chan Outputs),
	}
}

func (c *Controller) Run() {
	// handle input updates
	go func() {
		defer close(c.lastStateChange)
		defer close(c.StateChanged)
		defer close(c.InputsChanged)
		defer close(c.OutputsChanged)

		c.StateChanged <- c.state
		c.InputsChanged <- c.inputs
		c.outputs = computeOutputs(c.state)
		c.OutputsChanged <- c.outputs

		for f := range c.ChangeInput {
			c.updateInputs(f)
		}
	}()

	// auto update TimeInState
	go func() {
		ticker := time.NewTicker(inStateUpdateInterval)
		defer ticker.Stop()

		var lastChange time.Time

		for {
			select {
			case t, ok := <-c.lastStateChange:
				if !ok {
					return
				}
				lastChange = t
			case <-ticker.C:
				c.ChangeInput <- func(i Inputs) Inputs {
					i.TimeInState = time.Since(lastChange)
					return i
				}
			}
		}
	}()
}

func (c *Controller) updateInputs(f func(Inputs) Inputs) {
	now := time.Now()

	// 1. update inputs
	newInputs := f(c.inputs)
	if newInputs == c.inputs {
		// no change: nothing to do
		return
	}
	c.inputs = newInputs
	c.InputsChanged <- newInputs

	// 2. compute new state
	newState := computeState(c.state, c.config, c.inputs)
	if newState != c.state {
		c.state = newState
		c.StateChanged <- newState
		c.lastStateChange <- now
		c.inputs.TimeInState = 0
	}

	// 3. compute new outputs
	newOutputs := computeOutputs(c.state)
	if newOutputs != c.outputs {
		c.outputs = newOutputs
		c.OutputsChanged <- newOutputs
	}
}

func computeState(prev State, c Configuration, i Inputs) (next State) {
	MasterSwitch := i.ArmSwitch && i.CommandSwitch

	// Mutli state transitions
	// in every case: reset switch triggers the reset state
	if i.ResetSwitch {
		return Reset
	}

	// in every state except reset, off and failed: a temperature or fire detection triggers the failed state
	if !(prev == Reset || prev == Off || prev == Error) &&
		!temperaturAndFireCheck(c, i) {
		return Error
	}

	// in warm up, producing and engine cool down: a negative output check triggers the failed state
	if (prev == WarmUp || prev == Producing || prev == EngineCoolDown) &&
		!generatorOutputCheck(c, i) {
		return Error
	}

	// Single state transitions
	switch prev {
	case Reset:
		if !i.ResetSwitch {
			return Off
		}
	case Off:
		if i.IOAvailable {
			return Ready
		}
	case Ready:
		if MasterSwitch {
			return Priming
		}
	case Priming:
		if !MasterSwitch {
			return Ready
		}
		if i.TimeInState >= c.PrimingTimeout {
			return Cranking
		}
	case Cranking:
		if !MasterSwitch {
			return Ready
		}
		if i.TimeInState >= c.CrankingTmeout {
			return Error
		}
		if generatorOutputCheck(c, i) {
			return WarmUp
		}
	case WarmUp:
		if !MasterSwitch {
			return EnclosureCoolDown
		}
		if i.TimeInState >= c.WarmUpTimeout || i.EngineTemp >= c.WarmUpTemp {
			return Producing
		}
	case Producing:
		if !MasterSwitch {
			return EngineCoolDown
		}
	case EngineCoolDown:
		if MasterSwitch {
			return Producing
		}
		if i.TimeInState >= c.EngineCoolDownTimeout || i.EngineTemp <= c.EngineCoolDownTemp {
			return EnclosureCoolDown
		}
	case EnclosureCoolDown:
		if MasterSwitch {
			return Priming
		}
		if i.TimeInState >= c.EnclosureCoolDownTimeout || i.EngineTemp <= c.EnclosureCoolDownTemp {
			return Ready
		}
	}

	return prev
}

func computeOutputs(s State) Outputs {
	o := Outputs{}

	o.Ignition = s == Cranking ||
		s == WarmUp ||
		s == Producing ||
		s == EngineCoolDown

	o.Starter = s == Cranking

	o.Fan = s == Priming ||
		s == Cranking ||
		s == WarmUp ||
		s == Producing ||
		s == EngineCoolDown ||
		s == EnclosureCoolDown

	o.Pump = s == Priming ||
		s == Cranking ||
		s == WarmUp ||
		s == Producing ||
		s == EngineCoolDown

	o.Load = s == Producing

	return o
}

func temperaturAndFireCheck(c Configuration, i Inputs) bool {
	return !i.FireDetected && i.IOAvailable &&
		i.EngineTemp >= c.EngineTempMin && i.EngineTemp <= c.EngineTempMax &&
		i.AirIntakeTemp >= c.AirIntakeTempMin && i.AirIntakeTemp <= c.AirIntakeTempMax &&
		i.AirExhaustTemp >= c.AirExhaustTempMin && i.AirExhaustTemp <= c.AirExhaustTempMax
}

func generatorOutputCheck(c Configuration, i Inputs) bool {
	return i.MessurementAvailable &&
		i.F >= c.FMin && i.F <= c.FMax &&
		i.U0 >= c.UMin && i.U0 <= c.UMax &&
		i.U1 >= c.UMin && i.U1 <= c.UMax &&
		i.U2 >= c.UMin && i.U2 <= c.UMax &&
		i.L0 <= c.PMax &&
		i.L1 <= c.PMax &&
		i.L2 <= c.PMax &&
		i.L0+i.L1+i.L2 <= c.PTotMax
}
