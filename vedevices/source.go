package vedevices

import (
	"fmt"
	"github.com/koestler/go-victron-to-mqtt/config"
	"github.com/koestler/go-victron-to-mqtt/dataflow"
	"github.com/koestler/go-victron-to-mqtt/vedirect"
	"log"
	"math/rand"
	"time"
)

func CreateSource(cfg Config) (err error, source *dataflow.Source, product *vedirect.VeProduct) {
	switch cfg.Kind() {
	case config.RandomBmvKind:
		return CreateRandomSource(cfg.Name(), RegisterListBmv702)
	case config.RandomSolarKind:
		return CreateRandomSource(cfg.Name(), RegisterListSolar)
	case config.VedirectKind:
		return CreateVedirectSource(cfg)
	default:
		return fmt.Errorf("unknown kind: %s", cfg.Kind().String()), nil, nil
	}
}

func CreateRandomSource(deviceName string, registers Registers) (err error, source *dataflow.Source, product *vedirect.VeProduct) {
	// setup output chain
	output := make(chan dataflow.Value)

	// start source go routine
	go func() {
		// todo: add shutdown channel
		defer close(output)
		for _ = range time.Tick(time.Second) {
			for name, register := range registers {
				output <- dataflow.Value{
					DeviceName:    deviceName,
					Name:          name,
					Value:         1e2 * rand.Float64() * register.Factor,
					Unit:          register.Unit,
					RoundDecimals: register.RoundDecimals,
				}
			}
		}
	}()

	// return data source
	return nil, dataflow.CreateSource(output), nil
}

func CreateVedirectSource(cfg Config) (err error, source *dataflow.Source, product *vedirect.VeProduct) {
	// open vedirect device
	vd, err := vedirect.Open(cfg.Device())
	if err != nil {
		return err, nil, nil
	}

	// send ping
	if err := vd.VeCommandPing(); err != nil {
		return fmt.Errorf("Ping failed: %s", err), nil, nil
	}

	// get deviceId
	deviceId, err := vd.VeCommandDeviceId()
	if err != nil {
		return fmt.Errorf("Cannot get DeviceId: %s", err), nil, nil
	}

	deviceString := deviceId.String()
	if len(deviceString) < 1 {
		return fmt.Errorf("unknown deviceId=%x", err), nil, nil
	}
	log.Printf("device[%s]: source: connect to %s", cfg.Name(), deviceString)

	// get relevant registers
	registers := RegisterFactoryByProduct(deviceId)
	if registers == nil {
		return fmt.Errorf("no registers found for deviceId=%x", deviceId), nil, nil
	}

	// setup output chain with enough capacity to hold some values
	capacity := len(registers) / 4
	if capacity < 4 {
		capacity = 4
	}
	output := make(chan dataflow.Value, capacity)

	// start vedevices reader
	go func() {
		defer close(output)
		// flush buffer
		vd.RecvFlush()

		for _ = range time.Tick(100 * time.Millisecond) {
			if err := vd.VeCommandPing(); err != nil {
				log.Printf("device[%s]: source: VeCommandPing failed: %v", cfg.Name(), err)
				continue
			}

			for name, register := range registers {
				if numericValue, err := register.RecvNumeric(vd); err != nil {
					log.Printf("device[%s]: vedevices.RecvNumeric failed: %v", cfg.Name(), err)
				} else {
					output <- dataflow.Value{
						DeviceName:    cfg.Name(),
						Name:          name,
						Value:         numericValue.Value,
						Unit:          numericValue.Unit,
						RoundDecimals: register.RoundDecimals,
					}
				}
			}
		}
	}()

	// return data source
	return nil, dataflow.CreateSource(output), &deviceId
}
