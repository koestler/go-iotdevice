package generator_test

import (
	"github.com/koestler/go-iotdevice/v3/generator"
	"testing"
	"time"
)

func TestController(t *testing.T) {
	config := generator.Configuration{
		InStateResolution:        1 * time.Millisecond,
		PrimingTimeout:           10 * time.Millisecond,
		CrankingTmeout:           20 * time.Millisecond,
		WarmUpTimeout:            30 * time.Millisecond,
		WarmUpTemp:               40,
		EngineCoolDownTimeout:    40 * time.Millisecond,
		EngineCoolDownTemp:       60,
		EnclosureCoolDownTimeout: 50 * time.Millisecond,
		EnclosureCoolDownTemp:    50,
		IOCheck: func(i generator.Inputs) bool {
			return i.IOAvailable
		},
		OutputCheck: func(i generator.Inputs) bool {
			return i.OutputAvailable
		},
	}

	t.Run("simpleSuccessfulRun", func(t *testing.T) {
		c := generator.NewController(config)

		stateTracker := newTracker[generator.State](t)
		go stateTracker.Drain(c.State())
		outputTracker := newTracker[generator.Outputs](t)
		go outputTracker.Drain(c.Outputs())

		c.Run()
		defer c.End()

		// initial state
		stateTracker.AssertLatest(t, generator.Off)
		outputTracker.AssertLatest(t, generator.Outputs{})

		// go to ready
		c.UpdateInputsSync(func(i generator.Inputs) generator.Inputs {
			i.IOAvailable = true
			return i
		})
		stateTracker.AssertLatest(t, generator.Ready)
		outputTracker.AssertLatest(t, generator.Outputs{})

		// go to priming
		c.UpdateInputsSync(func(i generator.Inputs) generator.Inputs {
			i.ArmSwitch = true
			return i
		})
		stateTracker.AssertLatest(t, generator.Ready)
		outputTracker.AssertLatest(t, generator.Outputs{})

		c.UpdateInputsSync(func(i generator.Inputs) generator.Inputs {
			i.CommandSwitch = true
			return i
		})
		stateTracker.AssertLatest(t, generator.Priming)
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, Pump: true})

		// go to cranking
		time.Sleep(config.PrimingTimeout + config.InStateResolution)
		stateTracker.AssertLatest(t, generator.Cranking)
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, Pump: true, Ignition: true, Starter: true})

		// go to warm up
		c.UpdateInputsSync(func(i generator.Inputs) generator.Inputs {
			i.OutputAvailable = true
			return i
		})
		stateTracker.AssertLatest(t, generator.WarmUp)
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, Pump: true, Ignition: true})

		// go to producing
		c.UpdateInputsSync(func(i generator.Inputs) generator.Inputs {
			i.EngineTemp = 45
			return i
		})
		time.Sleep(time.Millisecond)
		stateTracker.AssertLatest(t, generator.Producing)
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, Pump: true, Ignition: true, Load: true})

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
		stateTracker.AssertLatest(t, generator.EngineCoolDown)
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, Pump: true, Ignition: true})

		// go to enclosure cool down
		c.UpdateInputsSync(func(i generator.Inputs) generator.Inputs {
			i.EngineTemp = 55
			return i
		})
		time.Sleep(time.Millisecond)
		stateTracker.AssertLatest(t, generator.EnclosureCoolDown)
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true})

		// go to ready
		c.UpdateInputsSync(func(i generator.Inputs) generator.Inputs {
			i.EngineTemp = 45
			return i
		})
		time.Sleep(time.Millisecond)
		stateTracker.AssertLatest(t, generator.Ready)
		outputTracker.AssertLatest(t, generator.Outputs{})
	})
}
