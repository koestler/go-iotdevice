package victron

import (
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/vedirect"
	"log"
	"time"
)

func RunVictron(c *device.Device, output chan dataflow.Value) (err error) {
	log.Printf("device[%s]: start vedirect source", c.Config().Name())

	// open vedirect device
	vd, err := vedirect.Open(c.cfg.Device(), c.cfg.LogComDebug())
	if err != nil {
		return err
	}

	// send ping
	if err := vd.VeCommandPing(); err != nil {
		return fmt.Errorf("ping failed: %s", err)
	}

	// get deviceId
	deviceId, err := vd.VeCommandDeviceId()
	if err != nil {
		return fmt.Errorf("cannot get DeviceId: %s", err)
	}

	deviceString := deviceId.String()
	if len(deviceString) < 1 {
		return fmt.Errorf("unknown deviceId=%x", err)
	}

	log.Printf("device[%s]: source: connect to %s", c.cfg.Name(), deviceString)

	// get relevant registers
	registers := victron.RegisterFactoryByProduct(deviceId)
	if registers == nil {
		return fmt.Errorf("no registers found for deviceId=%x", deviceId)
	}

	// start victron reader
	go func() {
		defer close(c.closed)
		defer close(output)

		// flush buffer
		vd.RecvFlush()

		ticker := time.NewTicker(100 * time.Millisecond)

		for {
			select {
			case <-c.shutdown:
				return
			case <-ticker.C:
				start := time.Now()

				if err := vd.VeCommandPing(); err != nil {
					log.Printf("device[%s]: source: VeCommandPing failed: %v", c.cfg.Name(), err)
					continue
				}

				for _, register := range registers {
					if numericValue, err := getRegister(vd, register); err != nil {
						log.Printf("device[%s]: victron.RecvNumeric failed: %v", c.cfg.Name(), err)
					} else {
						output <- dataflow.Value{
							DeviceName:    c.cfg.Name(),
							Name:          register.Name,
							Value:         numericValue.Value,
							Unit:          numericValue.Unit,
							RoundDecimals: 3,
						}
					}
				}

				if c.cfg.LogDebug() {
					log.Printf(
						"device[%s]: registers fetched, took=%.3fs",
						c.cfg.Name(),
						time.Since(start).Seconds(),
					)
				}
			}
		}
	}()

	return
}

func getRegister(vd *vedirect.Vedirect, reg dataflow.Register) (result NumericValue, err error) {
	var value float64

	switch reg.Type {
	case dataflow.SignedNumberRegister:
		var intValue int64
		intValue, err = vd.VeCommandGetInt(reg.Address)
		value = float64(intValue)
	case dataflow.UnsignedNumberRegister:
		var intValue uint64
		intValue, err = vd.VeCommandGetUint(reg.Address)
		value = float64(intValue)
	}

	if err != nil {
		log.Printf("victron.RecvNumeric failed: %v", err)
		return
	}

	result = NumericValue{
		Value: value * reg.Factor,
		Unit:  reg.Unit,
	}

	return
}
