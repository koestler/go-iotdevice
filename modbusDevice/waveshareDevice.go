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
		return fmt.Errorf("waveshareDevice[%s]: ReadSoftwareRevision failed: %s", c.deviceConfig.Name(), err)
	} else {
		log.Printf("waveshareDevice[%s]: source: version=%s", c.deviceConfig.Name(), version)
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
					log.Printf("waveshareDevice[%s]: read failed: %s", c.deviceConfig.Name(), err)
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
					"waveshareDevice[%s]: registers fetched, took=%.3fs",
					c.deviceConfig.Name(),
					time.Since(start).Seconds(),
				)
			}
		}
		execPoll()

		// setup subscription to listen for updates of controllable registers
		filter := dataflow.Filter{
			SkipNull:       true,
			IncludeDevices: map[string]bool{c.Config().Name(): true},
		}
		log.Printf(
			"waveshareDevice[%s]: subscribe: %v",
			c.Config().Name(), filter,
		)
		commandSubscription := c.commandStorage.Subscribe(filter)
		defer commandSubscription.Shutdown()

		execCommand := func(value dataflow.Value) {
			if c.Config().LogDebug() {
				log.Printf(
					"waveshareDevice[%s]: controllable command: %s",
					c.Config().Name(), value.String(),
				)
			}

			enumValue, ok := value.(dataflow.EnumRegisterValue)
			if !ok {
				// unable to handle non enum value
				return
			}

			var command modbus.Command
			switch enumValue.Value() {
			case "OPEN":
				command = modbus.RelayOpen
			case "CLOSED":
				command = modbus.RelayClose
			default:
				return
			}

			if c.Config().LogDebug() {
				log.Printf(
					"waveshareDevice[%s]: controllable command: %v",
					c.Config().Name(), command,
				)
			}

			var relayNr uint16
			if modbusRegister, ok := value.Register().(ModbusRegisterStruct); !ok {
				// unknown register
				return
			} else {
				relayNr = modbusRegister.Address()
			}

			if err := c.modbus.WriteRelay(c.modbusConfig.Address(), relayNr, command); err != nil {
				log.Printf(
					"waveshareDevice[%s]: control request genration failed: %s",
					c.Config().Name(), err,
				)
			} else {
				// set the current state immediately after a successful write
				c.stateStorage.Fill(dataflow.NewEnumRegisterValue(
					c.deviceConfig.Name(),
					value.Register(),
					enumValue.EnumIdx(),
				))

				if c.Config().LogDebug() {
					log.Printf("waveshareDevice[%s]: control request successful", c.Config().Name())
				}
			}

			// reset the command; this allows the same command (eg. toggle) to be sent again
			c.commandStorage.Fill(dataflow.NewNullRegisterValue(c.Config().Name(), value.Register()))
		}

		ticker := time.NewTicker(c.modbusConfig.PollInterval())
		defer ticker.Stop()
		defer close(c.closed)
		for {
			select {
			case <-c.shutdown:
				return
			case <-ticker.C:
				execPoll()
			case value := <-commandSubscription.GetOutput():
				execCommand(value)
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
