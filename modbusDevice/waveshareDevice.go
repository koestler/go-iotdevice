package modbusDevice

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"log"
	"time"
)

func runWaveshareRtuRelay8(ctx context.Context, c *DeviceStruct) (err error, immediateError bool) {
	log.Printf("device[%s]: start waveshare RTU Relay 8 source", c.Name())

	// get software version
	if version, err := ReadSoftwareRevision(c.modbus.WriteRead, c.modbusConfig.Address()); err != nil {
		return fmt.Errorf("waveshareDevice[%s]: ReadSoftwareRevision failed: %s", c.Name(), err), true
	} else {
		log.Printf("waveshareDevice[%s]: source: version=%s", c.Name(), version)
	}

	// assign registers
	registers := c.getModbusRegisters()
	registers = dataflow.FilterRegisters(registers, c.Config().RegisterFilter())

	addToRegisterDb(c.RegisterDb(), registers)

	// setup polling
	execPoll := func() error {
		start := time.Now()

		// fetch registers
		state, err := ReadRelays(c.modbus.WriteRead, c.modbusConfig.Address())
		if err != nil {
			return fmt.Errorf("waveshareDevice[%s]: read failed: %s", c.Name(), err)
		}

		for _, register := range registers {
			value := 0
			if state[register.address] {
				value = 1
			}

			c.StateStorage().Fill(dataflow.NewEnumRegisterValue(
				c.Name(),
				register,
				value,
			))
		}

		if c.Config().LogDebug() {
			log.Printf(
				"waveshareDevice[%s]: registers fetched, took=%.3fs",
				c.Name(),
				time.Since(start).Seconds(),
			)
		}

		return nil
	}

	if err := execPoll(); err != nil {
		return err, true
	}

	// send connected now, disconnected when this routine stops
	c.SetAvailable(true)
	defer func() {
		c.SetAvailable(false)
	}()

	// setup subscription to listen for updates of controllable registers
	_, commandSubscription := c.commandStorage.SubscribeReturnInitial(ctx, dataflow.DeviceNonNullValueFilter(c.Config().Name()))

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
		if modbusRegister, ok := value.Register().(ModbusRegister); !ok {
			// unknown register
			return
		} else {
			relayNr = modbusRegister.address
		}

		if err := WriteRelay(c.modbus.WriteRead, c.modbusConfig.Address(), relayNr, command); err != nil {
			log.Printf(
				"waveshareDevice[%s]: control request genration failed: %s",
				c.Config().Name(), err,
			)
		} else {
			// set the current state immediately after a successful write
			c.StateStorage().Fill(dataflow.NewEnumRegisterValue(
				c.Name(),
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
		case value := <-commandSubscription.Drain():
			execCommand(value)
		}
	}
}

func (c *DeviceStruct) getModbusRegisters() (registers []ModbusRegister) {
	category := "Relays"
	registers = make([]ModbusRegister, 0, 8)
	for i := uint16(0); i < 8; i += 1 {
		name := fmt.Sprintf("CH%d", i+1)

		description := c.modbusConfig.RelayDescription(name)
		enum := map[int]string{
			0: c.modbusConfig.RelayOpenLabel(name),
			1: c.modbusConfig.RelayClosedLabel(name),
		}

		r := NewModbusRegister(
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
