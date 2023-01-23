package victronDevice

import (
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/vedirect"
	"log"
	"strings"
	"time"
)

func startVedirect(c *DeviceStruct, output chan dataflow.Value) error {
	log.Printf("device[%s]: start vedirect source", c.deviceConfig.Name())

	// open vedirect device
	vd, err := vedirect.Open(c.victronConfig.Device(), c.deviceConfig.LogComDebug())
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

	log.Printf("device[%s]: source: connect to %s", c.deviceConfig.Name(), deviceString)
	c.model = deviceString

	// get relevant registers
	{
		registers := RegisterFactoryByProduct(deviceId)
		if registers == nil {
			return fmt.Errorf("no registers found for deviceId=%x", deviceId)
		}
		// filter registers by skip list
		c.registers = FilterRegisters(registers, c.deviceConfig.SkipFields(), c.deviceConfig.SkipCategories())
	}

	// start victron reader
	go func() {
		defer close(c.closed)
		defer close(output)

		fetchStaticCounter := 0
		ticker := time.NewTicker(100 * time.Millisecond)
		for {
			select {
			case <-c.shutdown:
				return
			case <-ticker.C:
				start := time.Now()

				// flush async data
				vd.RecvFlush()

				if err := vd.VeCommandPing(); err != nil {
					log.Printf("device[%s]: source: VeCommandPing failed: %v", c.deviceConfig.Name(), err)
					continue
				}

				for _, register := range c.registers {
					// only fetch static registers seldomly
					if register.Static() && (fetchStaticCounter%60 != 0) {
						continue
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
							log.Printf("device[%s]: fetching number register failed: %v", c.deviceConfig.Name(), err)
						} else {
							output <- dataflow.NewNumericRegisterValue(
								c.deviceConfig.Name(),
								register,
								value/float64(register.Factor())+register.Offset(),
							)
						}
					case dataflow.TextRegister:
						value, err := vd.VeCommandGetString(register.Address())

						if err != nil {
							log.Printf("device[%s]: fetching text register failed: %v", c.deviceConfig.Name(), err)
						} else {
							output <- dataflow.NewTextRegisterValue(
								c.deviceConfig.Name(),
								register,
								strings.TrimSpace(value),
							)
						}
					case dataflow.EnumRegister:
						var intValue uint64
						intValue, err = vd.VeCommandGetUint(register.Address())

						if err != nil {
							log.Printf("device[%s]: fetching enum register failed: %v", c.deviceConfig.Name(), err)
						} else {
							output <- dataflow.NewEnumRegisterValue(c.deviceConfig.Name(), register, int(intValue))
						}
					}
				}

				c.SetLastUpdatedNow()

				fetchStaticCounter++

				if c.deviceConfig.LogDebug() {
					log.Printf(
						"device[%s]: registers fetched, took=%.3fs",
						c.deviceConfig.Name(),
						time.Since(start).Seconds(),
					)
				}
			}
		}
	}()

	return nil
}
