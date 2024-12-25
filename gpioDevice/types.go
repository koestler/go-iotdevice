package gpioDevice

import (
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/device"
	"time"
)

type Config interface {
	Chip() string
	InputDebounce() time.Duration
	// InputOptions are used to configure the pin.
	// Valid options: WithBiasDisabled, WithPullDown, WithPullUp
	InputOptions() []string
	// OutputOptions are used to configure the pin.
	// Valid options: AsOpenDrain, AsOpenSource, AsPushPull
	OutputOptions() []string
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

func (d *DeviceStruct) Model() string {
	return "Gpio Device"
}
