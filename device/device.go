package device

import (
	"fmt"
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/vedirect"
	"github.com/koestler/go-iotdevice/victron"
	"log"
	"math/rand"
	"time"
)

type Config interface {
	Name() string
	Kind() config.DeviceKind
	Device() string
	LogDebug() bool
	LogComDebug() bool
}

type Device struct {
	// configuration
	cfg Config

	source    *dataflow.Source
	product   *vedirect.VeProduct
	registers dataflow.Registers

	shutdown chan struct{}
	closed   chan struct{}
}

func RunDevice(cfg Config, target dataflow.Fillable) (device *Device, err error) {
	// setup output chain
	output := make(chan dataflow.Value, 32)
	source := dataflow.CreateSource(output)
	// pipe all data to next stage
	source.Append(target)

	device = &Device{
		cfg:      cfg,
		source:   source,
		shutdown: make(chan struct{}),
		closed:   make(chan struct{}),
	}

	switch cfg.Kind() {
	case config.RandomBmvKind:
		return device, device.RunRandom(output, victron.RegisterListBmv702)
	case config.RandomSolarKind:
		return device, device.RunRandom(output, victron.RegisterListSolar)
	case config.VedirectKind:
		return device, victron.RunVictron(device, output)
	default:
		return nil, fmt.Errorf("unknown kind: %s", cfg.Kind().String())
	}
}

func (c *Device) Shutdown() {
	close(c.shutdown)
	<-c.closed
	log.Printf("device[%s]: shutdown completed", c.cfg.Name())
}

func (c *Device) Name() string {
	return c.cfg.Name()
}

func (c *Device) Config() Config {
	return c.cfg
}
func (c *Device) Registers() dataflow.Registers {
	return c.registers
}

func (c *Device) RunRandom(output chan dataflow.Value, registers dataflow.Registers) (err error) {
	if c.cfg.LogDebug() {
		log.Printf("device[%s]: start random source", c.cfg.Name())
	}

	// start source go routine
	go func() {
		defer close(c.closed)
		defer close(output)

		ticker := time.NewTicker(time.Second)

		for {
			select {
			case <-c.shutdown:
				return
			case <-ticker.C:
				for _, register := range registers {
					output <- dataflow.Value{
						DeviceName:    c.cfg.Name(),
						Name:          register.Name,
						Value:         1e2 * rand.Float64() * register.Factor,
						Unit:          register.Unit,
						RoundDecimals: 3,
					}
				}
			}
		}
	}()

	return
}
