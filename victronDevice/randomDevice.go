package victronDevice

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"log"
	"math/rand"
	"time"
)

func startRandom(c *DeviceStruct, output dataflow.Fillable, registers VictronRegisters) error {
	// filter registers by skip list
	c.registers = FilterRegisters(registers, c.deviceConfig.SkipFields(), c.deviceConfig.SkipCategories())

	if c.deviceConfig.LogDebug() {
		log.Printf("device[%s]: start random source", c.deviceConfig.Name())
	}

	// start source go routine
	go func() {
		defer close(c.closed)

		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-c.shutdown:
				return
			case <-ticker.C:
				for _, r := range registers {
					switch r.RegisterType() {
					case dataflow.NumberRegister:
						var value float64
						if r.Signed() {
							value = 1e2*(rand.Float64()-0.5)*2/float64(r.Factor()) + r.Offset()
						} else {
							value = 1e2*rand.Float64()/float64(r.Factor()) + r.Offset()
						}
						output.Fill(dataflow.NewNumericRegisterValue(c.deviceConfig.Name(), r, value))
					case dataflow.TextRegister:
						output.Fill(dataflow.NewTextRegisterValue(c.deviceConfig.Name(), r, randomString(8)))
					case dataflow.EnumRegister:
						output.Fill(dataflow.NewEnumRegisterValue(c.deviceConfig.Name(), r, randomEnum(r.Enum())))
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

func randomEnum(enum map[int]string) int {
	k := rand.Intn(len(enum))
	for idx := range enum {
		if k == 0 {
			return idx
		}
		k--
	}
	return 0
}
