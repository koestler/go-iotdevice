package genset

import (
	"errors"
	"time"
)

// Params contains the parameters for the genset controller.
type Params struct {
	// Transition params
	PrimingTimeout           time.Duration
	CrankingTimeout          time.Duration
	StabilizingTimeout       time.Duration
	WarmUpTimeout            time.Duration
	WarmUpMinTime            time.Duration
	WarmUpTemp               float64
	EngineCoolDownTimeout    time.Duration
	EngineCoolDownMinTime    time.Duration
	EngineCoolDownTemp       float64
	EnclosureCoolDownTimeout time.Duration
	EnclosureCoolDownMinTime time.Duration
	EnclosureCoolDownTemp    float64

	// IO Check
	EngineTempMin float64
	EngineTempMax float64
	AuxTemp0Min   float64
	AuxTemp0Max   float64
	AuxTemp1Min   float64
	AuxTemp1Max   float64

	// Output Check
	SinglePhase bool
	UMin        float64
	UMax        float64
	FMin        float64
	FMax        float64
	PMax        float64
	PTotMax     float64
}

type Inputs struct {
	// Time is an input to the controller to allow for time-based state transitions
	Time time.Time

	// Switches
	ArmSwitch     bool
	CommandSwitch bool
	ResetSwitch   bool

	// I/O controller inputs
	IOAvailable  bool
	FireDetected bool
	EngineTemp   float64
	AuxTemp0     float64
	AuxTemp1     float64

	// Output measurement inputs
	OutputAvailable bool
	U1              float64
	U2              float64
	U3              float64
	P1              float64
	P2              float64
	P3              float64
	F               float64
}

type State struct {
	Node         StateNode
	Changed      time.Time
	ErrorTrigger error
}

type Outputs struct {
	Ignition bool
	Starter  bool
	Fan      bool
	Pump     bool
	Load     bool

	TimeInState time.Duration
	IoCheck     bool
	OutputCheck bool
}

type Change struct {
	f    func(Inputs) Inputs
	done chan struct{}
}

type Controller struct {
	Params  Params
	inputs  Inputs
	state   State
	outputs Outputs

	changeInputs   chan Change
	OnStateUpdate  func(State)
	OnOutputUpdate func(Outputs)
}

// NewController creates a new controller with the given configuration, initial state and inputs.
// It runs a goroutine that listens for input changes and recomputes the state and outputs.
// To terminate the controller, call the End method.
func NewController(params Params, initialNode StateNode, initialInputs Inputs) *Controller {
	initialState := State{
		Node:         initialNode,
		Changed:      initialInputs.Time,
		ErrorTrigger: nil,
	}

	c := &Controller{
		Params:       params,
		inputs:       initialInputs,
		state:        initialState,
		outputs:      computeOutputs(params, initialInputs, initialState),
		changeInputs: make(chan Change),
	}
	return c
}

// UpdateInputs sends a change to the controller inputs.
func (c *Controller) UpdateInputs(f func(Inputs) Inputs) {
	c.changeInputs <- Change{f: f}
}

// UpdateInputsSync sends a change to the controller inputs and waits for both the state and outputs
// to be recomputed and the OnUpdate functions to return.
// This is useful for testing.
func (c *Controller) UpdateInputsSync(f func(Inputs) Inputs) {
	done := make(chan struct{})
	c.changeInputs <- Change{f: f, done: done}
	<-done
}

// Run is the main loop of the controller. It will send the initial state and outputs to the respective channels
func (c *Controller) Run() {
	initDone := make(chan struct{})

	go func() {
		// send the initial state and outputs
		if c.OnStateUpdate != nil {
			c.OnStateUpdate(c.state)
		}
		if c.OnOutputUpdate != nil {
			c.OnOutputUpdate(c.outputs)
		}

		// initial computation
		c.compute()

		close(initDone)

		// listen for input changes until the channel is closed (End is called)
		for change := range c.changeInputs {
			// whenever an input is changed, recompute the derived inputs and the state
			if nextI := change.f(c.inputs); nextI != c.inputs {
				c.inputs = nextI
				c.compute()
			}

			// signal that the change is done to the UpdateInputsSync function
			if change.done != nil {
				close(change.done)
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

func (c *Controller) compute() {
	// compute new state
	// assume we reach the final state in max 8 iterations
	for i := 0; i < 8; i++ {
		nextState := computeState(c.Params, c.inputs, c.state)
		if nextState == c.state {
			break
		}
		c.state = nextState
		if c.OnStateUpdate != nil {
			c.OnStateUpdate(c.state)
		}
	}

	nextOutput := computeOutputs(c.Params, c.inputs, c.state)
	if nextOutput == c.outputs {
		return
	}
	c.outputs = nextOutput
	if c.OnOutputUpdate != nil {
		c.OnOutputUpdate(c.outputs)
	}
}

func computeState(p Params, i Inputs, prev State) (next State) {
	nextNode, errorTrigger := computeStateNode(p, i, prev)
	nextChanged := prev.Changed
	if nextNode != prev.Node {
		nextChanged = i.Time
	}
	return State{
		Node:         nextNode,
		Changed:      nextChanged,
		ErrorTrigger: errorTrigger,
	}
}

var ErrCrankingTimeout = errors.New("cranking timeout reached")

func computeStateNode(p Params, i Inputs, prev State) (next StateNode, errorTrigger error) {
	// Multi state transitions
	// in every case: reset switch triggers the reset state
	if i.ResetSwitch {
		return Reset, nil
	}

	// in every state except reset, off and failed: a temperature or fire detection triggers the failed state
	if !(prev.Node == Reset || prev.Node == Off || prev.Node == Error) { //nolint:staticcheck
		ioErr := ioError(p, i)
		if ioErr != nil {
			return Error, ioErr
		}
	}

	// in warm up, producing and engine cool down: a negative output check triggers the failed state
	outCheck := outputCheck(p, i)
	if prev.Node == WarmUp || prev.Node == Producing || prev.Node == EngineCoolDown {
		outputErr := outputError(p, i)
		if outputErr != nil {
			return Error, outputErr
		}
	}

	masterSwitch := i.ArmSwitch && i.CommandSwitch
	timeInState := i.Time.Sub(prev.Changed)

	// Single state transitions
	switch prev.Node {
	case Error:
	case Reset:
		if !i.ResetSwitch && !masterSwitch {
			return Off, nil
		}
	case Off:
		if i.IOAvailable {
			return Ready, nil
		}
	case Ready:
		if masterSwitch {
			return Priming, nil
		}
	case Priming:
		if !masterSwitch {
			return Ready, nil
		}
		if timeInState >= p.PrimingTimeout {
			return Cranking, nil
		}
	case Cranking:
		if !masterSwitch {
			return Ready, nil
		}
		if timeInState >= p.CrankingTimeout {
			return Error, ErrCrankingTimeout
		}
		if outCheck {
			return Stabilizing, nil
		}
	case Stabilizing:
		if !masterSwitch {
			return Ready, nil
		}
		if timeInState >= p.StabilizingTimeout {
			return WarmUp, nil
		}
	case WarmUp:
		if !masterSwitch {
			return EnclosureCoolDown, nil
		}

		if timeInState >= p.WarmUpMinTime {
			if timeInState >= p.WarmUpTimeout || i.EngineTemp >= p.WarmUpTemp {
				return Producing, nil
			}
		}
	case Producing:
		if !masterSwitch {
			return EngineCoolDown, nil
		}
	case EngineCoolDown:
		if masterSwitch {
			return Producing, nil
		}
		if timeInState >= p.EngineCoolDownMinTime {
			if timeInState >= p.EngineCoolDownTimeout || i.EngineTemp <= p.EngineCoolDownTemp {
				return EnclosureCoolDown, nil
			}
		}
	case EnclosureCoolDown:
		if masterSwitch {
			return Priming, nil
		}
		if timeInState >= p.EnclosureCoolDownMinTime {
			if timeInState >= p.EnclosureCoolDownTimeout || i.EngineTemp <= p.EnclosureCoolDownTemp {
				return Ready, nil
			}
		}
	}

	return prev.Node, prev.ErrorTrigger
}

func computeOutputs(p Params, i Inputs, s State) Outputs {
	sN := s.Node
	return Outputs{
		Ignition: sN == Cranking ||
			sN == Stabilizing ||
			sN == WarmUp ||
			sN == Producing ||
			sN == EngineCoolDown,
		Starter: sN == Cranking,
		Fan: sN == Priming ||
			sN == Cranking ||
			sN == Stabilizing ||
			sN == WarmUp ||
			sN == Producing ||
			sN == EngineCoolDown ||
			sN == EnclosureCoolDown,
		Pump: sN == Priming ||
			sN == Cranking ||
			sN == Stabilizing ||
			sN == WarmUp ||
			sN == Producing ||
			sN == EngineCoolDown,
		Load: sN == Producing,

		TimeInState: i.Time.Sub(s.Changed),
		IoCheck:     ioCheck(p, i),
		OutputCheck: outputCheck(p, i),
	}
}

func ioCheck(p Params, i Inputs) bool {
	return ioError(p, i) == nil
}

var ErrFireDirected = errors.New("fire detected")
var ErrIOUnavailable = errors.New("I/O controller unavailable")
var ErrEngineTempLow = errors.New("engine temperature too low")
var ErrEngineTempHigh = errors.New("engine temperature too high")
var ErrAuxTemp0Low = errors.New("auxiliary temperature 0 too low")
var ErrAuxTemp0High = errors.New("auxiliary temperature 0 too high")
var ErrAuxTemp1Low = errors.New("auxiliary temperature 1 too low")
var ErrAuxTemp1High = errors.New("auxiliary temperature 1 too high")

func ioError(p Params, i Inputs) error {
	if i.FireDetected {
		return ErrFireDirected
	}

	if !i.IOAvailable {
		return ErrIOUnavailable
	}

	if i.EngineTemp < p.EngineTempMin {
		return ErrEngineTempLow
	}
	if i.EngineTemp > p.EngineTempMax {
		return ErrEngineTempHigh
	}

	if i.AuxTemp0 < p.AuxTemp0Min {
		return ErrAuxTemp0Low
	}
	if i.AuxTemp0 > p.AuxTemp0Max {
		return ErrAuxTemp0High
	}

	if i.AuxTemp1 < p.AuxTemp1Min {
		return ErrAuxTemp1Low
	}
	if i.AuxTemp1 > p.AuxTemp1Max {
		return ErrAuxTemp1High
	}

	return nil
}

func outputCheck(p Params, i Inputs) bool {
	return outputError(p, i) == nil
}

var ErrOutputUnavailable = errors.New("output measurements unavailable")
var ErrFrequencyLow = errors.New("frequency too low")
var ErrFrequencyHigh = errors.New("frequency too high")
var ErrU1Low = errors.New("U1 too low")
var ErrU1High = errors.New("U1 too high")
var ErrU2Low = errors.New("U2 too low")
var ErrU2High = errors.New("U2 too high")
var ErrU3Low = errors.New("U3 too low")
var ErrU3High = errors.New("U3 too high")
var ErrP1High = errors.New("P1 too high")
var ErrP2High = errors.New("P2 too high")
var ErrP3High = errors.New("P3 too high")
var ErrPTotHigh = errors.New("total power too high")

func outputError(p Params, i Inputs) error {
	if !i.OutputAvailable {
		return ErrOutputUnavailable
	}

	if i.F < p.FMin {
		return ErrFrequencyLow
	}
	if i.F > p.FMax {
		return ErrFrequencyHigh
	}

	if i.U1 < p.UMin {
		return ErrU1Low
	}
	if i.U1 > p.UMax {
		return ErrU1High
	}

	if i.P1 > p.PMax {
		return ErrP1High
	}

	if p.SinglePhase {
		if i.P1 > p.PTotMax {
			return ErrPTotHigh
		}
	} else {
		if i.U2 < p.UMin {
			return ErrU2Low
		}
		if i.U2 > p.UMax {
			return ErrU2High
		}

		if i.U3 < p.UMin {
			return ErrU3Low
		}
		if i.U3 > p.UMax {
			return ErrU3High
		}

		if i.P2 > p.PMax {
			return ErrP2High
		}

		if i.P3 > p.PMax {
			return ErrP3High
		}

		if i.P1+i.P2+i.P3 > p.PTotMax {
			return ErrPTotHigh
		}
	}

	return nil
}
