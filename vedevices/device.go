package vedevices

import (
	"github.com/koestler/go-victron-to-mqtt/config"
	"github.com/koestler/go-victron-to-mqtt/dataflow"
	"github.com/koestler/go-victron-to-mqtt/vedirect"
)

type Config interface {
	Name() string
	Kind() config.DeviceKind
	Device() string
	LogDebug() bool
}

type Device struct {
	// configuration
	cfg Config

	source    *dataflow.Source
	product   *vedirect.VeProduct
	registers Registers
}

func RunDevice(cfg Config, target dataflow.Fillable) (device *Device, err error) {
	err, source, product, registers := CreateSource(cfg)
	if err != nil {
		return nil, err
	}

	// pipe all data to next stage
	source.Append(target)

	Device := &Device{
		cfg:       cfg,
		source:    source,
		product:   product,
		registers: registers,
	}
	return Device, nil
}

func (c *Device) Shutdown() {
	// todo: implement proper shutdown
}

func (c *Device) Name() string {
	return c.cfg.Name()
}

func (c *Device) Config() Config {
	return c.cfg
}
func (c *Device) Registers() Registers {
	return c.registers
}
