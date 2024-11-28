package genset_test

import (
	"fmt"
	"github.com/koestler/go-iotdevice/v3/genset"
	"github.com/stretchr/testify/assert"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

func controllerWithTracker(
	t *testing.T,
	params genset.Params,
	initialNode genset.StateNode,
	initialInputs genset.Inputs,
) (
	*genset.Controller, *tracker[genset.State], *tracker[genset.Outputs],
) {
	t.Helper()

	c := genset.NewController(params, initialNode, initialInputs)

	stateTracker := newTracker[genset.State](t)
	c.OnStateUpdate = stateTracker.OnUpdateFunc()
	outputTracker := newTracker[genset.Outputs](t)
	c.OnOutputUpdate = outputTracker.OnUpdateFunc()

	return c, stateTracker, outputTracker
}

func setInp(t *testing.T, c *genset.Controller, f func(i genset.Inputs) genset.Inputs) {
	t.Helper()
	c.UpdateInputsSync(func(i genset.Inputs) genset.Inputs {
		i = f(i)
		t.Logf("inputs: %v", i)
		return i
	})
}

func TestController(t *testing.T) {
	params := genset.Params{
		PrimingTimeout:           10 * time.Second,
		CrankingTimeout:          20 * time.Second,
		WarmUpTimeout:            10 * time.Minute,
		WarmUpTemp:               40,
		EngineCoolDownTimeout:    5 * time.Minute,
		EngineCoolDownTemp:       60,
		EnclosureCoolDownTimeout: 15 * time.Minute,
		EnclosureCoolDownTemp:    50,

		// IO Check
		EngineTempMin: 10,
		EngineTempMax: 90,
		AuxTemp0Min:   0,
		AuxTemp0Max:   100,
		AuxTemp1Min:   -10,
		AuxTemp1Max:   150,

		// Output Check
		UMin:    210,
		UMax:    250,
		FMin:    45,
		FMax:    55,
		PMax:    1000,
		PTotMax: 2000,
	}

	t0, _ := time.Parse(time.RFC3339, "2000-01-01T00:00:00Z")
	t.Run("completeRun", func(t *testing.T) {
		c, stateTracker, outputTracker := controllerWithTracker(t, params, genset.Off, genset.Inputs{Time: t0})

		c.Run()
		defer c.End()

		// initial state
		stateTracker.AssertLatest(t, genset.State{Node: genset.Off, Changed: t0})
		outputTracker.AssertLatest(t, genset.Outputs{})

		// go to ready
		setInp(t, c, func(i genset.Inputs) genset.Inputs {
			i.IOAvailable = true
			i.EngineTemp = 20
			return i
		})
		stateTracker.AssertLatest(t, genset.State{Node: genset.Ready, Changed: t0})
		outputTracker.AssertLatest(t, genset.Outputs{IoCheck: true})

		// go to priming
		setInp(t, c, func(i genset.Inputs) genset.Inputs {
			i.ArmSwitch = true
			return i
		})
		stateTracker.AssertLatest(t, genset.State{Node: genset.Ready, Changed: t0})
		outputTracker.AssertLatest(t, genset.Outputs{IoCheck: true})

		setInp(t, c, func(i genset.Inputs) genset.Inputs {
			i.CommandSwitch = true
			return i
		})
		stateTracker.AssertLatest(t, genset.State{Node: genset.Priming, Changed: t0})
		outputTracker.AssertLatest(t, genset.Outputs{Fan: true, Pump: true, IoCheck: true})

		// stay in priming
		t1 := t0.Add(params.PrimingTimeout / 2)
		setInp(t, c, func(i genset.Inputs) genset.Inputs {
			i.Time = t1
			return i
		})

		// go to cranking
		t2 := t0.Add(params.PrimingTimeout)
		setInp(t, c, func(i genset.Inputs) genset.Inputs {
			i.Time = t2
			return i
		})

		stateTracker.AssertLatest(t, genset.State{Node: genset.Cranking, Changed: t2})
		outputTracker.AssertLatest(t, genset.Outputs{Fan: true, Pump: true, Ignition: true, Starter: true, IoCheck: true})

		// go to warm up
		setInp(t, c, func(i genset.Inputs) genset.Inputs {
			i.OutputAvailable = true
			i.U0 = 220
			i.U1 = 220
			i.U2 = 220
			i.F = 50
			return i
		})
		stateTracker.AssertLatest(t, genset.State{Node: genset.WarmUp, Changed: t2})
		outputTracker.AssertLatest(t, genset.Outputs{Fan: true, Pump: true, Ignition: true, IoCheck: true, OutputCheck: true})

		// go to producing
		setInp(t, c, func(i genset.Inputs) genset.Inputs {
			i.EngineTemp = 45
			return i
		})
		stateTracker.AssertLatest(t, genset.State{Node: genset.Producing, Changed: t2})
		outputTracker.AssertLatest(t, genset.Outputs{Fan: true, Pump: true, Ignition: true, Load: true, IoCheck: true, OutputCheck: true})

		// running, engine getting warm, frequency fluctuating
		t3 := t2.Add(time.Second)
		setInp(t, c, func(i genset.Inputs) genset.Inputs {
			i.Time = t3
			i.EngineTemp = 70
			i.F = 48
			i.L0 = 1000
			return i
		})
		stateTracker.AssertLatest(t, genset.State{Node: genset.Producing, Changed: t2})
		outputTracker.AssertLatest(t, genset.Outputs{
			Fan: true, Pump: true, Ignition: true, Load: true,
			TimeInState: time.Second, IoCheck: true, OutputCheck: true,
		})

		t4 := t3.Add(time.Second)
		setInp(t, c, func(i genset.Inputs) genset.Inputs {
			i.Time = t4
			i.EngineTemp = 72
			i.F = 51
			return i
		})
		stateTracker.AssertLatest(t, genset.State{Node: genset.Producing, Changed: t2})
		outputTracker.AssertLatest(t, genset.Outputs{
			Fan: true, Pump: true, Ignition: true, Load: true,
			TimeInState: 2 * time.Second, IoCheck: true, OutputCheck: true,
		})

		// go to engine cool down
		t5 := t4.Add(time.Second)
		setInp(t, c, func(i genset.Inputs) genset.Inputs {
			i.Time = t5
			i.CommandSwitch = false
			return i
		})
		stateTracker.AssertLatest(t, genset.State{Node: genset.EngineCoolDown, Changed: t5})
		outputTracker.AssertLatest(t, genset.Outputs{Fan: true, Pump: true, Ignition: true, IoCheck: true, OutputCheck: true})

		// go to enclosure cool down
		t6 := t5.Add(time.Second)
		setInp(t, c, func(i genset.Inputs) genset.Inputs {
			i.Time = t6
			i.EngineTemp = 55
			return i
		})
		stateTracker.AssertLatest(t, genset.State{Node: genset.EnclosureCoolDown, Changed: t6})
		outputTracker.AssertLatest(t, genset.Outputs{Fan: true, IoCheck: true, OutputCheck: true})

		// stay in enclosure cool down, engine has stopped
		t7 := t6.Add(time.Second)
		setInp(t, c, func(i genset.Inputs) genset.Inputs {
			i.Time = t7
			i.F = 0
			i.U0 = 10
			i.U1 = 10
			i.U2 = 10
			i.L0 = 2
			i.L1 = 2
			i.L2 = 2
			return i
		})
		stateTracker.AssertLatest(t, genset.State{Node: genset.EnclosureCoolDown, Changed: t6})
		outputTracker.AssertLatest(t, genset.Outputs{Fan: true, IoCheck: true, TimeInState: time.Second})

		// go to ready
		t8 := t7.Add(time.Minute)
		setInp(t, c, func(i genset.Inputs) genset.Inputs {
			i.Time = t8
			i.EngineTemp = 45
			return i
		})
		stateTracker.AssertLatest(t, genset.State{Node: genset.Ready, Changed: t8})
		outputTracker.AssertLatest(t, genset.Outputs{IoCheck: true})
	})

	t.Run("priming", func(t *testing.T) {
		initialState := genset.Priming
		initialInputs := genset.Inputs{
			Time:          t0,
			ArmSwitch:     true,
			CommandSwitch: true,
			IOAvailable:   true,
			EngineTemp:    20,
		}

		t.Run("success", func(t *testing.T) {
			c, stateTracker, outputTracker := controllerWithTracker(t, params, initialState, initialInputs)
			c.Run()
			defer c.End()

			// check pump on
			{
				l, ok := outputTracker.Latest()
				assert.True(t, ok)
				assert.Equal(t, true, l.Pump)
			}

			// not enough time passed
			t1 := t0.Add(time.Second)
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = t1
				return i
			})

			// enough time passed
			t2 := t0.Add(params.CrankingTimeout)
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = t2
				return i
			})

			stateTracker.Assert(t, []genset.State{
				{Node: genset.Priming, Changed: t0},
				{Node: genset.Cranking, Changed: t2},
			})

			// check pump still on
			{
				l, ok := outputTracker.Latest()
				assert.True(t, ok)
				assert.Equal(t, true, l.Pump)
			}
		})

		t.Run("abort", func(t *testing.T) {
			c, stateTracker, _ := controllerWithTracker(t, params, initialState, initialInputs)
			c.Run()
			defer c.End()

			t1 := t0.Add(time.Second)
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = t1
				i.CommandSwitch = false
				return i
			})

			stateTracker.Assert(t, []genset.State{
				{Node: genset.Priming, Changed: t0},
				{Node: genset.Ready, Changed: t1},
			})
		})
	})

	t.Run("cranking", func(t *testing.T) {
		initialState := genset.Cranking
		initialInputs := genset.Inputs{
			Time:          t0,
			ArmSwitch:     true,
			CommandSwitch: true,
			IOAvailable:   true,
			EngineTemp:    20,
		}

		t.Run("success", func(t *testing.T) {
			c, stateTracker, outputTracker := controllerWithTracker(t, params, initialState, initialInputs)
			c.Run()
			defer c.End()

			{ // check starter on
				l, ok := outputTracker.Latest()
				assert.True(t, ok)
				assert.Equal(t, true, l.Starter)
			}

			// engine started
			t1 := t0.Add(time.Second)
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = t1
				i.OutputAvailable = true
				i.U0 = 221
				i.U1 = 219
				i.U2 = 222
				i.F = 49
				return i
			})

			stateTracker.Assert(t, []genset.State{
				{Node: genset.Cranking, Changed: t0},
				{Node: genset.WarmUp, Changed: t1},
			})

			{ // check starter off
				l, ok := outputTracker.Latest()
				assert.True(t, ok)
				assert.Equal(t, false, l.Starter)
			}
		})

		t.Run("timeout", func(t *testing.T) {
			c, stateTracker, outputTracker := controllerWithTracker(t, params, initialState, initialInputs)
			c.Run()
			defer c.End()

			{ // check starter on
				l, ok := outputTracker.Latest()
				assert.True(t, ok)
				assert.Equal(t, true, l.Starter)
			}

			// run the clock
			for range 30 {
				setInp(t, c, func(i genset.Inputs) genset.Inputs {
					i.Time = i.Time.Add(time.Second)
					return i
				})
			}

			stateTracker.Assert(t, []genset.State{
				{Node: genset.Cranking, Changed: t0},
				{Node: genset.Error, Changed: t0.Add(params.CrankingTimeout)},
			})

			{ // check starter off
				l, ok := outputTracker.Latest()
				assert.True(t, ok)
				assert.Equal(t, false, l.Starter)
			}
		})

		t.Run("abort", func(t *testing.T) {
			c, stateTracker, _ := controllerWithTracker(t, params, initialState, initialInputs)
			c.Run()
			defer c.End()

			t1 := t0.Add(time.Second)
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = t1
				i.CommandSwitch = false
				return i
			})

			stateTracker.Assert(t, []genset.State{
				{Node: genset.Cranking, Changed: t0},
				{Node: genset.Ready, Changed: t1},
			})
		})
	})

	t.Run("wamUp", func(t *testing.T) {
		initialState := genset.WarmUp
		initialInputs := genset.Inputs{
			Time:          t0,
			ArmSwitch:     true,
			CommandSwitch: true,

			IOAvailable:     true,
			EngineTemp:      20,
			OutputAvailable: true,
			U0:              220,
			U1:              220,
			U2:              220,
			F:               50,
		}

		t.Run("byTime", func(t *testing.T) {
			c, stateTracker, outputTracker := controllerWithTracker(t, params, initialState, initialInputs)
			c.Run()
			defer c.End()

			// go to producing after timeout
			t1 := t0.Add(params.WarmUpTimeout)
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = t1
				return i
			})

			stateTracker.Assert(t, []genset.State{
				{Node: genset.WarmUp, Changed: t0},
				{Node: genset.Producing, Changed: t1},
			})

			l, ok := outputTracker.Latest()
			assert.True(t, ok)
			assert.Equal(t, true, l.Load)
		})

		t.Run("byTemp", func(t *testing.T) {
			c, stateTracker, outputTracker := controllerWithTracker(t, params, initialState, initialInputs)
			c.Run()
			defer c.End()

			// go to producing after timeout
			t1 := t0.Add(time.Second)
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = t1
				i.EngineTemp = 41
				return i
			})

			stateTracker.Assert(t, []genset.State{
				{Node: genset.WarmUp, Changed: t0},
				{Node: genset.Producing, Changed: t1},
			})

			l, ok := outputTracker.Latest()
			assert.True(t, ok)
			assert.Equal(t, true, l.Load)
		})
	})

	t.Run("producing", func(t *testing.T) {
		initialState := genset.Producing
		initialInputs := genset.Inputs{
			Time:          t0,
			ArmSwitch:     true,
			CommandSwitch: true,

			IOAvailable:     true,
			EngineTemp:      55,
			OutputAvailable: true,
			U0:              220,
			U1:              220,
			U2:              220,
			F:               50,
		}

		t.Run("running", func(t *testing.T) {
			c, stateTracker, outputTracker := controllerWithTracker(t, params, initialState, initialInputs)
			c.Run()
			defer c.End()

			// voltage and frequency fluctuating
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = i.Time.Add(time.Second)
				i.U0 = params.UMin + 1
				i.U1 = params.UMin + 1
				i.U2 = params.UMin + 1
				i.F = params.FMin + 1
				return i
			})

			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = i.Time.Add(time.Second)
				i.U0 = params.UMax - 1
				i.U1 = params.UMax - 1
				i.U2 = params.UMax - 1
				i.F = params.FMax - 1
				return i
			})

			// temperature fluctuating
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = i.Time.Add(time.Second)
				i.EngineTemp = params.EngineTempMin + 1
				i.AuxTemp0 = params.AuxTemp0Min + 1
				i.AuxTemp1 = params.AuxTemp1Min + 1
				return i
			})

			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = i.Time.Add(time.Second)
				i.EngineTemp = params.EngineTempMax - 1
				i.AuxTemp0 = params.AuxTemp0Max - 1
				i.AuxTemp1 = params.AuxTemp1Max - 1
				return i
			})

			// power fluctuating
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = i.Time.Add(time.Second)
				i.L0 = 0
				i.L1 = 0
				i.L2 = 0
				return i
			})

			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = i.Time.Add(time.Second)
				i.L0 = params.PMax - 1
				i.L1 = params.PMax - 1
				i.L2 = params.PTotMax - i.L1 - i.L0 - 1
				return i
			})

			stateTracker.Assert(t, []genset.State{
				{Node: genset.Producing, Changed: t0},
			})

			l, ok := outputTracker.Latest()
			assert.True(t, ok)
			assert.Equal(t, true, l.Load)
		})

		t.Run("commandOff", func(t *testing.T) {
			c, stateTracker, outputTracker := controllerWithTracker(t, params, initialState, initialInputs)
			c.Run()
			defer c.End()

			// warm up
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = i.Time.Add(time.Second)
				i.EngineTemp = 75
				return i
			})

			// switch off
			t1 := t0.Add(20 * time.Second)
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = t1
				i.CommandSwitch = false
				return i
			})

			stateTracker.Assert(t, []genset.State{
				{Node: genset.Producing, Changed: t0},
				{Node: genset.EngineCoolDown, Changed: t1},
			})

			l, ok := outputTracker.Latest()
			assert.True(t, ok)
			assert.Equal(t, false, l.Load)
		})

		t.Run("outputFailure", func(t *testing.T) {
			c, stateTracker, outputTracker := controllerWithTracker(t, params, initialState, initialInputs)
			c.Run()
			defer c.End()

			t1 := t0.Add(time.Second)
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = t1
				i.F = params.FMax + 1
				return i
			})

			stateTracker.Assert(t, []genset.State{
				{Node: genset.Producing, Changed: t0},
				{Node: genset.Error, Changed: t1},
			})

			// check ignition is off
			l, ok := outputTracker.Latest()
			assert.True(t, ok)
			assert.Equal(t, false, l.Ignition)
		})

		t.Run("ioFailure", func(t *testing.T) {
			c, stateTracker, outputTracker := controllerWithTracker(t, params, initialState, initialInputs)
			c.Run()
			defer c.End()

			t1 := t0.Add(time.Second)
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = t1
				i.IOAvailable = false
				return i
			})

			stateTracker.Assert(t, []genset.State{
				{Node: genset.Producing, Changed: t0},
				{Node: genset.Error, Changed: t1},
			})

			// check ignition is off
			l, ok := outputTracker.Latest()
			assert.True(t, ok)
			assert.Equal(t, false, l.Ignition)
		})
	})

	t.Run("reset", func(t *testing.T) {
		t.Run("fromError", func(t *testing.T) {
			initialState := genset.Error
			initialInputs := genset.Inputs{
				Time:        t0,
				IOAvailable: true,
				EngineTemp:  20,
			}

			c, stateTracker, _ := controllerWithTracker(t, params, initialState, initialInputs)
			c.Run()
			defer c.End()

			t1 := t0.Add(time.Second)
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = t1
				i.ResetSwitch = true
				return i
			})

			t2 := t1.Add(time.Millisecond)
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = t2
				i.ResetSwitch = false
				return i
			})

			stateTracker.Assert(t, []genset.State{
				{Node: genset.Error, Changed: t0},
				{Node: genset.Reset, Changed: t1},
				{Node: genset.Off, Changed: t2},
				{Node: genset.Ready, Changed: t2},
			})
		})

		t.Run("fromProducing", func(t *testing.T) {
			initialState := genset.Producing
			initialInputs := genset.Inputs{
				Time:            t0,
				IOAvailable:     true,
				OutputAvailable: true,
				U0:              220,
				U1:              220,
				U2:              220,
				F:               50,
				EngineTemp:      45,
			}

			c, stateTracker, outputTracker := controllerWithTracker(t, params, initialState, initialInputs)
			c.Run()
			defer c.End()

			// check ignition is on in producing
			l, ok := outputTracker.Latest()
			assert.True(t, ok)
			assert.Equal(t, true, l.Ignition)

			t1 := t0.Add(time.Second)
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = t1
				i.ResetSwitch = true
				return i
			})

			// check ignition is immediately off after reset
			l, ok = outputTracker.Latest()
			assert.True(t, ok)
			assert.Equal(t, false, l.Ignition)

			t2 := t1.Add(time.Millisecond)
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = t2
				i.ResetSwitch = false
				return i
			})

			stateTracker.Assert(t, []genset.State{
				{Node: genset.Producing, Changed: t0},
				{Node: genset.Reset, Changed: t1},
				{Node: genset.Off, Changed: t2},
				{Node: genset.Ready, Changed: t2},
			})
		})
	})

	t.Run("engineCoolDown", func(t *testing.T) {
		initialState := genset.EngineCoolDown
		initialInputs := genset.Inputs{
			Time:            t0,
			ArmSwitch:       true,
			CommandSwitch:   false,
			IOAvailable:     true,
			EngineTemp:      70,
			OutputAvailable: true,
			U0:              220,
			U1:              220,
			U2:              220,
			F:               50,
		}

		t.Run("byTime", func(t *testing.T) {
			c, stateTracker, outputTracker := controllerWithTracker(t, params, initialState, initialInputs)
			c.Run()
			defer c.End()

			l, ok := outputTracker.Latest()
			assert.True(t, ok)
			assert.Equal(t, true, l.Ignition)

			// wait for timeout
			t1 := t0.Add(params.EngineCoolDownTimeout)
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = t1
				return i
			})

			stateTracker.Assert(t, []genset.State{
				{Node: genset.EngineCoolDown, Changed: t0},
				{Node: genset.EnclosureCoolDown, Changed: t1},
			})

			l, ok = outputTracker.Latest()
			assert.True(t, ok)
			assert.Equal(t, false, l.Ignition)
		})

		t.Run("byTemp", func(t *testing.T) {
			c, stateTracker, outputTracker := controllerWithTracker(t, params, initialState, initialInputs)
			c.Run()
			defer c.End()

			l, ok := outputTracker.Latest()
			assert.True(t, ok)
			assert.Equal(t, true, l.Ignition)

			// cool down enough
			t1 := t0.Add(time.Minute)
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = t1
				i.EngineTemp = params.EngineCoolDownTemp - 1
				return i
			})

			stateTracker.Assert(t, []genset.State{
				{Node: genset.EngineCoolDown, Changed: t0},
				{Node: genset.EnclosureCoolDown, Changed: t1},
			})

			l, ok = outputTracker.Latest()
			assert.True(t, ok)
			assert.Equal(t, false, l.Ignition)
		})
	})

	t.Run("enclosureCoolDown", func(t *testing.T) {
		initialState := genset.EnclosureCoolDown
		initialInputs := genset.Inputs{
			Time:          t0,
			ArmSwitch:     true,
			CommandSwitch: false,
			IOAvailable:   true,
			EngineTemp:    params.EngineCoolDownTemp - 1,
		}

		t.Run("byTime", func(t *testing.T) {
			c, stateTracker, outputTracker := controllerWithTracker(t, params, initialState, initialInputs)
			c.Run()
			defer c.End()

			l, ok := outputTracker.Latest()
			assert.True(t, ok)
			assert.Equal(t, true, l.Fan)

			// pass some time
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = i.Time.Add(time.Millisecond)
				return i
			})

			// wait for timeout
			t1 := t0.Add(params.EnclosureCoolDownTimeout)
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = t1
				return i
			})

			stateTracker.Assert(t, []genset.State{
				{Node: genset.EnclosureCoolDown, Changed: t0},
				{Node: genset.Ready, Changed: t1},
			})

			l, ok = outputTracker.Latest()
			assert.True(t, ok)
			assert.Equal(t, false, l.Fan)
		})

		t.Run("byTemp", func(t *testing.T) {
			c, stateTracker, _ := controllerWithTracker(t, params, initialState, initialInputs)
			c.Run()
			defer c.End()

			// decrease temp
			t1 := t0.Add(time.Second)
			setInp(t, c, func(i genset.Inputs) genset.Inputs {
				i.Time = t1
				i.EngineTemp = params.EnclosureCoolDownTemp - 1
				return i
			})

			stateTracker.Assert(t, []genset.State{
				{Node: genset.EnclosureCoolDown, Changed: t0},
				{Node: genset.Ready, Changed: t1},
			})
		})
	})
}

func BenchmarkUpdateInputs(b *testing.B) {
	b.Run("inputChanges", func(b *testing.B) {
		c := genset.NewController(genset.Params{}, genset.Off, genset.Inputs{})
		c.Run()
		defer c.End()

		for i := 0; i < b.N; i++ {
			c.UpdateInputs(func(i genset.Inputs) genset.Inputs {
				i.Time = time.Now()
				return i
			})
		}
	})
}

func BenchmarkUpdateInputsSync(b *testing.B) {
	b.Run("inputChanges", func(b *testing.B) {
		c := genset.NewController(genset.Params{}, genset.Off, genset.Inputs{})
		c.Run()
		defer c.End()

		for i := 0; i < b.N; i++ {
			c.UpdateInputsSync(func(i genset.Inputs) genset.Inputs {
				i.Time = time.Now()
				return i
			})
		}
	})
}

func TestSyncWg(t *testing.T) {
	events := make([]string, 0, 20)
	eventsLock := sync.Mutex{}
	addEvent := func(pos string, v int) {
		eventsLock.Lock()
		defer eventsLock.Unlock()
		events = append(events, fmt.Sprintf("%02d: %v", v, pos))
	}

	type Change struct {
		v  int
		wg *sync.WaitGroup
	}

	inpChan := make(chan Change)
	oupChan := make(chan Change)

	go func() {
		for o := range oupChan {
			addEvent("C", o.v)
			o.wg.Done()
		}
	}()

	syncFunc := func(v int) {
		addEvent("A", v)
		wg := &sync.WaitGroup{}
		wg.Add(1)
		inpChan <- Change{
			v:  v,
			wg: wg,
		}
		wg.Wait()
		addEvent("D", v)
	}

	go func() {
		for c := range inpChan {
			addEvent("B", c.v)
			oupChan <- Change{
				v:  c.v,
				wg: c.wg,
			}
		}
	}()

	for i := 0; i < 20; i++ {
		syncFunc(i)
	}

	eventsLock.Lock()
	defer eventsLock.Unlock()

	sortedEvents := append([]string{}, events...)
	sort.Strings(sortedEvents)

	if !reflect.DeepEqual(events, sortedEvents) {
		t.Error("events not sorted")
		t.Logf("events: %s", strings.Join(events, "\n"))
		t.Logf("sortedEvents: %s", strings.Join(sortedEvents, "\n"))
	}
}
