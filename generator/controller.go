package generator

import (
	"time"
)

// Params contains the parameters for the generator controller.
type Params struct {
	// Transition params
	PrimingTimeout           time.Duration
	CrankingTimeout          time.Duration
	WarmUpTimeout            time.Duration
	WarmUpTemp               float64
	EngineCoolDownTimeout    time.Duration
	EngineCoolDownTemp       float64
	EnclosureCoolDownTimeout time.Duration
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

type StateNode int

const (
	Error StateNode = iota
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
	// Time is an input to the controller to allow for time-based state transitions
	Time time.Time

	// Switches
	CommandSwitch bool
	ResetSwitch   bool

	// I/O controller inputs
	IOAvailable  bool
	ArmSwitch    bool
	FireDetected bool
	EngineTemp   float64
	AuxTemp0     float64
	AuxTemp1     float64

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

type State struct {
	Node    StateNode
	Changed time.Time
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
		Node:    initialNode,
		Changed: initialInputs.Time,
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
	// assume we transition at most 8 states once when inputs change
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
	nextNode := computeStateNode(p, i, prev)
	nextChanged := prev.Changed
	if nextNode != prev.Node {
		nextChanged = i.Time
	}
	return State{
		Node:    nextNode,
		Changed: nextChanged,
	}
}

func computeStateNode(p Params, i Inputs, prev State) (next StateNode) {
	// Multi state transitions
	// in every case: reset switch triggers the reset state
	if i.ResetSwitch {
		return Reset
	}

	// in every state except reset, off and failed: a temperature or fire detection triggers the failed state
	if !(prev.Node == Reset || prev.Node == Off || prev.Node == Error) && !ioCheck(p, i) {
		return Error
	}

	// in warm up, producing and engine cool down: a negative output check triggers the failed state
	outCheck := outputCheck(p, i)
	if (prev.Node == WarmUp || prev.Node == Producing || prev.Node == EngineCoolDown) && !outCheck {
		return Error
	}

	masterSwitch := i.ArmSwitch && i.CommandSwitch
	timeInState := i.Time.Sub(prev.Changed)

	// Single state transitions
	switch prev.Node {
	case Error:
	case Reset:
		if !i.ResetSwitch {
			return Off
		}
	case Off:
		if i.IOAvailable {
			return Ready
		}
	case Ready:
		if masterSwitch {
			return Priming
		}
	case Priming:
		if !masterSwitch {
			return Ready
		}
		if timeInState >= p.PrimingTimeout {
			return Cranking
		}
	case Cranking:
		if !masterSwitch {
			return Ready
		}
		if timeInState >= p.CrankingTimeout {
			return Error
		}
		if outCheck {
			return WarmUp
		}
	case WarmUp:
		if !masterSwitch {
			return EnclosureCoolDown
		}
		if timeInState >= p.WarmUpTimeout || i.EngineTemp >= p.WarmUpTemp {
			return Producing
		}
	case Producing:
		if !masterSwitch {
			return EngineCoolDown
		}
	case EngineCoolDown:
		if masterSwitch {
			return Producing
		}
		if timeInState >= p.EngineCoolDownTimeout || i.EngineTemp <= p.EngineCoolDownTemp {
			return EnclosureCoolDown
		}
	case EnclosureCoolDown:
		if masterSwitch {
			return Priming
		}
		if timeInState >= p.EnclosureCoolDownTimeout || i.EngineTemp <= p.EnclosureCoolDownTemp {
			return Ready
		}
	}

	return prev.Node
}

func computeOutputs(p Params, i Inputs, s State) Outputs {
	sN := s.Node
	return Outputs{
		Ignition: sN == Cranking ||
			sN == WarmUp ||
			sN == Producing ||
			sN == EngineCoolDown,
		Starter: sN == Cranking,
		Fan: sN == Priming ||
			sN == Cranking ||
			sN == WarmUp ||
			sN == Producing ||
			sN == EngineCoolDown ||
			sN == EnclosureCoolDown,
		Pump: sN == Priming ||
			sN == Cranking ||
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
	return !i.FireDetected && i.IOAvailable &&
		i.EngineTemp >= p.EngineTempMin && i.EngineTemp <= p.EngineTempMax &&
		i.AuxTemp0 >= p.AuxTemp0Min && i.AuxTemp0 <= p.AuxTemp0Max &&
		i.AuxTemp1 >= p.AuxTemp1Min && i.AuxTemp1 <= p.AuxTemp1Max
}

func outputCheck(p Params, i Inputs) bool {
	if !i.OutputAvailable {
		return false
	}

	if i.F < p.FMin || i.F > p.FMax {
		return false
	}

	if p.SinglePhase {
		return i.U0 >= p.UMin && i.U0 <= p.UMax &&
			i.L0 <= p.PMax &&
			i.L0 <= p.PTotMax
	}
	return i.F >= p.FMin && i.F <= p.FMax &&
		i.U0 >= p.UMin && i.U0 <= p.UMax &&
		i.U1 >= p.UMin && i.U1 <= p.UMax &&
		i.U2 >= p.UMin && i.U2 <= p.UMax &&
		i.L0 <= p.PMax &&
		i.L1 <= p.PMax &&
		i.L2 <= p.PMax &&
		i.L0+i.L1+i.L2 <= p.PTotMax
}
