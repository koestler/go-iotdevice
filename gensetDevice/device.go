package gensetDevice

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/device"
	"github.com/koestler/go-iotdevice/v3/genset"
	"log"
	"sync"
	"time"
)

const clockUpdateInterval = 500 * time.Millisecond

type Config interface {
	InputBindings() []Binding
	OutputBindings() []Binding

	PrimingTimeout() time.Duration
	CrankingTimeout() time.Duration
	WarmUpTimeout() time.Duration
	WarmUpMinTime() time.Duration
	WarmUpTemp() float64
	EngineCoolDownTimeout() time.Duration
	EngineCoolDownMinTime() time.Duration
	EngineCoolDownTemp() float64
	EnclosureCoolDownTimeout() time.Duration
	EnclosureCoolDownMinTime() time.Duration
	EnclosureCoolDownTemp() float64

	EngineTempMin() float64
	EngineTempMax() float64
	AuxTemp0Min() float64
	AuxTemp0Max() float64
	AuxTemp1Min() float64
	AuxTemp1Max() float64

	SinglePhase() bool
	UMin() float64
	UMax() float64
	FMin() float64
	FMax() float64
	PMax() float64
	PTotMax() float64
}

type Binding interface {
	Name() string
	DeviceName() string
	RegisterName() string
}

type RegisterDbOfDeviceFunc func(deviceName string) *dataflow.RegisterDb

type DeviceStruct struct {
	device.State
	gensetConfig Config

	commandStorage     *dataflow.ValueStorage
	registerDbOfDevice RegisterDbOfDeviceFunc
	controller         *genset.Controller
}

func NewDevice(
	deviceConfig device.Config,
	gensetConfig Config,
	stateStorage *dataflow.ValueStorage,
	commandStorage *dataflow.ValueStorage,
	registerDbOfDevice RegisterDbOfDeviceFunc,
) *DeviceStruct {
	return &DeviceStruct{
		State: device.NewState(
			deviceConfig,
			stateStorage,
		),
		gensetConfig:       gensetConfig,
		commandStorage:     commandStorage,
		registerDbOfDevice: registerDbOfDevice,
	}
}

func (d *DeviceStruct) Run(ctx context.Context) (err error, immediateError bool) {
	dName := d.Config().Name()
	ss := d.StateStorage()

	initialState := genset.Off
	d.controller = genset.NewController(
		genset.Params{
			// Transition params
			PrimingTimeout:           d.gensetConfig.PrimingTimeout(),
			CrankingTimeout:          d.gensetConfig.CrankingTimeout(),
			WarmUpTimeout:            d.gensetConfig.WarmUpTimeout(),
			WarmUpMinTime:            d.gensetConfig.WarmUpMinTime(),
			WarmUpTemp:               d.gensetConfig.WarmUpTemp(),
			EngineCoolDownTimeout:    d.gensetConfig.EngineCoolDownTimeout(),
			EngineCoolDownMinTime:    d.gensetConfig.EngineCoolDownMinTime(),
			EngineCoolDownTemp:       d.gensetConfig.EngineCoolDownTemp(),
			EnclosureCoolDownTimeout: d.gensetConfig.EnclosureCoolDownTimeout(),
			EnclosureCoolDownMinTime: d.gensetConfig.EnclosureCoolDownMinTime(),
			EnclosureCoolDownTemp:    d.gensetConfig.EnclosureCoolDownTemp(),

			// IO Check
			EngineTempMin: d.gensetConfig.EngineTempMin(),
			EngineTempMax: d.gensetConfig.EngineTempMax(),
			AuxTemp0Min:   d.gensetConfig.AuxTemp0Min(),
			AuxTemp0Max:   d.gensetConfig.AuxTemp0Max(),
			AuxTemp1Min:   d.gensetConfig.AuxTemp1Min(),
			AuxTemp1Max:   d.gensetConfig.AuxTemp1Max(),

			// Output Check
			SinglePhase: d.gensetConfig.SinglePhase(),
			UMin:        d.gensetConfig.UMin(),
			UMax:        d.gensetConfig.UMax(),
			FMin:        d.gensetConfig.FMin(),
			FMax:        d.gensetConfig.FMax(),
			PMax:        d.gensetConfig.PMax(),
			PTotMax:     d.gensetConfig.PTotMax(),
		},
		initialState,
		genset.Inputs{
			Time: time.Now(),
		},
	)

	var shutdownWg sync.WaitGroup

	// setup registers
	commandRegisters := addToRegisterDb(d.State.RegisterDb(), d.gensetConfig.SinglePhase(), d.gensetConfig.InputBindings())

	// send connected now, disconnected when this routine stops
	d.SetAvailable(true)
	defer func() {
		d.SetAvailable(false)
	}()

	// bind inputs
	for _, b := range d.gensetConfig.InputBindings() {
		deviceName := b.DeviceName()
		registerName := b.RegisterName()

		sub := d.StateStorage().SubscribeSendInitial(ctx, func(v dataflow.Value) bool {
			return v.DeviceName() == deviceName && v.Register().Name() == registerName
		})

		setter, err := d.inpSetter(b.Name())
		if err != nil {
			return fmt.Errorf("gensetDevice[%s]: input setter failed: %s", dName, err), true
		}

		shutdownWg.Add(1)
		go func() {
			defer shutdownWg.Done()
			// routine will return when ctx of the subscription is cancelled
			for v := range sub.Drain() {
				setter(d.controller, v)
			}
		}()
	}

	// setup command subscription
	{
		for _, r := range commandRegisters {
			registerName := r.Name()
			_, sub := d.commandStorage.SubscribeReturnInitial(ctx, func(v dataflow.Value) bool {
				return v.DeviceName() == dName && v.Register().Name() == registerName
			})

			setter, err := d.inpSetter(registerName)
			if err != nil {
				return fmt.Errorf("gensetDevice[%s]: input setter failed: %s", d.Name(), err), true
			}

			shutdownWg.Add(1)
			go func() {
				defer shutdownWg.Done()
				// routine will return when ctx of the subscription is cancelled
				for v := range sub.Drain() {
					log.Printf("gensetDevice[%s]: command %v", dName, v)
					setter(d.controller, v)
				}
			}()
		}
	}

	// update clock
	shutdownWg.Add(1)
	go func() {
		defer shutdownWg.Done()
		ticker := time.NewTicker(clockUpdateInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				d.controller.UpdateInputs(func(i genset.Inputs) genset.Inputs {
					i.Time = t
					return i
				})
			}
		}
	}()

	// handle state updates
	lastState := initialState
	d.controller.OnStateUpdate = func(s genset.State) {
		if lastState != s.Node {
			log.Printf("gensetDevice[%s]: state changed: %s", dName, s.Node)
			ss.Fill(dataflow.NewEnumRegisterValue(dName, StateRegister, int(s.Node)))
		}
		ss.Fill(dataflow.NewTextRegisterValue(dName, StateChangedRegister, s.Changed.Format(time.RFC1123)))
	}

	// handle output updates
	d.controller.OnOutputUpdate = func(o genset.Outputs) {
		ss.Fill(dataflow.NewEnumRegisterValue(dName, IgnitionRegister, boolToOnOff(o.Ignition)))
		ss.Fill(dataflow.NewEnumRegisterValue(dName, StarterRegister, boolToOnOff(o.Starter)))
		ss.Fill(dataflow.NewEnumRegisterValue(dName, FanRegister, boolToOnOff(o.Fan)))
		ss.Fill(dataflow.NewEnumRegisterValue(dName, PumpRegister, boolToOnOff(o.Pump)))
		ss.Fill(dataflow.NewEnumRegisterValue(dName, LoadRegister, boolToOnOff(o.Load)))
		ss.Fill(dataflow.NewNumericRegisterValue(dName, TimeInStateRegister, o.TimeInState.Seconds()))
		ss.Fill(dataflow.NewEnumRegisterValue(dName, IoCheckRegister, boolToOnOff(o.IoCheck)))
		ss.Fill(dataflow.NewEnumRegisterValue(dName, OutputCheckRegister, boolToOnOff(o.OutputCheck)))
	}

	// start the controller
	d.controller.Run()
	defer d.controller.End()

	// after context got cancelled, all goroutines will stop. Wait for that to finish
	shutdownWg.Wait()

	return nil, false
}

func (d *DeviceStruct) Model() string {
	return "Genset Controller"
}
