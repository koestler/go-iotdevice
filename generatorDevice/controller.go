package generatorDevice

import (
	"fmt"
	"time"
)

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

type Combined struct {
	DerivedInputs DerivedInputs
	State         State
	Outputs       Outputs
}

type Controller struct {
	config        Configuration
	inputs        Inputs
	derivedInputs DerivedInputs
	state         State
	outputs       Outputs

	lastStateChange time.Time
	ticker          *time.Ticker

	ChangeInput chan func(Inputs) Inputs
	Update      chan Combined
}

func NewController(config Configuration) *Controller {
	if config.IOCheck == nil {
		panic("IOCheck is nil")
	}
	if config.OutputCheck == nil {
		panic("OutputCheck is nil")
	}
	return &Controller{
		config:      config,
		state:       Off,
		ChangeInput: make(chan func(Inputs) Inputs),
		Update:      make(chan Combined),
	}
}

func (c *Controller) Run() {
	go func() {
		defer close(c.Update)

		c.ticker = time.NewTicker(c.config.InStateResolution)
		defer c.ticker.Stop()

		// send initial state
		c.sendUpdate()

		for {
			select {
			case f, ok := <-c.ChangeInput:
				if !ok {
					// channel closed, terminate the controller
					return
				}
				// whenever an input is changed, recompute the derived inputs and the state
				if nextI := f(c.inputs); nextI != c.inputs {
					c.inputs = nextI
					if c.compute() {
						c.sendUpdate()
					}
				}
			case <-c.ticker.C:
				// whenever the ticker fires, recompute the derived inputs and the state
				if c.compute() {
					c.sendUpdate()
				}
			}
		}
	}()
}

func (c *Controller) compute() bool {
	fmt.Println("compute")
	var change bool
	if nextDI := computeDerivedInputs(c.config, c.inputs, c.lastStateChange); nextDI != c.derivedInputs {
		if nextState := computeState(c.state, c.config, c.inputs, c.derivedInputs); nextState != c.state {
			c.state = nextState
			c.lastStateChange = time.Now()
			c.ticker.Reset(c.config.InStateResolution)

			c.outputs = computeOutputs(c.state)
		}
		change = true
	}
	return change
}

func (c *Controller) sendUpdate() {
	fmt.Printf("sendUpdate: %v\n", c)
	c.Update <- Combined{
		DerivedInputs: c.derivedInputs,
		State:         c.state,
		Outputs:       c.outputs,
	}
}

func computeDerivedInputs(c Configuration, i Inputs, lastStateChange time.Time) DerivedInputs {
	fmt.Println("computeDerivedInputs")
	return DerivedInputs{
		MasterSwitch: i.ArmSwitch && i.CommandSwitch,
		IOCheck:      c.IOCheck(i),
		OutputCheck:  c.OutputCheck(i),
		TimeInState:  time.Since(lastStateChange).Truncate(c.InStateResolution),
	}
}

func computeState(prev State, c Configuration, i Inputs, di DerivedInputs) (next State) {
	fmt.Println("computeState")

	// Mutli state transitions
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
	fmt.Println("computeOutputs")

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
