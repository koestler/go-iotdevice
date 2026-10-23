package modbusDevice

// protocol documentation https://www.waveshare.com/wiki/Protocol_Manual_of_Modbus_RTU_Relay

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	WaveshareFunctionReadRelay             FunctionCode = 0x01
	WaveshareFunctionReadAddressAndVersion FunctionCode = 0x03
	WaveshareFunctionWriteRelay            FunctionCode = 0x05
)

type Command uint16

// Open / closed is reversed compared to the documentation of Waveshare.
// However, I opened the unit and found that sending a FF00 energize the relay and the LED.
const (
	RelayOpen  Command = 0x0000
	RelayClose Command = 0xFF00
)

func WaveshareWriteRelay(writeRead WriteReadBusFunc, deviceAddress byte, relayNr uint16, command Command) (err error) {
	if relayNr > 7 {
		return fmt.Errorf("invalid relayNr: %d, it must be between 0 and 7", relayNr)
	}

	// payload structure:
	// 2 bytes for register address of controlled relay, 0x0000 - 0x0007
	// 2 bytes for command: 0xFF00 open relay, 0x0000 close relay, 0x5500 flip relay
	var payload bytes.Buffer

	err = binary.Write(&payload, byteOrder, relayNr)
	if err != nil {
		return
	}

	err = binary.Write(&payload, byteOrder, command)
	if err != nil {
		return
	}

	_, err = callFunction(
		writeRead,
		deviceAddress,
		WaveshareFunctionWriteRelay,
		payload.Bytes(),
		4,
	)

	return err
}

func WaveshareReadSoftwareRevision(writeRead WriteReadBusFunc, deviceAddress byte) (version string, err error) {
	response, err := callFunction(
		writeRead,
		deviceAddress,
		WaveshareFunctionReadAddressAndVersion,
		[]byte{
			0x20, 0x00, // 0x0200 read software revision
			0x00, 0x01, // number of bytes, Fixed 0x0001
		},
		2, // number, revision of software
	)

	if err != nil {
		return version, fmt.Errorf("cannot read address and version: %s", err)
	}

	// extract version
	// Convert it to DEX and multiply by 0.01 is the value of software revision.
	version = fmt.Sprintf("V%d.%02d", response[1]/100, response[1]%100)

	return
}

func WaveshareReadRelays(writeRead WriteReadBusFunc, deviceAddress byte) (state [8]bool, err error) {
	response, err := callFunction(
		writeRead,
		deviceAddress,
		WaveshareFunctionReadRelay,
		[]byte{
			0x00, 0xFF, // fixed
			0x00, 0x01, // fixed
		},
		2, // number, state
	)

	if err != nil {
		return state, fmt.Errorf("cannot read state of realys: %s", err)
	}

	// extract bits of response into boolean state
	for i := 0; i < 8; i += 1 {
		state[i] = (response[1] & (1 << i)) != 0
	}

	return
}
