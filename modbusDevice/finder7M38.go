package modbusDevice

// protocol documentations:
// - https://cdn.findernet.com/app/uploads/Benutzerhandbuch_Typ_7M38_DE.pdf
// - https://cdn.findernet.com/app/uploads/2021/09/20090052/Modbus-7M24-7M38_v2_30062021.pdf

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
	addr                  uint16
	sort                  int
	cat, name, desc, unit string
}{
	{32480, 620, "Device Info", "RunTime", "Run time", "s"},
	{32484, 0, "Essential", "UAvgPN", "Uavg (phase to neutral)", "V"},
	{32486, 1, "Essential", "UAvgPP", "Uavg (phase to phase)", "V"},
	{32488, 100, "Current", "SI", "S I", "A"},
	{32490, 0, "Essential", "Pt", "Active Power Total", "W"},
	{32492, 200, "Power", "Qt", "Reactive Power Total", "var"},
	{32494, 201, "Power", "St", "Apparent Power Total", "VA"},
	{32496, 2, "Essential", "PFt", "Power Factor Total", ""},
	{32498, 3, "Essential", "F", "Frequency", "Hz"},
	{32500, 4, "Essential", "U1", "U1", "V"},
	{32502, 5, "Essential", "U2", "U2", "V"},
	{32504, 6, "Essential", "U3", "U3", "V"},
	{32508, 300, "Phase Geometry", "U12", "U12", "V"},
	{32510, 301, "Phase Geometry", "U23", "U23", "V"},
	{32512, 302, "Phase Geometry", "U31", "U31", "V"},
	{32516, 101, "Current", "I1", "I1", "A"},
	{32518, 102, "Current", "I2", "I2", "A"},
	{32520, 103, "Current", "I3", "I3", "A"},
	{32524, 104, "Current", "INCalc", "I neutral (calculated)", "A"},
	{32526, 105, "Current", "InMeas", "I neutral (measured)", "A"},
	{32528, 106, "Current", "Iavg", "Iavg", "A"},
	{32530, 7, "Essential", "P1", "Active Power Phase L1", "W"},
	{32532, 7, "Essential", "P2", "Active Power Phase L2", "W"},
	{32534, 8, "Essential", "P3", "Active Power Phase L3", "W"},
	{32538, 202, "Power", "Q1", "Reactive Power Phase L1", "var"},
	{32540, 203, "Power", "Q2", "Reactive Power Phase L2", "var"},
	{32542, 204, "Power", "Q3", "Reactive Power Phase L3", "var"},
	{32544, 205, "Power", "Qt", "Reactive Power Total", "var"},
	{32546, 206, "Power", "S1", "Apparent Power Phase L1 ", "VA"},
	{32548, 207, "Power", "S2", "Apparent Power Phase L2 ", "VA"},
	{32550, 208, "Power", "S3", "Apparent Power Phase L3 ", "VA"},
	{32552, 209, "Power", "St", "Apparent Power Total", "VA"},
	{32554, 210, "Power", "PF1", "Power Factor Phase 1", ""},
	{32556, 211, "Power", "PF2", "Power Factor Phase 2", ""},
	{32558, 212, "Power", "PF3", "Power Factor Phase 3", ""},
	{32560, 213, "Power", "PFt", "Power Factor Total", ""},
	{32562, 214, "Power", "PF1", "CAP/IND P. F. Phase 1", ""},
	{32564, 215, "Power", "PF2", "CAP/IND P. F. Phase 2", ""},
	{32566, 216, "Power", "PF3", "CAP/IND P. F. Phase 3", ""},
	{32568, 217, "Power", "PFt", "CAP/IND P. F. Total", ""},
	{32570, 310, "Phase Geometry", "J1", "j1 (angle between U1 and I1)", "°"},
	{32572, 311, "Phase Geometry", "J2", "j2 (angle between U2 and I2)", "°"},
	{32574, 312, "Phase Geometry", "J3", "j3 (angle between U3 and I3) ", "°"},
	{32576, 313, "Phase Geometry", "Jt", "Power Angle Total (atan2(Pt,Qt))", "°"},
	{32578, 314, "Phase Geometry", "J12", "j12 (angle between U1 and U2)", "°"},
	{32580, 315, "Phase Geometry", "J23", "j23 (angle between U2 and U3)", "°"},
	{32582, 316, "Phase Geometry", "J31", "j31 (angle between U3 and U1)", "°"},
	{32588, 400, "Distortion", "I1Thd", "I1 THD", "%"},
	{32590, 401, "Distortion", "I2Thd", "I2 THD", "%"},
	{32592, 402, "Distortion", "I3Thd", "I3 THD", "%"},
	{32594, 403, "Distortion", "U1Thd", "U1 THD", "%"},
	{32596, 404, "Distortion", "U2Thd", "U2 THD", "%"},
	{32598, 405, "Distortion", "U3Thd", "U3 THD", "%"},
	{32638, 500, "Energy Counter", "EcN1", "Energy Counter n1", "Wh"},
	{32640, 501, "Energy Counter", "EcN2", "Energy Counter n2", "varh"},
	{32642, 502, "Energy Counter", "EcN3", "Energy Counter n3", "Wh"},
	{32644, 503, "Energy Counter", "EcN4", "Energy Counter n4", "varh"},
	{32658, 9, "Essential", "InternalTemp", "Internal Temperature", "°C"},
	{32985, 610, "Device Info", "Unom", "nominal phase voltage", "V"},
	{32987, 611, "Device Info", "Inom", "nominal phase current", "A"},
	{32989, 612, "Device Info", "Pnom", "nominal phase power", "W"},
	{32991, 613, "Device Info", "Ptot", "nominal total power", "W"},
	{32993, 614, "Device Info", "Itot", "nominal total current", "A"},
	{32995, 615, "Device Info", "Fnom", "nominal frequency", "Hz"},
}

var invalidEnum = map[int]string{
	0: "valid",
	1: "invalid",
}

var RegisterList7M38 = func() []FinderRegister {
	productRegisters := []FinderRegister{
		NewFinderRegister(
			"Device Info",
			"ModelNumber",
			"Model Number",
			FinderTStr16,
			30001, 30008,
			nil, -1,
			"",
			600,
		),
		NewFinderRegister(
			"Device Info",
			"SerialNumber",
			"Serial Number",
			FinderTStr8,
			30009, 30012,
			nil, -1,
			"",
			601,
		),
		NewFinderRegister(
			"Device Info",
			"SoftwareReference",
			"Software Reference",
			FinderT1,
			30013, 30013,
			nil, -1,
			"",
			602,
		),
		NewFinderRegister(
			"Device Info",
			"HardwareReference",
			"Hardware Reference",
			FinderTStr2,
			30014, 30014,
			nil, -1,
			"",
			603,
		),
		NewFinderRegister(
			"Essential",
			"P1valid",
			"Phase 1 measurement",
			FinderT1,
			30101, 30101,
			invalidEnum,
			0,
			"",
			50,
		),
		NewFinderRegister(
			"Essential",
			"P2valid",
			"Phase 2 measurement",
			FinderT1,
			30101, 30101,
			invalidEnum,
			1,
			"",
			51,
		),
		NewFinderRegister(
			"Essential",
			"P3valid",
			"Phase 3 measurement",
			FinderT1,
			30101, 30101,
			invalidEnum,
			2,
			"",
			52,
		),
	}

	ret := make([]FinderRegister, 0, len(productRegisters)+len(RegisterList7M38FloatRegisters))
	ret = append(ret, productRegisters...)

	for _, fr := range RegisterList7M38FloatRegisters {
		ret = append(ret,
			NewFinderRegister(
				fr.cat,
				fr.name,
				fr.desc,
				FinderTFloat,
				fr.addr, fr.addr+1,
				nil, -1,
				fr.unit,
				fr.sort,
			),
		)
	}

	return ret
}

func FinderReadRegister(c *DeviceStruct, reg FinderRegister) (v dataflow.Value, err error) {
	if c.Config().LogComDebug() {
		log.Printf("finder7N38Device[%s]: FinderReadRegister, registerName=%s, addressBegin=%d, addressEnd=%d",
			c.Name(), reg.Name(), reg.addressBegin, reg.addressEnd,
		)
	}

	switch rt := reg.RegisterType(); rt {
	case dataflow.NumberRegister:
		switch frt := reg.registerType; frt {
		case FinderTFloat:
			return FinderReadFloatRegister(c, reg)
		case FinderT1:
			return FinderReadUInt16Register(c, reg)
		default:
			return nil, fmt.Errorf("FinderReadRegister does not implement finderRegisterType=%d", frt)
		}
	case dataflow.EnumRegister:
		return FinderReadEnumRegister(c, reg)
	case dataflow.TextRegister:
		return FinderReadStringRegister(c, reg)
	default:
		return nil, fmt.Errorf("FinderReadRegister does not implement registerType=%s", rt)
	}
}

func FinderReadFloatRegister(c *DeviceStruct, register FinderRegister) (v dataflow.Value, err error) {
	response, err := FinderReadInputRegisters(c, register)
	if err != nil {
		return nil, err
	}

	var floatValue float32
	buf := bytes.NewReader(response)
	if err := binary.Read(buf, binary.BigEndian, &floatValue); err != nil {
		return nil, fmt.Errorf("conversion to float32 failed: %s", err)
	}

	if c.Config().LogDebug() {
		log.Printf("FinderReadFloatRegister: registerName=%s, floatValue=%f", register.Name(), floatValue)
	}

	v = dataflow.NewNumericRegisterValue(
		c.Name(),
		register,
		float64(floatValue),
	)

	return
}

func FinderReadUInt16Register(c *DeviceStruct, register FinderRegister) (v dataflow.Value, err error) {
	response, err := FinderReadInputRegisters(c, register)
	if err != nil {
		return nil, err
	}

	var uint16Value uint16
	buf := bytes.NewReader(response)
	if err := binary.Read(buf, binary.BigEndian, &uint16Value); err != nil {
		return nil, fmt.Errorf("conversion to uint16 failed: %s", err)
	}

	if c.Config().LogDebug() {
		log.Printf("FinderReadUInt16Register: registerName=%s, uint16Value=%d", register.Name(), uint16Value)
	}

	v = dataflow.NewNumericRegisterValue(
		c.Name(),
		register,
		float64(uint16Value),
	)

	return
}

func FinderReadEnumRegister(c *DeviceStruct, register FinderRegister) (v dataflow.Value, err error) {
	response, err := FinderReadInputRegisters(c, register)
	if err != nil {
		return nil, err
	}

	var uint16Value uint16
	buf := bytes.NewReader(response)
	if err := binary.Read(buf, binary.BigEndian, &uint16Value); err != nil {
		return nil, fmt.Errorf("conversion to uint16 failed: %s", err)
	}

	if c.Config().LogDebug() {
		log.Printf("FinderReadEnumRegister: registerName=%s, uint16Value=%d", register.Name(), uint16Value)
	}

	enumIdx := int(uint16Value)

	if bit := register.bit; bit >= 0 {
		enumIdx = (enumIdx >> bit) & 1
	}

	if _, ok := register.Enum()[enumIdx]; !ok {
		return nil, fmt.Errorf("invalid enumIdx=%d", enumIdx)
	}

	v = dataflow.NewEnumRegisterValue(
		c.Name(),
		register,
		enumIdx,
	)

	return
}

func FinderReadStringRegister(c *DeviceStruct, register FinderRegister) (v dataflow.Value, err error) {
	response, err := FinderReadInputRegisters(c, register)
	if err != nil {
		return nil, err
	}

	if c.Config().LogDebug() {
		log.Printf("FinderReadStringRegister: registerName=%s, stringValue=%s", register.Name(), response)
	}

	v = dataflow.NewTextRegisterValue(
		c.Name(),
		register,
		string(response),
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
	err = binary.Write(&requestPayload, byteOrder, uint16(register.CountRegisters()))
	if err != nil {
		return
	}

	// finder registers are 16 bit wide
	responsePayloadLength := register.CountBytes()

	begin := time.Now()
	response, err = callFunction(
		c.modbus.WriteRead,
		c.modbusConfig.Address(),
		FinderFunctionReadInputRegisters,
		requestPayload.Bytes(),
		1+responsePayloadLength, // 1 byte for byte count + payload
	)
	if c.Config().LogDebug() {
		log.Printf("FinderReadInputRegisters: callFunction: took=%s", time.Since(begin))
	}

	if err != nil {
		return
	}

	byteCount := response[0]

	if int(byteCount) != responsePayloadLength {
		err = fmt.Errorf("FinderReadInputRegisters: expected byte count to be %d but got %d", responsePayloadLength, byteCount)
		return
	}

	// first byte should be payload length
	response = response[1:]

	return
}
