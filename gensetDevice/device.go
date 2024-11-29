package gensetDevice

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/device"
	"github.com/koestler/go-iotdevice/v3/genset"
	"time"
)

type Config interface {
	InputBindings() []Binding
	OutputBindings() []Binding

	PrimingTimeout() time.Duration
	CrankingTimeout() time.Duration
	WarmUpTimeout() time.Duration
	WarmUpMinTime() time.Duration
	WarmUpTemp() float64
	EngineCoolDownTimeout() time.Duration
	EngineCoolDownTemp() float64
	EnclosureCoolDownTimeout() time.Duration
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

type DeviceStruct struct {
	device.State
	gensetConfig Config

	commandStorage *dataflow.ValueStorage
	controller     *genset.Controller
}

func NewDevice(
	deviceConfig device.Config,
	gensetConfig Config,
	stateStorage *dataflow.ValueStorage,
	commandStorage *dataflow.ValueStorage,
) *DeviceStruct {
	return &DeviceStruct{
		State: device.NewState(
			deviceConfig,
			stateStorage,
		),
		gensetConfig:   gensetConfig,
		commandStorage: commandStorage,
	}
}

func (d *DeviceStruct) Run(ctx context.Context) (err error, immediateError bool) {
	addToRegisterDb(d.State.RegisterDb(), d.gensetConfig.SinglePhase())

	d.controller = genset.NewController(
		genset.Params{
			// Transition params
			PrimingTimeout:           d.gensetConfig.PrimingTimeout(),
			CrankingTimeout:          d.gensetConfig.CrankingTimeout(),
			WarmUpTimeout:            d.gensetConfig.WarmUpTimeout(),
			WarmUpMinTime:            d.gensetConfig.WarmUpMinTime(),
			WarmUpTemp:               d.gensetConfig.WarmUpTemp(),
			EngineCoolDownTimeout:    d.gensetConfig.EngineCoolDownTimeout(),
			EngineCoolDownTemp:       d.gensetConfig.EngineCoolDownTemp(),
			EnclosureCoolDownTimeout: d.gensetConfig.EnclosureCoolDownTimeout(),
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
		genset.Off,
		genset.Inputs{},
	)

	// bind inputs
	for _, b := range d.gensetConfig.InputBindings() {
		deviceName := b.DeviceName()
		registerName := b.RegisterName()

		sub := d.StateStorage().SubscribeSendInitial(ctx, func(v dataflow.Value) bool {
			return v.DeviceName() == deviceName && v.Register().Name() == registerName
		})

		setter, err := d.inpSetter(b.Name())
		if err != nil {
			return fmt.Errorf("gensetDevice[%s]: input setter failed: %s", d.Name(), err), true
		}

		go func() {
			// routine will return when ctx of the subscription is cancelled
			for v := range sub.Drain() {
				setter(d.controller, v)
			}
		}()
	}

	// update time input
	go func() {
		ticker := time.NewTicker(time.Second)
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
	d.controller.OnStateUpdate = func(s genset.State) {
		name := d.Config().Name()
		ss := d.StateStorage()

		ss.Fill(dataflow.NewEnumRegisterValue(name, StateRegister, int(s.Node)))
		ss.Fill(dataflow.NewTextRegisterValue(name, StateChangedRegister, s.Changed.String()))
	}

	// handle output updates
	d.controller.OnOutputUpdate = func(o genset.Outputs) {
		name := d.Config().Name()
		ss := d.StateStorage()

		ss.Fill(dataflow.NewEnumRegisterValue(name, IgnitionRegister, boolToOnOff(o.Ignition)))
		ss.Fill(dataflow.NewEnumRegisterValue(name, StarterRegister, boolToOnOff(o.Starter)))
		ss.Fill(dataflow.NewEnumRegisterValue(name, FanRegister, boolToOnOff(o.Fan)))
		ss.Fill(dataflow.NewEnumRegisterValue(name, PumpRegister, boolToOnOff(o.Pump)))
		ss.Fill(dataflow.NewEnumRegisterValue(name, LoadRegister, boolToOnOff(o.Load)))
		ss.Fill(dataflow.NewNumericRegisterValue(name, TimeInStateRegister, o.TimeInState.Seconds()))
		ss.Fill(dataflow.NewEnumRegisterValue(name, IoCheckRegister, boolToOnOff(o.IoCheck)))
		ss.Fill(dataflow.NewEnumRegisterValue(name, OutputCheckRegister, boolToOnOff(o.OutputCheck)))
	}

	// start the controller
	d.controller.Run()
	defer d.controller.End()

	// wait for context to be done
	<-ctx.Done()

	return nil, false
}

func (d *DeviceStruct) Model() string {
	return "Genset Controller"
}
