package victronDevice

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"log"
	"math/rand"
	"time"
)

func startRandom(c *DeviceStruct, output chan dataflow.Value, registers dataflow.Registers) error {
	// filter registers by skip list
	c.registers = dataflow.FilterRegisters(registers, c.deviceConfig.SkipFields(), c.deviceConfig.SkipCategories())

	if c.deviceConfig.LogDebug() {
		log.Printf("device[%s]: start random source", c.deviceConfig.Name())
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
					if numberRegister, ok := register.(dataflow.NumberRegisterStruct); ok {
						var value float64
						if numberRegister.Signed() {
							value = 1e2 * (rand.Float64() - 0.5) * 2 / float64(numberRegister.Factor())
						} else {
							value = 1e2 * rand.Float64() / float64(numberRegister.Factor())
						}

						output <- dataflow.NewNumericRegisterValue(c.deviceConfig.Name(), register, value)
					} else if _, ok := register.(dataflow.TextRegisterStruct); ok {
						output <- dataflow.NewTextRegisterValue(c.deviceConfig.Name(), register, randomString(8))
					}
				}
				c.SetLastUpdatedNow()
			}
		}
	}()

	return nil
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
