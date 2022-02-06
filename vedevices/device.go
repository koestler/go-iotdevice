package vedevices

import (
	"fmt"
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/vedirect"
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
	registers Registers

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
		return device, device.RunRandom(output, RegisterListBmv702)
	case config.RandomSolarKind:
		return device, device.RunRandom(output, RegisterListSolar)
	case config.VedirectKind:
		return device, device.RunVedirect(output)
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
func (c *Device) Registers() Registers {
	return c.registers
}

func (c *Device) RunRandom(output chan dataflow.Value, registers Registers) (err error) {
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

func (c *Device) RunVedirect(output chan dataflow.Value) (err error) {
	log.Printf("device[%s]: start vedirect source", c.cfg.Name())

	// open vedirect device
	vd, err := vedirect.Open(c.cfg.Device(), c.cfg.LogComDebug())
	if err != nil {
		return err
	}

	// send ping
	if err := vd.VeCommandPing(); err != nil {
		return fmt.Errorf("ping failed: %s", err)
	}

	// get deviceId
	deviceId, err := vd.VeCommandDeviceId()
	if err != nil {
		return fmt.Errorf("cannot get DeviceId: %s", err)
	}

	deviceString := deviceId.String()
	if len(deviceString) < 1 {
		return fmt.Errorf("unknown deviceId=%x", err)
	}

	log.Printf("device[%s]: source: connect to %s", c.cfg.Name(), deviceString)

	// get relevant registers
	registers := RegisterFactoryByProduct(deviceId)
	if registers == nil {
		return fmt.Errorf("no registers found for deviceId=%x", deviceId)
	}

	// start vedevices reader
	go func() {
		defer close(c.closed)
		defer close(output)

		// flush buffer
		vd.RecvFlush()

		ticker := time.NewTicker(100 * time.Millisecond)

		for {
			select {
			case <-c.shutdown:
				return
			case <-ticker.C:
				start := time.Now()

				if err := vd.VeCommandPing(); err != nil {
					log.Printf("device[%s]: source: VeCommandPing failed: %v", c.cfg.Name(), err)
					continue
				}

				for _, register := range registers {
					if numericValue, err := register.RecvNumeric(vd); err != nil {
						log.Printf("device[%s]: vedevices.RecvNumeric failed: %v", c.cfg.Name(), err)
					} else {
						output <- dataflow.Value{
							DeviceName:    c.cfg.Name(),
							Name:          register.Name,
							Value:         numericValue.Value,
							Unit:          numericValue.Unit,
							RoundDecimals: 3,
						}
					}
				}

				if c.cfg.LogDebug() {
					log.Printf(
						"device[%s]: registers fetched, took=%.3fs",
						c.cfg.Name(),
						time.Since(start).Seconds(),
					)
				}
			}
		}
	}()

	return
}
