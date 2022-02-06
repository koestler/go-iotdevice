package device

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"log"
	"math/rand"
	"time"
)

func CreateRandomDeviceFactory(registers dataflow.Registers) Creator {
	return func(deviceStruct DeviceStruct, output chan dataflow.Value) (device Device, err error) {
		// store given registers
		deviceStruct.registers = registers

		cfg := deviceStruct.Config()

		if cfg.LogDebug() {
			log.Printf("device[%s]: start random source", cfg.Name())
		}

		// start source go routine
		go func() {
			defer close(deviceStruct.GetClosedChan())
			defer close(output)

			ticker := time.NewTicker(time.Second)

			for {
				select {
				case <-deviceStruct.GetShutdownChan():
					return
				case <-ticker.C:
					for _, register := range deviceStruct.Registers() {
						if numberRegister, ok := register.(dataflow.NumberRegisterStruct); ok {
							var value float64
							if numberRegister.Signed() {
								value = 1e2 * (rand.Float64() - 0.5) * 2 * numberRegister.Factor()
							} else {
								value = 1e2 * rand.Float64() * numberRegister.Factor()
							}

							output <- dataflow.NewNumericRegisterValue(
								deviceStruct.Config().Name(),
								register,
								value,
							)
						}
					}
				}
			}
		}()

		return &deviceStruct, nil
	}

}
