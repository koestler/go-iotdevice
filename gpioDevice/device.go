package gpioDevice

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/device"
	"golang.org/x/exp/maps"
	"log"
	"periph.io/x/conn/v3/gpio"
	"time"
)

type Config interface {
	Inputs() []Pin
	Outputs() []Pin
	PollInterval() time.Duration
}

type Pin interface {
	Pin() string
	Name() string
	Description() string
	LowLabel() string
	HighLabel() string
}

type DeviceStruct struct {
	device.State
	gpioConfig Config

	commandStorage *dataflow.ValueStorage
}

func NewDevice(
	deviceConfig device.Config,
	gpioConfig Config,
	stateStorage *dataflow.ValueStorage,
	commandStorage *dataflow.ValueStorage,
) *DeviceStruct {
	return &DeviceStruct{
		State: device.NewState(
			deviceConfig,
			stateStorage,
		),
		gpioConfig:     gpioConfig,
		commandStorage: commandStorage,
	}
}

func (d *DeviceStruct) Run(ctx context.Context) (err error, immediateError bool) {
	if err = hostInitOnce(); err != nil {
		return fmt.Errorf("gpioDevice: host init failed: %w", err), true
	}

	dName := d.Config().Name()

	// setup registers
	inpRegisters, err := pinToRegisterMap(d.gpioConfig.Inputs(), "Inputs", 0, false)
	if err != nil {
		return fmt.Errorf("gpioDevice[%s]: input setup failed: %w", dName, err), true
	}
	oupRegisters, err := pinToRegisterMap(d.gpioConfig.Outputs(), "Outputs", 100, true)
	if err != nil {
		return fmt.Errorf("gpioDevice[%s]: output setup failed: %w", dName, err), true
	}
	addToRegisterDb(d.State.RegisterDb(), inpRegisters)
	addToRegisterDb(d.State.RegisterDb(), oupRegisters)

	// also fetch output registers
	maps.Copy(inpRegisters, oupRegisters)

	// fetch initial state
	d.execPoll(inpRegisters)

	// send connected now, disconnected when this routine stops
	d.SetAvailable(true)
	defer func() {
		d.SetAvailable(false)
	}()

	// setup subscription to listen for updates of writable registers
	_, commandSubscription := d.commandStorage.SubscribeReturnInitial(ctx, dataflow.DeviceNonNullValueFilter(dName))

	ticker := time.NewTicker(d.gpioConfig.PollInterval())
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, false
		case <-ticker.C:
			d.execPoll(inpRegisters)
		case value := <-commandSubscription.Drain():
			d.execCommand(oupRegisters, value)
		}
	}
}

func (d *DeviceStruct) execPoll(inpRegisters map[string]GpioRegister) {
	start := time.Now()

	// fetch registers
	for _, reg := range inpRegisters {
		value := 0
		if reg.pin.Read() {
			value = 1
		}

		d.StateStorage().Fill(dataflow.NewEnumRegisterValue(
			d.Name(),
			reg,
			value,
		))
	}

	if d.Config().LogDebug() {
		log.Printf(
			"gpioDevice[%s]: registers fetched, took=%.3fs",
			d.Name(),
			time.Since(start).Seconds(),
		)
	}
}

func (d *DeviceStruct) execCommand(oupRegisters map[string]GpioRegister, value dataflow.Value) {
	dName := d.Config().Name()

	if d.Config().LogDebug() {
		log.Printf("gpioDevice[%s]: value command: %s", dName, value.String())
	}

	reg, ok := oupRegisters[value.Register().Name()]
	if !ok {
		log.Printf("gpioDevice[%s]: register ignored: %s", dName, value.Register().Name())
		return
	}

	enumValue, ok := value.(dataflow.EnumRegisterValue)
	if !ok {
		// ignore non enum values
		return
	}

	var command gpio.Level
	switch enumValue.EnumIdx() {
	case 0:
		command = gpio.Low
	case 1:
		command = gpio.High
	default:
		return
	}

	if d.Config().LogDebug() {
		log.Printf("gpioDevice[%s]: write register=%s, pin=%s, level: %s",
			dName, reg.Name(), reg.pin.Name(), command,
		)
	}

	if err := reg.pin.Out(command); err != nil {
		log.Printf("gpioDevice[%s]: write failed: %s", dName, err)
	} else {
		// set the current state immediately after a successful write
		d.StateStorage().Fill(dataflow.NewEnumRegisterValue(
			dName,
			value.Register(),
			enumValue.EnumIdx(),
		))

		if d.Config().LogDebug() {
			log.Printf("gpioDevice[%s]: command request successful", dName)
		}
	}

	// reset the command; this allows the same command (e.g. toggle) to be sent again
	d.commandStorage.Fill(dataflow.NewNullRegisterValue(dName, value.Register()))
}

func (d *DeviceStruct) Model() string {
	return "Gpio Device"
}
