package modbusDevice

// protocol documentation: https://cdn.findernet.com/app/uploads/Benutzerhandbuch_Typ_7M38_DE.pdf

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"log"
	"time"
)

const (
	FinderFunctionReadHoldingResgisters FunctionCode = 0x03
	FinderFunctionReadInputRegisters    FunctionCode = 0x04
)

const InputRegisterAddressOffset = 30000

func runFinder7M38(ctx context.Context, c *DeviceStruct) (err error, immediateError bool) {
	log.Printf("device[%s]: start Finder 7M.38 source", c.Name())

	// assign registers
	registers := RegisterList7M38()
	registers = dataflow.FilterRegisters(registers, c.Config().Filter())
	addToRegisterDb(c.RegisterDb(), registers)

	// setup polling
	execPoll := func() error {
		start := time.Now()

		// fetch registers
		for _, register := range registers {
			v, err := FinderReadRegister(c, register)

			if err != nil {
				return err
			}

			c.StateStorage().Fill(v)
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

var RegisterList7M38FloatRegisters = []struct {
	addr             uint16
	name, desc, unit string
}{
	{32480, "RunTime", "Run time", "s"},
	{32484, "UAvgPN", "Uavg (phase to neutral)", "V"},
	{32486, "UAvgPP", "Uavg (phase to phase)", "V"},
	{32488, "SI", "S I", ""},
	{32490, "Pt", "Active Power Total", "W"},
	{32492, "Qt", "Reactive Power Total", "W"},
	{32494, "St", "Apparent Power Total", "W"},
	{32496, "PFt", "Power Factor Total", ""},
	{32498, "F", "Frequency", "Hz"},
	{32500, "U1", "U1", "V"},
	{32502, "U2", "U2", "V"},
	{32504, "U3", "U3", "V"},
	{32506, "UAvgPN2", "Uavg (phase to neutral)", "V"},
	{32508, "U12", "U12", "V"},
	{32510, "U23", "U23", "V"},
	{32512, "U31", "U31", "V"},
	{32514, "UAvgPP2", "Uavg (phase to phase)", "V"},
	{32516, "I1", "I1", "A"},
	{32518, "I2", "I2", "A"},
	{32520, "I3", "I3", "A"},
	{32522, "SI2", "S I (??)", ""},
	{32524, "INCalc", "I neutral (calculated)", "A"},
	{32526, "InMeas", "I neutral (measured)", "A"},
	{32528, "Iavg", "Iavg", "A"},
	{32530, "P1", "Active Power Phase L1", "W"},
	{32532, "P2", "Active Power Phase L2", "W"},
	{32534, "P3", "Active Power Phase L3", "W"},
	{32536, "Pt", "Active Power Total", "W"},
	{32538, "Q1", "Reactive Power Phase L1", "W"},
	{32540, "Q2", "Reactive Power Phase L2", "W"},
	{32542, "Q3", "Reactive Power Phase L3", "W"},
	{32544, "Qt", "Reactive Power Total", "W"},
	{32546, "S1", "Apparent Power Phase L1 ", "W"},
	{32548, "S2", "Apparent Power Phase L2 ", "W"},
	{32550, "S3", "Apparent Power Phase L3 ", "W"},
	{32552, "St", "Apparent Power Total", "W"},
	{32554, "PF1", "Power Factor Phase 1", ""},
	{32556, "PF2", "Power Factor Phase 2", ""},
	{32558, "PF3", "Power Factor Phase 3", ""},
	{32560, "PFt", "Power Factor Total", ""},
	{32562, "PF1", "CAP/IND P. F. Phase 1", ""},
	{32564, "PF2", "CAP/IND P. F. Phase 2", ""},
	{32566, "PF3", "CAP/IND P. F. Phase 3", ""},
	{32568, "PFt", "CAP/IND P. F. Total", ""},
	{32570, "J1", "j1 (angle between U1 and I1)", ""},
	{32572, "J2", "j2 (angle between U2 and I2)", ""},
	{32574, "J3", "j3 (angle between U3 and I3) ", ""},
	{32576, "Jt", "Power Angle Total (atan2(Pt,Qt))", ""},
	{32578, "J12", "j12 (angle between U1 and U2)", ""},
	{32580, "J23", "j23 (angle between U2 and U3)", ""},
	{32582, "J31", "j31 (angle between U3 and U1)", ""},
	{32584, "F2", "Frequency (?)", ""},
	{32588, "I1Thd", "I1 THD%", ""},
	{32590, "I2Thd", "I2 THD%", ""},
	{32592, "I3Thd", "I3 THD%", ""},
	{32638, "EnergyCounterN1", "Energy Counter n1", ""},
	{32640, "EnergyCounterN2", "Energy Counter n2", ""},
	{32642, "EnergyCounterN3", "Energy Counter n3", ""},
	{32644, "EnergyCounterN4", "Energy Counter n4", ""},
	{32658, "InternalTemp", "Internal Temperature", ""},
	{32985, "Unom", "nominal phase voltage", ""},
	{32987, "Inom", "nominal phase current", ""},
	{32989, "Pnom", "nominal phase power", ""},
	{32991, "Ptot", "nominal total power", ""},
	{32993, "Itot", "nominal total current", ""},
	{32995, "Fnom", "nominal frequency", ""},
	{34999, "RunTime2", "Run time", "s"},
}

var RegisterList7M38 = func() []FinderRegister {
	productResgisters := []FinderRegister{
		NewFinderRegister(
			"Product",
			"ModelNumber",
			"Model Number",
			FinderTStr16,
			30001, 30008,
			nil,
			"",
			100,
		),
		NewFinderRegister(
			"Product",
			"SerialNumber",
			"Serial Number",
			FinderTStr8,
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
			FinderTStr2,
			30014, 30014,
			nil,
			"",
			103,
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
			104,
		),
	}

	ret := make([]FinderRegister, 0, len(productResgisters)+len(RegisterList7M38FloatRegisters))
	ret = append(ret, productResgisters...)

	for idx, fr := range RegisterList7M38FloatRegisters {
		ret = append(ret,
			NewFinderRegister(
				"Measurements",
				fr.name,
				fr.desc,
				FinderTFloat,
				fr.addr, fr.addr+1,
				nil,
				fr.unit,
				200+idx,
			),
		)
	}

	return ret
}

func FinderReadRegister(c *DeviceStruct, register FinderRegister) (v dataflow.Value, err error) {
	switch register.RegisterType() {
	case dataflow.NumberRegister:
		return FinderReadFloatRegister(c, register)
	case dataflow.TextRegister:
		return FinderReadStringRegister(c, register)
	default:
		return nil, fmt.Errorf("FinderReadRegister does not implement registerType=%s", register.RegisterType())
	}
}

func FinderReadStringRegister(c *DeviceStruct, register FinderRegister) (v dataflow.Value, err error) {
	response, err := FinderReadInputRegisters(c, register)
	log.Printf("FinderReadStringRegister: got response=%s, err=%s", response, err)
	if err != nil {
		return nil, err
	}

	v = dataflow.NewTextRegisterValue(
		c.Name(),
		register,
		string(response),
	)

	return
}

func FinderReadFloatRegister(c *DeviceStruct, register FinderRegister) (v dataflow.Value, err error) {
	response, err := FinderReadInputRegisters(c, register)
	log.Printf("FinderReadFloatRegister: got response=%s, err=%s", response, err)
	if err != nil {
		return nil, err
	}

	floatValue, err := bytesToFloat32(response)
	if err != nil {
		return nil, err
	}

	v = dataflow.NewNumericRegisterValue(
		c.Name(),
		register,
		float64(floatValue),
	)

	return
}

func FinderReadInputRegisters(c *DeviceStruct, register FinderRegister) (response []byte, err error) {
	var requestPayload bytes.Buffer

	// write starting register
	err = binary.Write(&requestPayload, byteOrder, register.addressBegin-InputRegisterAddressOffset)
	if err != nil {
		return
	}
	// write register count
	countRegisters := register.addressEnd - register.addressBegin + 1
	err = binary.Write(&requestPayload, byteOrder, countRegisters)
	if err != nil {
		return
	}

	return callFunction(
		c.modbus.WriteRead,
		c.modbusConfig.Address(),
		FinderFunctionReadInputRegisters,
		requestPayload.Bytes(),
		int(2*countRegisters), // finder registers are 16 bit wide
	)
}

func bytesToFloat32(inp []byte) (float32, error) {
	var f float32

	buf := bytes.NewReader(inp)
	if err := binary.Read(buf, binary.LittleEndian, &f); err != nil {
		return 0, fmt.Errorf("bytesToFloat32 failed: %s", err)
	}

	return f, nil
}
