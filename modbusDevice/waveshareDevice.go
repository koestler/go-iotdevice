package modbusDevice

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/modbus"
	"log"
	"time"
)

func startWaveshareRtuRelay8(c *DeviceStruct, output dataflow.Fillable) error {
	log.Printf("device[%s]: start waveshare RTU Relay 8 source", c.deviceConfig.Name())

	// open vedirect device
	md, err := modbus.Open(c.modbusConfig.Device(), c.deviceConfig.LogComDebug())
	if err != nil {
		return err
	}

	// send ping
	//if err := vd.VeCommandPing(); err != nil {
	//	return fmt.Errorf("ping failed: %s", err)
	//}

	// get deviceId
	//deviceId, err := vd.VeCommandDeviceId()
	//if err != nil {
	//	return fmt.Errorf("cannot get DeviceId: %s", err)
	//}

	// assign registers
	c.registers = RegisterListRtuRelay8

	// start reader
	go func() {
		defer close(c.closed)

		ticker := time.NewTicker(10000 * time.Millisecond)
		for {
			select {
			case <-c.shutdown:
				if err := md.Close(); err != nil {
					log.Printf("device[%s]: vd.Close failed: %s", c.deviceConfig.Name(), err)
				}
				return
			case <-ticker.C:
				start := time.Now()

				for _, register := range c.registers {
					// fetch register
					log.Printf("try to fetch register: %v", register)
				}

				c.SetLastUpdatedNow()

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
