package generatorDevice

import (
	"testing"
	"time"
)

func TestController(t *testing.T) {
	config := Configuration{
		InStateUpdateInterval:    1 * time.Millisecond,
		PrimingTimeout:           10 * time.Millisecond,
		CrankingTmeout:           20 * time.Millisecond,
		WarmUpTimeout:            30 * time.Millisecond,
		WarmUpTemp:               40,
		EngineCoolDownTimeout:    40 * time.Millisecond,
		EngineCoolDownTemp:       50,
		EnclosureCoolDownTimeout: 50 * time.Millisecond,
		EnclosureCoolDownTemp:    60,
		EngineTempMin:            0,
		EngineTempMax:            100,
		AirIntakeTempMin:         -10,
		AirIntakeTempMax:         40,
		AirExhaustTempMin:        -10,
		AirExhaustTempMax:        90,
		UMin:                     210,
		UMax:                     240,
		FMin:                     47,
		FMax:                     53,
		PMax:                     6000,
		PTotMax:                  10000,
	}

	t.Run("simpleSuccessfulRun", func(t *testing.T) {
		c := NewController(config)
		c.Run()
		defer close(c.ChangeInput)

		expectNewState(t, c, Off)
		<-c.InputsChanged

		c.ChangeInput <- func(i Inputs) Inputs {
			i.IOAvailable = true
			return i
		}

		expectSameState(t, c)

	})
}

func expectSameState(t *testing.T, c *Controller) {
	t.Helper()

	// we expect the controller to update the state within 1ms
	time.Sleep(1 * time.Millisecond)

	select {
	case s := <-c.StateChanged:
		t.Errorf("expected no state change but got %v", s)
	default:
		return
	}
}

func expectNewState(t *testing.T, c *Controller, expected State) {
	t.Helper()

	// we expect the controller to update the state within 1ms
	timeout := time.NewTimer(time.Millisecond)
	defer timeout.Stop()

	select {
	case s, ok := <-c.StateChanged:
		if !ok {
			t.Errorf("state channel closed")
		}
		if s != expected {
			t.Errorf("expected state %v but got %v", expected, s)
		}
	case <-timeout.C:
		t.Errorf("expected state update to %v but got nothing", expected)
	}
}
