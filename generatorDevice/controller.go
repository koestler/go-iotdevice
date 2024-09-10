package generatorDevice

import (
	"context"
	"sync"
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

type ChangeCallback func(State, Inputs, Outputs)

type Controller struct {
	lock            sync.RWMutex
	config          Configuration
	state           State
	inputs          Inputs
	outputs         Outputs
	lastStateChange time.Time
	onChange        ChangeCallback
}

func NewController(config Configuration, onChange ChangeCallback) *Controller {
	if onChange == nil {
		panic("onChange must be provided")
	}
	return &Controller{
		config:   config,
		state:    Off,
		onChange: onChange,
	}
}

func (c *Controller) Run(ctx context.Context) {
	go c.updateInState(ctx)
}

func (c *Controller) updateInState(ctx context.Context) {
	ticker := time.NewTicker(inStateUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.UpdateInputs(func(i Inputs) Inputs {
				i.TimeInState = time.Since(c.lastStateChange)
				return i
			})
		}
	}
}

func (c *Controller) UpdateInputs(f func(Inputs) Inputs) {
	c.lock.Lock()
	defer c.lock.Unlock()

	now := time.Now()

	// 1. update inputs
	newInputs := f(c.inputs)
	if newInputs == c.inputs {
		return
	}
	c.inputs = newInputs

	// 2. compute new state
	newState := computeState(c.state, c.config, c.inputs)
	if newState != c.state {
		c.state = newState
		c.lastStateChange = now
		c.inputs.TimeInState = 0
	}

	// 3. compute new outputs
	c.outputs = computeOutputs(c.state)

	// 4. notify change
	c.onChange(c.state, c.inputs, c.outputs)
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
