package generator

import (
	"time"
)

type Configuration struct {
	InStateResolution        time.Duration
	PrimingTimeout           time.Duration
	CrankingTmeout           time.Duration
	WarmUpTimeout            time.Duration
	WarmUpTemp               float64
	EngineCoolDownTimeout    time.Duration
	EngineCoolDownTemp       float64
	EnclosureCoolDownTimeout time.Duration
	EnclosureCoolDownTemp    float64
	IOCheck                  func(Inputs) bool
	OutputCheck              func(Inputs) bool
}

type State int

const InitialState = Off

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
	OutputAvailable bool
	U0              float64
	U1              float64
	U2              float64
	L0              float64
	L1              float64
	L2              float64
	F               float64
}

type DerivedInputs struct {
	MasterSwitch bool
	IOCheck      bool
	OutputCheck  bool
	TimeInState  time.Duration
}

type Outputs struct {
	Ignition bool
	Starter  bool
	Fan      bool
	Pump     bool
	Load     bool
}

type Change struct {
	f    func(Inputs) Inputs
	done chan struct{}
}

type Controller struct {
	config        Configuration
	inputs        Inputs
	derivedInputs DerivedInputs
	state         State
	outputs       Outputs

	lastStateChange time.Time
	ticker          *time.Ticker

	changeInputs        chan Change
	derivedInputsUpdate chan DerivedInputs
	stateUpdate         chan State
	outputUpdate        chan Outputs
}

func NewController(config Configuration) *Controller {
	if config.InStateResolution < 1*time.Millisecond {
		panic("InStateResolution is too low")
	}
	if config.IOCheck == nil {
		panic("IOCheck is nil")
	}
	if config.OutputCheck == nil {
		panic("OutputCheck is nil")
	}
	return &Controller{
		config:       config,
		state:        InitialState,
		changeInputs: make(chan Change),
	}
}

func (c *Controller) UpdateInputs(f func(Inputs) Inputs) {
	c.changeInputs <- Change{f: f}
}

func (c *Controller) UpdateInputsSync(f func(Inputs) Inputs) {
	done := make(chan struct{})
	c.changeInputs <- Change{f: f, done: done}
	<-done
}

func (c *Controller) DerivedInputs() <-chan DerivedInputs {
	if c.derivedInputsUpdate == nil {
		c.derivedInputsUpdate = make(chan DerivedInputs)
		return c.derivedInputsUpdate
	}
	panic("DerivedInputs channel already in use")
}

func (c *Controller) State() <-chan State {
	if c.stateUpdate == nil {
		c.stateUpdate = make(chan State)
		return c.stateUpdate
	}
	panic("State channel already in use")
}

func (c *Controller) Outputs() <-chan Outputs {
	if c.outputUpdate == nil {
		c.outputUpdate = make(chan Outputs)
		return c.outputUpdate
	}
	panic("Outputs channel already in use")
}

func (c *Controller) Run() {
	initDone := make(chan struct{})

	go func() {
		defer func() {
			if c.derivedInputsUpdate != nil {
				close(c.derivedInputsUpdate)
			}
			if c.stateUpdate != nil {
				close(c.stateUpdate)
			}
			if c.outputUpdate != nil {
				close(c.outputUpdate)
			}
		}()

		c.ticker = time.NewTicker(c.config.InStateResolution)
		defer c.ticker.Stop()

		c.computeInitial()

		close(initDone)

		for {
			select {
			case change, ok := <-c.changeInputs:
				if !ok {
					// channel closed, terminate the controller
					return
				}
				// whenever an input is changed, recompute the derived inputs and the state
				if nextI := change.f(c.inputs); nextI != c.inputs {
					c.inputs = nextI
					c.compute()
				}
				if change.done != nil {
					close(change.done)
				}
			case <-c.ticker.C:
				// whenever the ticker fires, recompute the derived inputs and the state
				c.compute()
			}
		}
	}()

	<-initDone
}

func (c *Controller) End() {
	if c.changeInputs != nil {
		close(c.changeInputs)
	}
}

func (c *Controller) computeInitial() {
	c.derivedInputs = computeDerivedInputs(c.config, c.inputs, c.lastStateChange)
	c.state = computeState(c.state, c.config, c.inputs, c.derivedInputs)
	c.outputs = computeOutputs(c.state)
	c.lastStateChange = time.Now()

	if c.derivedInputsUpdate != nil {
		c.derivedInputsUpdate <- c.derivedInputs
	}
	if c.stateUpdate != nil {
		c.stateUpdate <- c.state
	}
	if c.outputUpdate != nil {
		c.outputUpdate <- c.outputs
	}
}

func (c *Controller) compute() {
	if nextDI := computeDerivedInputs(c.config, c.inputs, c.lastStateChange); nextDI != c.derivedInputs {
		c.derivedInputs = nextDI
		if nextState := computeState(c.state, c.config, c.inputs, c.derivedInputs); nextState != c.state {
			c.state = nextState
			c.lastStateChange = time.Now()
			c.ticker.Reset(c.config.InStateResolution)
			if nextOutputs := computeOutputs(c.state); nextOutputs != c.outputs {
				c.outputs = nextOutputs
				if c.outputUpdate != nil {
					c.outputUpdate <- c.outputs
				}
			}
			if c.stateUpdate != nil {
				c.stateUpdate <- c.state
			}
		}
		if c.derivedInputsUpdate != nil {
			c.derivedInputsUpdate <- nextDI
		}
	}
}

func computeDerivedInputs(c Configuration, i Inputs, lastStateChange time.Time) DerivedInputs {
	return DerivedInputs{
		MasterSwitch: i.ArmSwitch && i.CommandSwitch,
		IOCheck:      c.IOCheck(i),
		OutputCheck:  c.OutputCheck(i),
		TimeInState:  time.Since(lastStateChange).Truncate(c.InStateResolution),
	}
}

func computeState(prev State, c Configuration, i Inputs, di DerivedInputs) (next State) {
	// Multi state transitions
	// in every case: reset switch triggers the reset state
	if i.ResetSwitch {
		return Reset
	}

	// in every state except reset, off and failed: a temperature or fire detection triggers the failed state
	if !(prev == Reset || prev == Off || prev == Error) && !di.IOCheck {
		return Error
	}

	// in warm up, producing and engine cool down: a negative output check triggers the failed state
	if (prev == WarmUp || prev == Producing || prev == EngineCoolDown) && !di.OutputCheck {
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
		if di.MasterSwitch {
			return Priming
		}
	case Priming:
		if !di.MasterSwitch {
			return Ready
		}
		if di.TimeInState >= c.PrimingTimeout {
			return Cranking
		}
	case Cranking:
		if !di.MasterSwitch {
			return Ready
		}
		if di.TimeInState >= c.CrankingTmeout {
			return Error
		}
		if di.OutputCheck {
			return WarmUp
		}
	case WarmUp:
		if !di.MasterSwitch {
			return EnclosureCoolDown
		}
		if di.TimeInState >= c.WarmUpTimeout || i.EngineTemp >= c.WarmUpTemp {
			return Producing
		}
	case Producing:
		if !di.MasterSwitch {
			return EngineCoolDown
		}
	case EngineCoolDown:
		if di.MasterSwitch {
			return Producing
		}
		if di.TimeInState >= c.EngineCoolDownTimeout || i.EngineTemp <= c.EngineCoolDownTemp {
			return EnclosureCoolDown
		}
	case EnclosureCoolDown:
		if di.MasterSwitch {
			return Priming
		}
		if di.TimeInState >= c.EnclosureCoolDownTimeout || i.EngineTemp <= c.EnclosureCoolDownTemp {
			return Ready
		}
	}

	return prev
}

func computeOutputs(s State) Outputs {
	return Outputs{
		Ignition: s == Cranking ||
			s == WarmUp ||
			s == Producing ||
			s == EngineCoolDown,
		Starter: s == Cranking,
		Fan: s == Priming ||
			s == Cranking ||
			s == WarmUp ||
			s == Producing ||
			s == EngineCoolDown ||
			s == EnclosureCoolDown,
		Pump: s == Priming ||
			s == Cranking ||
			s == WarmUp ||
			s == Producing ||
			s == EngineCoolDown,
		Load: s == Producing,
	}
}
