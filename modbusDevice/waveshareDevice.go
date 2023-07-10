package modbusDevice

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"log"
	"time"
)

func runWaveshareRtuRelay8(ctx context.Context, c *DeviceStruct) (err error, immediateError bool) {
	log.Printf("device[%s]: start waveshare RTU Relay 8 source", c.deviceConfig.Name())

	// get software version
	if version, err := ReadSoftwareRevision(c.modbus.WriteRead, c.modbusConfig.Address()); err != nil {
		return fmt.Errorf("waveshareDevice[%s]: ReadSoftwareRevision failed: %s", c.deviceConfig.Name(), err), true
	} else {
		log.Printf("waveshareDevice[%s]: source: version=%s", c.deviceConfig.Name(), version)
	}

	// assign registers
	c.registers = c.getModbusRegisters()

	// setup polling
	execPoll := func() error {
		start := time.Now()

		// fetch registers
		state, err := ReadRelays(c.modbus.WriteRead, c.modbusConfig.Address())
		if err != nil {
			return fmt.Errorf("waveshareDevice[%s]: read failed: %s", c.deviceConfig.Name(), err)
		}

		for _, register := range c.registers {
			value := 0
			if state[register.Address()] {
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

		return nil
	}

	if err := execPoll(); err != nil {
		return err, true
	}

	// send connected now, disconnected when this routine stops
	device.SendConnteced(c.Config().Name(), c.stateStorage)
	defer func() {
		device.SendDisconnected(c.Config().Name(), c.stateStorage)
	}()

	// setup subscription to listen for updates of controllable registers
	filter := dataflow.Filter{
		SkipNull:       true,
		IncludeDevices: map[string]bool{c.Config().Name(): true},
	}
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

		var command Command
		switch enumValue.EnumIdx() {
		case 0:
			command = RelayOpen
		case 1:
			command = RelayClose
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

		if err := WriteRelay(c.modbus.WriteRead, c.modbusConfig.Address(), relayNr, command); err != nil {
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
	for {
		select {
		case <-ctx.Done():
			return nil, false
		case <-ticker.C:
			if err := execPoll(); err != nil {
				return err, false
			}
		case value := <-commandSubscription.GetOutput():
			execCommand(value)
		}
	}
}

func (c *DeviceStruct) getModbusRegisters() (registers ModbusRegisters) {
	category := "Relays"
	registers = make(ModbusRegisters, 0, 8)
	for i := uint16(0); i < 8; i += 1 {
		name := fmt.Sprintf("CH%d", i+1)

		if device.IsExcluded(name, category, c.deviceConfig) {
			continue
		}

		description := c.modbusConfig.RelayDescription(name)
		enum := map[int]string{
			0: c.modbusConfig.RelayOpenLabel(name),
			1: c.modbusConfig.RelayClosedLabel(name),
		}

		r := CreateEnumRegisterStruct(
			category,
			name,
			description,
			i,
			enum,
			int(i),
		)
		registers = append(registers, r)
	}

	return
}
