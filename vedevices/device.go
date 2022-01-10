package vedevices

import (
	"github.com/koestler/go-victron-to-mqtt/config"
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
	config Config

	product vedirect.VeProduct
}

func RunDevice(config Config) (*Device, error) {
	Device := &Device{
		config: config,
	}

	return Device, nil
}

func (c *Device) Shutdown() {
	// todo: implement proper shutdown
}

func (c *Device) Name() string {
	return c.config.Name()
}

func (c *Device) Config() Config {
	return c.config
}
