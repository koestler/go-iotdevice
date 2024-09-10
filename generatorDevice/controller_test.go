package generatorDevice

import (
	"testing"
	"time"
)

func TestController(t *testing.T) {
	config := Configuration{
		InStateResolution:        1 * time.Millisecond,
		PrimingTimeout:           10 * time.Millisecond,
		CrankingTmeout:           20 * time.Millisecond,
		WarmUpTimeout:            30 * time.Millisecond,
		WarmUpTemp:               40,
		EngineCoolDownTimeout:    40 * time.Millisecond,
		EngineCoolDownTemp:       50,
		EnclosureCoolDownTimeout: 50 * time.Millisecond,
		EnclosureCoolDownTemp:    60,
		IOCheck: func(i Inputs) bool {
			return i.IOAvailable
		},
		OutputCheck: func(i Inputs) bool {
			return i.OutputAvailable
		},
	}

	t.Run("simpleSuccessfulRun", func(t *testing.T) {
		c := NewController(config)
		c.Run()
		defer close(c.ChangeInput)

		expectNewState(t, c, Off)

		c.ChangeInput <- func(i Inputs) Inputs {
			i.IOAvailable = true
			return i
		}

		expectNewState(t, c, Ready)

		c.ChangeInput <- func(i Inputs) Inputs {
			i.ArmSwitch = true
			return i
		}

		expectNoUpdate(t, c)

	})
}

func expectNoUpdate(t *testing.T, c *Controller) {
	t.Helper()

	// we expect the controller to update the state within 1ms
	time.Sleep(1 * time.Millisecond)

	select {
	case u := <-c.Update:
		t.Errorf("expected no update but got %v", u)
	default:
		return
	}
}

func expectUpdate(t *testing.T, c *Controller) Combined {
	t.Helper()

	// we expect the controller to update the state within 1ms
	timeout := time.NewTimer(time.Millisecond)
	defer timeout.Stop()

	select {
	case u, ok := <-c.Update:
		if !ok {
			t.Errorf("update channel closed")
		}
		return u
	case <-timeout.C:
		t.Errorf("expected update but got nothing")
	}
	return Combined{}
}

func expectNewState(t *testing.T, c *Controller, s State) {
	t.Helper()
	u := expectUpdate(t, c)
	if u.State != s {
		t.Errorf("expected state %v but got %v", s, u.State)
	}
}
