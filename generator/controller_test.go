package generator_test

import (
	"github.com/koestler/go-iotdevice/v3/generator"
	"testing"
	"time"
)

func TestController(t *testing.T) {
	params := generator.Params{
		PrimingTimeout:           10 * time.Millisecond,
		CrankingTimeout:          20 * time.Millisecond,
		WarmUpTimeout:            30 * time.Millisecond,
		WarmUpTemp:               40,
		EngineCoolDownTimeout:    40 * time.Millisecond,
		EngineCoolDownTemp:       60,
		EnclosureCoolDownTimeout: 50 * time.Millisecond,
		EnclosureCoolDownTemp:    50,

		// IO Check
		EngineTempMin:     0,
		EngineTempMax:     100,
		AirIntakeTempMin:  0,
		AirIntakeTempMax:  100,
		AirExhaustTempMin: 0,
		AirExhaustTempMax: 100,

		// Output Check
		UMin:    210,
		UMax:    250,
		FMin:    45,
		FMax:    55,
		PMax:    1000,
		PTotMax: 2000,
	}

	t0 := time.Unix(0, 0).UTC()
	initialInputs := generator.Inputs{
		Time: t0,
	}

	t.Run("simpleSuccessfulRun", func(t *testing.T) {
		c := generator.NewController(params, generator.Off, initialInputs)

		stateTracker := newTracker[generator.State](t)
		go stateTracker.Drain(c.State())
		outputTracker := newTracker[generator.Outputs](t)
		go outputTracker.Drain(c.Outputs())

		c.Run()
		defer c.End()

		// initial state
		stateTracker.AssertLatest(t, generator.State{Node: generator.Off, Changed: t0})
		outputTracker.AssertLatest(t, generator.Outputs{})

		// go to ready
		c.UpdateInputsSync(func(i generator.Inputs) generator.Inputs {
			i.IOAvailable = true
			i.EngineTemp = 20
			i.AirIntakeTemp = 20
			i.AirExhaustTemp = 20
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.Ready, Changed: t0})
		outputTracker.AssertLatest(t, generator.Outputs{IoCheck: true})

		// go to priming
		c.UpdateInputsSync(func(i generator.Inputs) generator.Inputs {
			i.ArmSwitch = true
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.Ready, Changed: t0})
		outputTracker.AssertLatest(t, generator.Outputs{IoCheck: true})

		c.UpdateInputsSync(func(i generator.Inputs) generator.Inputs {
			i.CommandSwitch = true
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.Priming, Changed: t0})
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, Pump: true, IoCheck: true})

		// go to cranking
		t1 := t0.Add(params.PrimingTimeout)
		c.UpdateInputsSync(func(i generator.Inputs) generator.Inputs {
			i.Time = t1
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.Cranking, Changed: t1})
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, Pump: true, Ignition: true, Starter: true, IoCheck: true})

		// go to warm up
		c.UpdateInputsSync(func(i generator.Inputs) generator.Inputs {
			i.OutputAvailable = true
			i.U0 = 220
			i.U1 = 220
			i.U2 = 220
			i.F = 50
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.WarmUp, Changed: t0})
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, Pump: true, Ignition: true, IoCheck: true})

		// go to producing
		c.UpdateInputsSync(func(i generator.Inputs) generator.Inputs {
			i.EngineTemp = 45
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.Producing, Changed: t0})
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, Pump: true, Ignition: true, Load: true, IoCheck: true})

		// running, engine getting warm
		c.UpdateInputsSync(func(i generator.Inputs) generator.Inputs {
			i.EngineTemp = 70
			return i
		})

		// go to engine cool down
		c.UpdateInputsSync(func(i generator.Inputs) generator.Inputs {
			i.CommandSwitch = false
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.EngineCoolDown, Changed: t0})
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, Pump: true, Ignition: true, IoCheck: true})

		// go to enclosure cool down
		c.UpdateInputsSync(func(i generator.Inputs) generator.Inputs {
			i.EngineTemp = 55
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.EnclosureCoolDown, Changed: t0})
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, IoCheck: true})

		// go to ready
		c.UpdateInputsSync(func(i generator.Inputs) generator.Inputs {
			i.EngineTemp = 45
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.Ready, Changed: t0})
		outputTracker.AssertLatest(t, generator.Outputs{})
	})
}
