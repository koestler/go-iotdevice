package victronDevice

import (
	"context"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-victron/veregister"
	"log"
	"math/rand"
	"time"
)

func runRandom(ctx context.Context, c *DeviceStruct, output dataflow.Fillable, rl veregister.RegisterList) (err error, immediateError bool) {
	// send connected now, disconnected when this routine stops
	c.SetAvailable(true)
	defer func() {
		c.SetAvailable(false)
	}()

	// filter registers by skip list
	rl.FilterRegister(func() func(r veregister.Register) bool {
		rf := dataflow.RegisterFilter(c.Config().Filter())
		return func(r veregister.Register) bool {
			return rf(r)
		}
	}())
	addToRegisterDb(c.RegisterDb(), rl)

	if c.Config().LogDebug() {
		log.Printf("victronDevice[%s]: start random source", c.Name())
	}

	// start source loop
	ticker := time.NewTicker(c.victronConfig.PollInterval())
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, false
		case <-ticker.C:
			for _, r := range rl.NumberRegisters {
				var value float64
				if r.Signed() {
					value = 1e2*(rand.Float64()-0.5)*2/float64(r.Factor()) + r.Offset()
				} else {
					value = 1e2*rand.Float64()/float64(r.Factor()) + r.Offset()
				}
				output.Fill(dataflow.NewNumericRegisterValue(c.Name(), Register{r}, value))
			}
			for _, r := range rl.TextRegisters {
				output.Fill(dataflow.NewTextRegisterValue(c.Name(), Register{r}, randomString(8)))
			}
			for _, r := range rl.EnumRegisters {
				output.Fill(dataflow.NewEnumRegisterValue(c.Name(), Register{r}, randomEnum(r.Factory().IntToStringMap())))
			}
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
