package modbusDevice

import (
	"fmt"
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

		ticker := time.NewTicker(2000 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-c.shutdown:
				if err := md.Close(); err != nil {
					log.Printf("device[%s]: vd.Close failed: %s", c.deviceConfig.Name(), err)
				}
				return
			case <-ticker.C:
				start := time.Now()

				// fetch registers
				if state, err := md.ReadRelays(0x01); err != nil {
					log.Printf("read failed: %s", err)
				} else {
					log.Printf("state is: %v", state)
				}

				/*
									for i := range c.registers {
						md.WriteRelay(0x01, i, modbus.RelayFlip)
					}
				*/

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
