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
					for _, register := range deviceStruct.GetRegisters() {
						if numberRegister, ok := register.(dataflow.NumberRegisterStruct); ok {
							var value float64
							if numberRegister.Signed() {
								value = 1e2 * (rand.Float64() - 0.5) * 2 / float64(numberRegister.Factor())
							} else {
								value = 1e2 * rand.Float64() / float64(numberRegister.Factor())
							}

							output <- dataflow.NewNumericRegisterValue(
								deviceStruct.Config().Name(),
								register,
								value,
							)
						} else if _, ok := register.(dataflow.TextRegisterStruct); ok {
							output <- dataflow.NewTextRegisterValue(
								deviceStruct.Config().Name(),
								register,
								randomString(8),
							)
						}
					}
					device.SetLastUpdatedNow()
				}
			}
		}()

		return &deviceStruct, nil
	}

}

func randomString(n int) string {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num := rand.Intn(len(letters))
		ret[i] = letters[num]
	}

	return string(ret)
}
