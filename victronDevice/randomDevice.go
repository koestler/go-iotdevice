package victronDevice

import (
	"context"
	"github.com/koestler/go-iotdevice/dataflow"
	"log"
	"math/rand"
	"time"
)

func runRandom(ctx context.Context, c *DeviceStruct, output dataflow.Fillable, registers VictronRegisters) (err error, immediateError bool) {
	// send connected now, disconnected when this routine stops
	c.SetAvailable(true)
	defer func() {
		c.SetAvailable(false)
	}()

	// filter registers by skip list
	c.registers = FilterRegisters(registers, c.Config().SkipFields(), c.Config().SkipCategories())

	if c.Config().LogDebug() {
		log.Printf("device[%s]: start random source", c.Name())
	}

	// start source loop
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, false
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
					output.Fill(dataflow.NewNumericRegisterValue(c.Name(), r, value))
				case dataflow.TextRegister:
					output.Fill(dataflow.NewTextRegisterValue(c.Name(), r, randomString(8)))
				case dataflow.EnumRegister:
					output.Fill(dataflow.NewEnumRegisterValue(c.Name(), r, randomEnum(r.Enum())))
				}
			}
			c.SetLastUpdatedNow()
		}
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
