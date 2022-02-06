package victron

import (
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/vedirect"
	"log"
	"time"
)

type VictronDeviceStruct struct {
	device.DeviceStruct
	deviceId vedirect.VeProduct
}

func CreateVictronDevice(deviceStruct device.DeviceStruct, output chan dataflow.Value) (device device.Device, err error) {
	device = VictronDeviceStruct{
		DeviceStruct: deviceStruct,
	}
	cfg := device.Config()

	log.Printf("device[%s]: start vedirect source", cfg.Name())

	// open vedirect device
	vd, err := vedirect.Open(cfg.Device(), cfg.LogComDebug())
	if err != nil {
		return nil, err
	}

	// send ping
	if err := vd.VeCommandPing(); err != nil {
		return nil, fmt.Errorf("ping failed: %s", err)
	}

	// get deviceId
	deviceId, err := vd.VeCommandDeviceId()
	if err != nil {
		return nil, fmt.Errorf("cannot get DeviceId: %s", err)
	}

	deviceString := deviceId.String()
	if len(deviceString) < 1 {
		return nil, fmt.Errorf("unknown deviceId=%x", err)
	}

	log.Printf("device[%s]: source: connect to %s", cfg.Name(), deviceString)

	// get relevant registers
	registers := RegisterFactoryByProduct(deviceId)
	if registers == nil {
		return nil, fmt.Errorf("no registers found for deviceId=%x", deviceId)
	}

	// start victron reader
	go func() {
		defer close(deviceStruct.GetClosedChan())
		defer close(output)

		// flush buffer
		vd.RecvFlush()

		ticker := time.NewTicker(100 * time.Millisecond)

		for {
			select {
			case <-deviceStruct.GetShutdownChan():
				return
			case <-ticker.C:
				start := time.Now()

				if err := vd.VeCommandPing(); err != nil {
					log.Printf("device[%s]: source: VeCommandPing failed: %v", cfg.Name(), err)
					continue
				}

				for _, register := range registers {
					if numberRegister, ok := register.(dataflow.NumberRegisterStruct); ok {
						var value float64
						if numberRegister.Signed() {
							var intValue int64
							intValue, err = vd.VeCommandGetInt(register.Address())
							value = float64(intValue)
						} else {
							var intValue uint64
							intValue, err = vd.VeCommandGetUint(register.Address())
							value = float64(intValue)
						}

						if err != nil {
							log.Printf("device[%s]: victron.RecvNumeric failed: %v", cfg.Name(), err)
						} else {
							output <- dataflow.NewNumericRegisterValue(
								deviceStruct.Config().Name(),
								register,
								value,
							)
						}
					}
				}

				if cfg.LogDebug() {
					log.Printf(
						"device[%s]: registers fetched, took=%.3fs",
						cfg.Name(),
						time.Since(start).Seconds(),
					)
				}
			}
		}
	}()

	return
}
