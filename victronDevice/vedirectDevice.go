package victronDevice

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/vedirect"
	"log"
	"strings"
	"time"
)

func runVedirect(ctx context.Context, c *DeviceStruct, output dataflow.Fillable) (err error, immediateError bool) {
	log.Printf("device[%s]: start vedirect source", c.Name())

	// open vedirect device
	vd, err := vedirect.Open(c.victronConfig.Device(), c.Config().LogComDebug())
	if err != nil {
		return err, true
	}
	defer func() {
		if err := vd.Close(); err != nil {
			log.Printf("device[%s]: vd.Close failed: %s", c.Name(), err)
		}
	}()

	// send ping
	if err := vd.VeCommandPing(); err != nil {
		return fmt.Errorf("ping failed: %s", err), true
	}

	// send connected now, disconnected when this routine stops
	c.SetAvailable(true)
	defer func() {
		c.SetAvailable(false)
	}()

	// get deviceId
	deviceId, err := vd.VeCommandDeviceId()
	if err != nil {
		return fmt.Errorf("cannot get DeviceId: %s", err), true
	}

	deviceString := deviceId.String()
	if len(deviceString) < 1 {
		return fmt.Errorf("unknown deviceId=%x", err), true
	}

	log.Printf("device[%s]: source: connect to %s", c.Name(), deviceString)
	c.model = deviceString

	// get relevant registers
	{
		registers := RegisterFactoryByProduct(deviceId)
		if registers == nil {
			return fmt.Errorf("no registers found for deviceId=%x", deviceId), true
		}
		// filter registers by skip list
		c.registers = FilterRegisters(registers, c.Config().SkipFields(), c.Config().SkipCategories())
	}

	// start polling loop
	fetchStaticCounter := 0
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, false
		case <-ticker.C:
			start := time.Now()

			// flush async data
			vd.RecvFlush()

			// execute a Ping at the beginning and after each error
			pingNeeded := true

			var enumCacheAddr uint16
			var enumCacheValue uint64
			for _, register := range c.registers {
				// only fetch static registers seldomly
				if register.Static() && (fetchStaticCounter%60 != 0) {
					continue
				}

				if pingNeeded {
					if err := vd.VeCommandPing(); err != nil {
						return fmt.Errorf("device[%s]: source: VeCommandPing failed: %s", c.Name(), err), false
					}
				}

				switch register.RegisterType() {
				case dataflow.NumberRegister:
					var value float64
					if register.Signed() {
						var intValue int64
						intValue, err = vd.VeCommandGetInt(register.Address())
						value = float64(intValue)
					} else {
						var intValue uint64
						intValue, err = vd.VeCommandGetUint(register.Address())
						value = float64(intValue)
					}

					if err != nil {
						log.Printf("device[%s]: fetching number register failed: %v", c.Name(), err)
					} else {
						output.Fill(dataflow.NewNumericRegisterValue(
							c.Name(),
							register,
							value/float64(register.Factor())+register.Offset(),
						))
					}
				case dataflow.TextRegister:
					if value, err := vd.VeCommandGetString(register.Address()); err != nil {
						log.Printf("device[%s]: fetching text register failed: %v", c.Name(), err)
					} else {
						output.Fill(dataflow.NewTextRegisterValue(
							c.Name(),
							register,
							strings.TrimSpace(value),
						))
					}
				case dataflow.EnumRegister:
					var intValue uint64

					if addr := register.Address(); enumCacheAddr != 0 && enumCacheAddr == addr {
						intValue = enumCacheValue
						err = nil
					} else {
						intValue, err = vd.VeCommandGetUint(addr)
						if err == nil {
							enumCacheAddr = addr
							enumCacheValue = intValue
						}
					}

					if err != nil {
						log.Printf("device[%s]: fetching enum register failed: %v", c.Name(), err)
					} else {
						if bit := register.Bit(); bit >= 0 {
							intValue = (intValue >> bit) & 1
						}

						output.Fill(dataflow.NewEnumRegisterValue(
							c.Name(),
							register,
							int(intValue),
						))
					}
				}

				pingNeeded = err != nil
			}

			c.SetLastUpdatedNow()

			fetchStaticCounter++

			if c.Config().LogDebug() {
				log.Printf(
					"device[%s]: registers fetched, took=%.3fs",
					c.Name(),
					time.Since(start).Seconds(),
				)
			}
		}
	}
}
