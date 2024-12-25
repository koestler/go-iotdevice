package gpioDevice

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/device"
	"github.com/warthog618/go-gpiocdev"
	"golang.org/x/exp/maps"
	"log"
)

type Config interface {
	Chip() string
	Inputs() []Pin
	Outputs() []Pin
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
	dName := d.Config().Name()

	// initialize chip
	var chip *gpiocdev.Chip
	chip, err = gpiocdev.NewChip(d.gpioConfig.Chip(), gpiocdev.WithConsumer("go-iotdevice"))
	if err != nil {
		return fmt.Errorf("gpioDevice[%s]: chip initialization failed: %w", dName, err), true
	} else if d.Config().LogDebug() {
		log.Printf("gpioDevice[%s]: chip '%s' initialized", dName, chip.Name)
	}
	defer func() {
		if err := chip.Close(); err != nil {
			log.Printf("gpioDevice[%s]: error while closing chip: %s", dName, err)
		}
	}()

	// setup registers
	inpRegisters, err := pinToRegisterMap(chip, d.gpioConfig.Inputs(), "Inputs", 0, false)
	if err != nil {
		return fmt.Errorf("gpioDevice[%s]: input setup failed: %w", dName, err), true
	}
	oupRegisters, err := pinToRegisterMap(chip, d.gpioConfig.Outputs(), "Outputs", 100, true)
	if err != nil {
		return fmt.Errorf("gpioDevice[%s]: output setup failed: %w", dName, err), true
	}
	addToRegisterDb(d.State.RegisterDb(), inpRegisters)
	addToRegisterDb(d.State.RegisterDb(), oupRegisters)

	// setup inputs
	/*
		for _, reg := range inpRegisters {
			go func() {
				if d.Config().LogDebug() {
					log.Printf("gpioDevice[%s]: setup input register %s",			dName, reg,					)
				}

				l, err := gpiocdev.RequestLine(
					reg.chip, reg.offset,
					gpiocdev.WithPullUp, // todo: make bias configurable
					gpiocdev.WithBothEdges,
					gpiocdev.WithEventHandler(eventHandler))
				if err != nil {
					fmt.Printf("RequestLine returned error: %s\n", err)
					if err == syscall.Errno(22) {
						fmt.Println("Note that the WithPullUp option requires Linux 5.5 or later - check your kernel version.")
					}
					os.Exit(1)
				}
				defer l.Close()


			}()
		}
	*/

	// initial read output registers and configure them as outputs
	err = d.setupOutputs(chip, oupRegisters)
	if err != nil {
		return fmt.Errorf("gpioDevice[%s]: setup outputs failed: %w", dName, err), true
	}

	// send connected now, disconnected when this routine stops
	d.SetAvailable(true)
	defer func() {
		d.SetAvailable(false)
	}()

	// setup subscription to listen for updates of writable registers
	_, commandSubscription := d.commandStorage.SubscribeReturnInitial(ctx, dataflow.DeviceNonNullValueFilter(dName))

	for {
		select {
		case <-ctx.Done():
			return nil, false
		case value := <-commandSubscription.Drain():
			d.execCommand(chip, oupRegisters, value)
		}
	}
}

func (d *DeviceStruct) setupOutputs(chip *gpiocdev.Chip, regMap map[string]GpioRegister) error {
	// generate ordered list of registers
	regList := maps.Values(regMap)

	// compute ordered list of offsets
	offsets := make([]int, len(regList))
	for i, reg := range regList {
		offsets[i] = reg.offset
	}

	// fetch the line values
	lines, err := chip.RequestLines(offsets)
	if err != nil {
		return err
	}
	defer func() {
		if err := lines.Close(); err != nil {
			log.Printf("gpioDevice[%s]: error while closing lines: %s", d.Name(), err)
		}
	}()

	values := make([]int, len(offsets))
	err = lines.Values(values)
	if err != nil {
		return err
	}

	// send initial state to the state storage
	for i, reg := range regList {
		v := values[i]

		if d.Config().LogDebug() {
			log.Printf("gpioDevice[%s]: read %s, value=%d", d.Name(), reg, v)
		}

		if !isValidValue(v) {
			return fmt.Errorf("invalid value %d for register %s", v, reg)
		}

		d.StateStorage().Fill(dataflow.NewEnumRegisterValue(d.Name(), reg, v))
	}

	// configure as output
	err = lines.Reconfigure(gpiocdev.AsOutput(values...))
	if err != nil {
		return fmt.Errorf("reconfigure as output failed: %w", err)
	}

	if d.Config().LogDebug() {
		log.Printf("gpioDevice[%s]: registers setup", d.Name())
	}

	return nil
}

func (d *DeviceStruct) execCommand(chip *gpiocdev.Chip, oupRegisters map[string]GpioRegister, value dataflow.Value) {
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

	v := enumValue.EnumIdx()
	if !isValidValue(v) {
		log.Printf("gpioDevice[%s]: invalid value %d for register %s", dName, v, reg)
		return
	}

	if d.Config().LogDebug() {
		log.Printf("gpioDevice[%s]: write register %s, value=%d", dName, reg, v)
	}

	l, err := chip.RequestLine(reg.offset, gpiocdev.AsOutput(0))
	if err != nil {
		log.Printf("gpioDevice[%s]: request line failed: %s", dName, err)
		return
	}
	defer func() {
		if err := l.Close(); err != nil {
			log.Printf("gpioDevice[%s]: error while closing line: %s", dName, err)
		}
	}()

	err = l.SetValue(v)
	if err != nil {
		log.Printf("gpioDevice[%s]: set register %s, value=%d failed: %s", dName, reg, v, err)
		return
	}

	// set the current state immediately after a successful write
	d.StateStorage().Fill(dataflow.NewEnumRegisterValue(
		dName,
		value.Register(),
		enumValue.EnumIdx(),
	))

	// reset the command; this allows the same command (e.g. toggle) to be sent again
	d.commandStorage.Fill(dataflow.NewNullRegisterValue(dName, value.Register()))
}

func (d *DeviceStruct) Model() string {
	return "Gpio Device"
}
