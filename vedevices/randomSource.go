package vedevices

import (
	"github.com/koestler/go-victron-to-mqtt/dataflow"
	"math/rand"
	"time"
)

func CreateRandomSource(deviceName string, registers Registers) *dataflow.Source {
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
	return dataflow.CreateSource(output)
}