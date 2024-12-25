package gpioDevice

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/device"
	"github.com/warthog618/go-gpiocdev"
	"golang.org/x/exp/maps"
	"log"
	"slices"
)

func NewDevice(
	deviceConfig device.Config,
	gpioConfig Config,
	stateStorage *dataflow.ValueStorage,
	commandStorage *dataflow.ValueStorage,
) (*DeviceStruct, error) {
	return &DeviceStruct{
		State: device.NewState(
			deviceConfig,
			stateStorage,
		),
		gpioConfig:     gpioConfig,
		commandStorage: commandStorage,
	}, nil
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

	// watch inputs
	lines, err := d.setupInputs(chip, inpRegisters)
	if err != nil {
		return fmt.Errorf("gpioDevice[%s]: setup inputs failed: %w", dName, err), true
	}
	defer func() {
		if err := lines.Close(); err != nil {
			log.Printf("gpioDevice[%s]: error while closing lines: %s", d.Config().Name(), err)
		}
	}()

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

// setupInputs configures the given registers as inputs and registers an event listener.
// The caller shall close the returned lines whenever there is no error returned.
func (d *DeviceStruct) setupInputs(chip *gpiocdev.Chip, regMap map[string]GpioRegister) (*gpiocdev.Lines, error) {
	regList := maps.Values(regMap)
	offsets := offsetList(regList)

	if d.Config().LogDebug() {
		for i, reg := range regList {
			log.Printf("gpioDevice[%s]: setup input register: %s, offset=%d", d.Name(), reg, offsets[i])
		}
	}

	// configure as input and set additional options
	opts := []gpiocdev.LineReqOption{
		gpiocdev.WithBothEdges,
		gpiocdev.WithEventHandler(d.eventHandler(regList)),
	}
	if d := d.gpioConfig.InputDebounce(); d > 0 {
		opts = append(opts, gpiocdev.WithDebounce(d))
	}
	{
		inputOpts := d.gpioConfig.InputOptions()
		slices.Sort(inputOpts)
		for _, o := range slices.Compact(inputOpts) {
			switch o {
			case "WithBiasDisabled":
				opts = append(opts, gpiocdev.WithBiasDisabled)
				break
			case "WithPullDown":
				opts = append(opts, gpiocdev.WithPullDown)
				break
			case "WithPullUp":
				opts = append(opts, gpiocdev.WithPullUp)
				break
			}
		}
	}

	lines, err := chip.RequestLines(offsets, opts...)
	if err != nil {
		return nil, fmt.Errorf("request inputs lines failed: %w", err)
	}

	// fetch initial values
	values := make([]int, len(offsets))
	err = lines.Values(values)
	if err != nil {
		if err := lines.Close(); err != nil {
			log.Printf("gpioDevice[%s]: error while closing lines: %s", d.Config().Name(), err)
		}
		return nil, fmt.Errorf("fetch initial values failed: %w", err)
	}

	// send initial state to the state storage
	for i, reg := range regList {
		v := values[i]

		if d.Config().LogDebug() {
			log.Printf("gpioDevice[%s]: read input register %s, value=%d", d.Name(), reg, v)
		}

		if !isValidValue(v) {
			log.Printf("gpioDevice[%s]: ignoring invalid value for input register %s, value=%d", d.Name(), reg, v)
			continue
		}

		d.StateStorage().Fill(dataflow.NewEnumRegisterValue(d.Name(), reg, v))
	}

	// do not close lines, the caller should do this
	return lines, nil
}

func (d *DeviceStruct) eventHandler(regList []GpioRegister) func(e gpiocdev.LineEvent) {
	offsetToRegMap := make(map[int]GpioRegister, len(regList))
	for _, reg := range regList {
		offsetToRegMap[reg.offset] = reg
	}

	return func(e gpiocdev.LineEvent) {
		reg, ok := offsetToRegMap[e.Offset]
		if !ok {
			log.Printf("gpioDevice[%s]: register not found for offset %d", d.Name(), e.Offset)
			return
		}

		v := 0
		if e.Type == gpiocdev.LineEventRisingEdge {
			v = 1
		}

		if d.Config().LogDebug() {
			log.Printf("gpioDevice[%s]: set input: register %s, value=%v", d.Name(), reg, v)
		}

		d.StateStorage().Fill(dataflow.NewEnumRegisterValue(d.Name(), reg, v))
	}
}

func (d *DeviceStruct) setupOutputs(chip *gpiocdev.Chip, regMap map[string]GpioRegister) error {
	// generate ordered list of registers and offsets
	regList := maps.Values(regMap)
	offsets := offsetList(regList)

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
			log.Printf("gpioDevice[%s]: read output register %s, value=%d", d.Name(), reg, v)
		}

		if !isValidValue(v) {
			log.Printf("gpioDevice[%s]: ignoring invalid value for output register %s, value=%d", d.Name(), reg, v)
			continue
		}

		d.StateStorage().Fill(dataflow.NewEnumRegisterValue(d.Name(), reg, v))
	}

	// configure as output, set the initial values, and additional options
	opts := []gpiocdev.LineConfigOption{
		gpiocdev.AsOutput(values...),
	}
	{
		outputOpts := d.gpioConfig.OutputOptions()
		slices.Sort(outputOpts)
		for _, o := range slices.Compact(outputOpts) {
			switch o {
			case "AsOpenDrain":
				opts = append(opts, gpiocdev.AsOpenDrain)
				break
			case "AsOpenSource":
				opts = append(opts, gpiocdev.AsOpenSource)
				break
			case "AsPushPull":
				opts = append(opts, gpiocdev.AsPushPull)
				break
			}
		}
	}

	err = lines.Reconfigure(opts...)
	if err != nil {
		return fmt.Errorf("reconfigure as output failed: %w", err)
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

func offsetList(regList []GpioRegister) []int {
	offsets := make([]int, len(regList))
	for i, reg := range regList {
		offsets[i] = reg.offset
	}
	return offsets
}
