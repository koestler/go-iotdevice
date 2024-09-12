package generator_test

import (
	"fmt"
	"github.com/koestler/go-iotdevice/v3/generator"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestController(t *testing.T) {
	params := generator.Params{
		PrimingTimeout:           10 * time.Second,
		CrankingTimeout:          20 * time.Second,
		WarmUpTimeout:            10 * time.Minute,
		WarmUpTemp:               40,
		EngineCoolDownTimeout:    5 * time.Minute,
		EngineCoolDownTemp:       60,
		EnclosureCoolDownTimeout: 15 * time.Minute,
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

	t0, _ := time.Parse(time.RFC3339, "2000-01-01T00:00:00Z")
	initialInputs := generator.Inputs{Time: t0}

	t.Run("simpleSuccessfulRun", func(t *testing.T) {
		c := generator.NewController(params, generator.Off, initialInputs)

		stateTracker := newTracker[generator.State](t)
		c.OnStateUpdate = stateTracker.OnUpdateFunc()
		outputTracker := newTracker[generator.Outputs](t)
		c.OnOutputUpdate = outputTracker.OnUpdateFunc()

		c.Run()
		defer c.End()

		setInp := func(f func(i generator.Inputs) generator.Inputs) {
			c.UpdateInputsSync(func(i generator.Inputs) generator.Inputs {
				i = f(i)
				t.Logf("inputs: %v", i)
				return i
			})
		}

		// initial state
		stateTracker.AssertLatest(t, generator.State{Node: generator.Off, Changed: t0})
		outputTracker.AssertLatest(t, generator.Outputs{})

		// go to ready
		setInp(func(i generator.Inputs) generator.Inputs {
			i.IOAvailable = true
			i.EngineTemp = 20
			i.AirIntakeTemp = 20
			i.AirExhaustTemp = 20
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.Ready, Changed: t0})
		outputTracker.AssertLatest(t, generator.Outputs{IoCheck: true})

		// go to priming
		setInp(func(i generator.Inputs) generator.Inputs {
			i.ArmSwitch = true
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.Ready, Changed: t0})
		outputTracker.AssertLatest(t, generator.Outputs{IoCheck: true})

		setInp(func(i generator.Inputs) generator.Inputs {
			i.CommandSwitch = true
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.Priming, Changed: t0})
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, Pump: true, IoCheck: true})

		// stay in priming
		t1 := t0.Add(params.PrimingTimeout / 2)
		setInp(func(i generator.Inputs) generator.Inputs {
			i.Time = t1
			return i
		})

		// go to cranking
		t2 := t0.Add(params.PrimingTimeout)
		setInp(func(i generator.Inputs) generator.Inputs {
			i.Time = t2
			return i
		})

		stateTracker.AssertLatest(t, generator.State{Node: generator.Cranking, Changed: t2})
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, Pump: true, Ignition: true, Starter: true, IoCheck: true})

		// go to warm up
		setInp(func(i generator.Inputs) generator.Inputs {
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
		setInp(func(i generator.Inputs) generator.Inputs {
			i.EngineTemp = 45
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.Producing, Changed: t0})
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, Pump: true, Ignition: true, Load: true, IoCheck: true})

		// running, engine getting warm
		setInp(func(i generator.Inputs) generator.Inputs {
			i.EngineTemp = 70
			return i
		})

		// go to engine cool down
		setInp(func(i generator.Inputs) generator.Inputs {
			i.CommandSwitch = false
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.EngineCoolDown, Changed: t0})
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, Pump: true, Ignition: true, IoCheck: true})

		// go to enclosure cool down
		setInp(func(i generator.Inputs) generator.Inputs {
			i.EngineTemp = 55
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.EnclosureCoolDown, Changed: t0})
		outputTracker.AssertLatest(t, generator.Outputs{Fan: true, IoCheck: true})

		// go to ready
		setInp(func(i generator.Inputs) generator.Inputs {
			i.EngineTemp = 45
			return i
		})
		stateTracker.AssertLatest(t, generator.State{Node: generator.Ready, Changed: t0})
		outputTracker.AssertLatest(t, generator.Outputs{})
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
