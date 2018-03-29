package vedevices

import (
	"github.com/koestler/go-ve-sensor/dataflow"
	"time"
	"math/rand"
	"log"
	"github.com/koestler/go-ve-sensor/vedirect"
	"github.com/koestler/go-ve-sensor/config"
	"github.com/koestler/go-ve-sensor/storage"
	"fmt"
	"errors"
)

func CreateDummySource(device *storage.Device, config *config.VedeviceConfig) (*dataflow.Source) {
	// get relevant registers
	registers := RegisterFactoryByModel(config.Model);

	// setup output chain
	output := make(chan dataflow.Value)

	// start source go routine
	go func() {
		defer close(output)
		for _ = range time.Tick(time.Second) {
			for name, register := range registers {
				output <- dataflow.Value{
					Device:        device,
					Name:          name,
					Value:         1e2 * rand.Float64() * register.Factor,
					Unit:          register.Unit,
					RoundDecimals: register.RoundDecimals,
				}
			}
		}
	}()

	// return data source
	return dataflow.CreateSource(output)
}

func CreateSource(device *storage.Device, config *config.VedeviceConfig) (err error, source *dataflow.Source) {
	// open vedirect device
	vd, err := vedirect.Open(config.Device)
	if err != nil {
		return err, nil
	}

	// send ping
	if err := vd.VeCommandPing(); err != nil {
		log.Printf("vedevices source: VeCommandPing failed: %v", err)
		return err, nil
	}

	// get deviceId
	deviceId, err := vd.VeCommandDeviceId();
	if err != nil {
		log.Printf("vedevices source: VeCommandDeviceId failed: %v", err)
		return err, nil
	}

	product := deviceId.String()
	if len(product) < 1 {
		log.Printf("vedevices source: unknown deviceId=%x", deviceId)
		return err, nil
	}

	log.Printf("vedevices source: setup product=%v", product)

	// get relevant registers
	registers := RegisterFactoryByProduct(deviceId);
	if registers == nil {
		log.Printf("vedevices source: no registers found for deviceId=%x", deviceId)
		return errors.New(fmt.Sprintf("no registers found for deviceId=%x", deviceId)), nil
	}

	// setup output chain with enough space to hold some values
	output := make(chan dataflow.Value, len(registers)/4)

	// start vedevices reader
	go func() {
		defer close(output)
		// flush buffer
		vd.RecvFlush()

		for _ = range time.Tick(100*time.Millisecond) {
			if err := vd.VeCommandPing(); err != nil {
				log.Printf("vedevices source: VeCommandPing failed: %v", err)
				continue
			}

			for name, register := range registers {
				if numericValue, err := register.RecvNumeric(vd); err != nil {
					log.Printf(
						"device: vedevices.RecvNumeric failed device=%v name=%v err=%v", device, name, err,
					)
				} else {
					output <- dataflow.Value{
						Device:        device,
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
	return nil, dataflow.CreateSource(output)
}
