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
		EngineCoolDownTemp:       50,
		EnclosureCoolDownTimeout: 50 * time.Millisecond,
		EnclosureCoolDownTemp:    60,
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

		c.Run()
		defer c.End()

		stateTracker.AssertLatest(t, generator.Off)

		c.UpdateInputsSync(func(i generator.Inputs) generator.Inputs {
			i.IOAvailable = true
			return i
		})

		stateTracker.AssertLatest(t, generator.Ready)
	})
}
