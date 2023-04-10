package modbusDevice

import (
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/modbus"
	"log"
	"time"
)

func startWaveshareRtuRelay8(c *DeviceStruct) error {
	log.Printf("device[%s]: start waveshare RTU Relay 8 source", c.deviceConfig.Name())

	// open modbus device
	var err error
	c.modbus, err = modbus.Open(c.modbusConfig.Device(), c.deviceConfig.LogComDebug())
	if err != nil {
		return err
	}

	// get software version
	if version, err := c.modbus.ReadSoftwareRevision(c.modbusConfig.Address()); err != nil {
		return fmt.Errorf("device[%s]: ReadSoftwareRevision failed: %s", c.deviceConfig.Name(), err)
	} else {
		log.Printf("device[%s]: source: version=%s", c.deviceConfig.Name(), version)
	}

	// assign registers
	c.registers = RegisterListRtuRelay8

	c.mainRoutine()

	return nil
}

func (c *DeviceStruct) mainRoutine() {
	if c.deviceConfig.LogDebug() {
		log.Printf("waveshareDevice[%s]: start polling", c.deviceConfig.Name())
	}

	go func() {
		// setup polling
		execPoll := func() {
			start := time.Now()

			// fetch registers
			state, err := c.modbus.ReadRelays(c.modbusConfig.Address())
			if err != nil {
				if c.deviceConfig.LogDebug() {
					log.Printf("device[%s]: read failed: %s", c.deviceConfig.Name(), err)
				}
				device.SendDisconnected(c.Config().Name(), c.stateStorage)
				return
			}
			device.SendConnteced(c.Config().Name(), c.stateStorage)

			for i, register := range c.registers {
				value := 0
				if state[i] {
					value = 1
				}

				c.stateStorage.Fill(dataflow.NewEnumRegisterValue(
					c.deviceConfig.Name(),
					register,
					value,
				))
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
		execPoll()

		defer close(c.closed)

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-c.shutdown:
				return
			case <-ticker.C:
				execPoll()
			}
		}
	}()

}

var RegisterListRtuRelay8 ModbusRegisters

func init() {
	enum := map[int]string{
		0: "OPEN",
		1: "CLOSED",
	}
	RegisterListRtuRelay8 = make(ModbusRegisters, 8)
	for i := uint16(0); i < 8; i += 1 {
		RegisterListRtuRelay8[i] = CreateEnumRegisterStruct(
			"Relays",
			fmt.Sprintf("CH%d", i+1),
			fmt.Sprintf("Relay CH%d", i+1),
			i,
			enum,
			int(i),
		)
	}
}
