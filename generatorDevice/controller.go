package generatorDevice

import (
	"context"
	"sync"
	"time"
)

const inStateUpdateInterval = 100 * time.Millisecond

type State int

const (
	Failed State = iota
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
	WarmUpTimeout            time.Duration
	WarmUpTemp               float64
	EngineCoolDownTimeout    time.Duration
	EngineCoolDownTemp       float64
	EnclosureCoolDownTimeout time.Duration
	EnclosureCoolDownTemp    float64
	UMin                     float64
	UMax                     float64
	FMin                     float64
	FMax                     float64
	PMax                     float64
}

type Inputs struct {
	// Switches
	MasterSwitch bool
	ResetSwitch  bool

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
	newState := computeState(c.state, c.inputs)
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

func computeState(prev State, inp Inputs) (next State) {
	if inp.ResetSwitch {
		return Reset
	}

	switch prev {
	case Reset:
		if !inp.ResetSwitch {
			return Off
		}
	case Off:
		if inp.IOAvailable {
			return Ready
		}
	case Ready:
	case Cranking:
	case WarmUp:
	case Producing:
	case EngineCoolDown:
	case EnclosureCoolDown:
	}

	return prev
}

func computeOutputs(s State) Outputs {
	o := Outputs{}

	o.Ignition = (s == Cranking ||
		s == WarmUp ||
		s == Producing ||
		s == EngineCoolDown)

	o.Starter = (s == Cranking)

	o.Fan = (s == Cranking ||
		s == WarmUp ||
		s == Producing ||
		s == EngineCoolDown ||
		s == EnclosureCoolDown)

	o.Pump = (s == Priming ||
		s == Cranking ||
		s == WarmUp ||
		s == Producing ||
		s == EngineCoolDown ||
		s == EnclosureCoolDown)

	o.Load = (s == Producing)

	return o
}
