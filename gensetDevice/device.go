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

func (c *DeviceStruct) Run(ctx context.Context) (err error, immediateError bool) {
	c.controller = genset.NewController(
		genset.Params{
			// Transition params
			PrimingTimeout:           c.gensetConfig.PrimingTimeout(),
			CrankingTimeout:          c.gensetConfig.CrankingTimeout(),
			WarmUpTimeout:            c.gensetConfig.WarmUpTimeout(),
			WarmUpMinTime:            c.gensetConfig.WarmUpMinTime(),
			WarmUpTemp:               c.gensetConfig.WarmUpTemp(),
			EngineCoolDownTimeout:    c.gensetConfig.EngineCoolDownTimeout(),
			EngineCoolDownTemp:       c.gensetConfig.EngineCoolDownTemp(),
			EnclosureCoolDownTimeout: c.gensetConfig.EnclosureCoolDownTimeout(),
			EnclosureCoolDownTemp:    c.gensetConfig.EnclosureCoolDownTemp(),

			// IO Check
			EngineTempMin: c.gensetConfig.EngineTempMin(),
			EngineTempMax: c.gensetConfig.EngineTempMax(),
			AuxTemp0Min:   c.gensetConfig.AuxTemp0Min(),
			AuxTemp0Max:   c.gensetConfig.AuxTemp0Max(),
			AuxTemp1Min:   c.gensetConfig.AuxTemp1Min(),
			AuxTemp1Max:   c.gensetConfig.AuxTemp1Max(),

			// Output Check
			SinglePhase: c.gensetConfig.SinglePhase(),
			UMin:        c.gensetConfig.UMin(),
			UMax:        c.gensetConfig.UMax(),
			FMin:        c.gensetConfig.FMin(),
			FMax:        c.gensetConfig.FMax(),
			PMax:        c.gensetConfig.PMax(),
			PTotMax:     c.gensetConfig.PTotMax(),
		},
		genset.Off,
		genset.Inputs{},
	)

	// bind inputs
	for _, b := range c.gensetConfig.InputBindings() {
		deviceName := b.DeviceName()
		registerName := b.RegisterName()

		sub := c.StateStorage().SubscribeSendInitial(ctx, func(v dataflow.Value) bool {
			return v.DeviceName() == deviceName && v.Register().Name() == registerName
		})

		setter, err := inpSetter(b.Name())
		if err != nil {
			return fmt.Errorf("gensetDevice[%s]: input setter failed: %s", c.Name(), err), true
		}

		go func() {
			// routine will return when ctx of the subscription is cancelled
			for v := range sub.Drain() {
				setter(c.controller, v)
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
				c.controller.UpdateInputs(func(i genset.Inputs) genset.Inputs {
					i.Time = t
					return i
				})
			}
		}
	}()

	// start the controller
	c.controller.Run()
	defer c.controller.End()

	// stream outputs

	// wait for context to be done
	<-ctx.Done()

	return nil, false
}

func (c *DeviceStruct) Model() string {
	return "Genset Controller"
}
