package modbusDevice

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"log"
	"time"
)

func runFinder7M38(ctx context.Context, c *DeviceStruct) (err error, immediateError bool) {
	log.Printf("device[%s]: start Finder 7M.38 source", c.Name())

	// get software version
	if version, err := ReadSoftwareRevision(c.modbus.WriteRead, c.modbusConfig.Address()); err != nil {
		return fmt.Errorf("finder7N38Device[%s]: ReadSoftwareRevision failed: %s", c.Name(), err), true
	} else {
		log.Printf("finder7N38Device[%s]: source: version=%s", c.Name(), version)
	}

	// assign registers
	registers := c.getWaveshareRtuRelay8Registers()
	registers = dataflow.FilterRegisters(registers, c.Config().Filter())
	c.RegisterDb().AddStruct(registers...)

	// setup polling
	execPoll := func() error {
		start := time.Now()

		// fetch registers
		state, err := ReadRelays(c.modbus.WriteRead, c.modbusConfig.Address())
		if err != nil {
			return fmt.Errorf("finder7N38Device[%s]: read failed: %s", c.Name(), err)
		}

		for _, register := range registers {
			value := 0
			if address, err := waveshareRtuRelay8RegisterAddress(register); err == nil {
				if state[address] {
					value = 1
				}
			}

			c.StateStorage().Fill(dataflow.NewEnumRegisterValue(
				c.Name(),
				register,
				value,
			))
		}

		if c.Config().LogDebug() {
			log.Printf(
				"finder7N38Device[%s]: registers fetched, took=%.3fs",
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
		}
	}
}

var RegisterList7M38 = []FinderRegister{
	NewFinderRegister(
		"Product",
		"ModelNumber",
		"Model Number",
		FinderT_Str16,
		30001, 30008,
		nil,
		"",
		100,
	),
	NewFinderRegister(
		"Product",
		"SerialNumber",
		"Serial Number",
		FinderT_Str8,
		30009, 30012,
		nil,
		"",
		101,
	),
	NewFinderRegister(
		"Product",
		"SoftwareReference",
		"Software Reference",
		FinderT1,
		30013, 30013,
		nil,
		"",
		102,
	),
	NewFinderRegister(
		"Product",
		"HardwareReference",
		"Hardware Reference",
		FinderT_Str2,
		30014, 30014,
		nil,
		"",
		102,
	),
	NewFinderRegister(
		"Actual Measurements",
		"PhaseValid",
		"Phase valid measurement",
		FinderT1,
		30101, 30101,
		map[int]string{
			0x00: "phase 1 valid  , phase 2 valid  , phase 3 valid",
			0x01: "phase 1 invalid, phase 2 valid  , phase 3 valid",
			0x02: "phase 1 valid  , phase 2 invalid, phase 3 valid",
			0x03: "phase 1 invalid, phase 2 invalid, phase 3 valid",
			0x04: "phase 1 valid  , phase 2 valid  , phase 3 invalid",
			0x05: "phase 1 invalid, phase 2 valid  , phase 3 invalid",
			0x06: "phase 1 valid  , phase 2 invalid, phase 3 invalid",
			0x07: "phase 1 invalid, phase 2 invalid, phase 3 invalid",
		},
		"",
		102,
	),
	NewFinderRegister(
		"Actual Measurements",
		"RunTime",
		"Run time",
		FinderT3,
		30103, 30104,
		nil,
		"s",
		102,
	),
	NewFinderRegister(
		"Actual Measurements",
		"Frequency",
		"Frequency",
		FinderT5,
		30105, 30106,
		nil,
		"Hz",
		102,
	),
	NewFinderRegister(
		"Actual Measurements",
		"U1",
		"U1",
		FinderT5,
		30107, 30108,
		nil,
		"",
		102,
	),
	NewFinderRegister(
		"Actual Measurements",
		"U2",
		"U2",
		FinderT5,
		30109, 30110,
		nil,
		"",
		102,
	),
	NewFinderRegister(
		"Actual Measurements",
		"U3",
		"U3",
		FinderT5,
		30111, 30112,
		nil,
		"",
		102,
	),
	NewFinderRegister(
		"Actual Measurements",
		"UavgPN",
		"Uavg (phase to neutral)",
		FinderT5,
		30113, 30114,
		nil,
		"",
		102,
	),
	NewFinderRegister(
		"Actual Measurements",
		"j12",
		"j12",
		FinderT17,
		30115, 30115,
		nil,
		"",
		102,
	),
	NewFinderRegister(
		"Actual Measurements",
		"j23",
		"j23",
		FinderT17,
		30116, 30116,
		nil,
		"",
		102,
	),
	NewFinderRegister(
		"Actual Measurements",
		"j31",
		"j31",
		FinderT17,
		30117, 30117,
		nil,
		"",
		102,
	),
	NewFinderRegister(
		"Actual Measurements",
		"U12",
		"U12",
		FinderT5,
		30118, 30119,
		nil,
		"",
		102,
	),
	NewFinderRegister(
		"Actual Measurements",
		"U23",
		"U23",
		FinderT5,
		30120, 30121,
		nil,
		"",
		102,
	),
	NewFinderRegister(
		"Actual Measurements",
		"U31",
		"U31",
		FinderT5,
		30122, 30123,
		nil,
		"",
		102,
	),
	NewFinderRegister(
		"Actual Measurements",
		"UavgPP",
		"Uavg (phase to phase)",
		FinderT5,
		30124, 30125,
		nil,
		"",
		102,
	),
	NewFinderRegister(
		"Actual Measurements",
		"I1",
		"I1",
		FinderT5,
		30126, 30127,
		nil,
		"A",
		102,
	),
	NewFinderRegister(
		"Actual Measurements",
		"I2",
		"I2",
		FinderT5,
		30128, 30129,
		nil,
		"A",
		102,
	),
	NewFinderRegister(
		"Actual Measurements",
		"I3",
		"I3",
		FinderT5,
		30130, 30131,
		nil,
		"A",
		102,
	),
	NewFinderRegister(
		"Actual Measurements",
		"INc",
		"INc",
		FinderT5,
		30132, 30133,
		nil,
		"A",
		102,
	),
	NewFinderRegister(
		"Actual Measurements",
		"Iavg",
		"Iavg",
		FinderT5,
		30136, 30137,
		nil,
		"A",
		102,
	),
}
