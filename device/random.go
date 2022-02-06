package device

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"log"
	"math/rand"
	"time"
)

func CreateRandom(deviceStruct DeviceStruct, output chan dataflow.Value) (device Device, err error) {
	cfg := deviceStruct.Config()

	if cfg.LogDebug() {
		log.Printf("device[%s]: start random source", cfg.Name())
	}

	// start source go routine
	go func() {
		defer close(deviceStruct.closed)
		defer close(output)

		ticker := time.NewTicker(time.Second)

		for {
			select {
			case <-deviceStruct.shutdown:
				return
			case <-ticker.C:
				for _, register := range deviceStruct.Registers() {
					if signedNumberRegister, ok := register.(dataflow.SignedNumberRegisterStruct); ok {
						output <- dataflow.NewNumericRegisterValue(
							deviceStruct.Config().Name(),
							register,
							1e2*(rand.Float64()-0.5)*2*signedNumberRegister.Factor(),
						)
					} else if unsignedNumberRegister, ok := register.(dataflow.UnsignedNumberRegisterStruct); ok {
						output <- dataflow.NewNumericRegisterValue(
							deviceStruct.Config().Name(),
							register,
							1e2*rand.Float64()*unsignedNumberRegister.Factor(),
						)
					}
				}
			}
		}
	}()

	return &deviceStruct, nil
}

func init() {
	RegisterCreator(config.RandomBmvKind, CreateRandom)
}
