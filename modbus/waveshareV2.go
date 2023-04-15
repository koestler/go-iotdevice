package modbus

// protocol documentation https://www.waveshare.com/wiki/Protocol_Manual_of_Modbus_RTU_Relay

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/sigurn/crc16"
	"io"
)

type FunctionCode byte

const (
	WaveshareFunctionReadRelay             FunctionCode = 0x01
	WaveshareFunctionReadAddressAndVersion FunctionCode = 0x03
	WaveshareFunctionWriteRelay            FunctionCode = 0x05
	WaveshareFunctionSetBaudRateAndAddress FunctionCode = 0x06
	WaveshareFunctionWriteAllRelays        FunctionCode = 0x0F
)

type Command uint16

// Open / closed is reversed compared to the documentation of Waveshare.
// However, I opened the unit and found that sending a FF00 energize the relay and the LED.
const (
	RelayOpen  Command = 0x0000
	RelayClose Command = 0xFF00
	RelayFlip  Command = 0x5500
)

var byteOrder = binary.BigEndian
var checksumByteOrder = binary.LittleEndian

func (md *Modbus) WriteRelay(deviceAddress byte, relayNr uint16, command Command) (err error) {
	if relayNr > 7 {
		return fmt.Errorf("invalid relayNr: %d, it must be between 0 and 7", relayNr)
	}

	// payload strucutre:
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

	_, err = md.callFunction(
		deviceAddress,
		WaveshareFunctionWriteRelay,
		payload.Bytes(),
		6,
	)

	return err
}

func (md *Modbus) ReadSoftwareRevision(deviceAddress byte) (version string, err error) {
	response, err := md.callFunction(
		deviceAddress,
		WaveshareFunctionReadAddressAndVersion,
		[]byte{
			0x20, 0x00, // 0x0200 read software revision
			0x00, 0x01, // number of bytes, Fixed 0x0001
		},
		4, // device address, command, number, revision of software
	)

	if err != nil {
		return version, fmt.Errorf("cannot read address and version: %s", err)
	}

	// extract version
	// Convert it to DEX and multiply by 0.01 is the value of software revision.
	version = fmt.Sprintf("V%d.%02d", response[3]/100, response[3]%100)

	return
}

func (md *Modbus) ReadRelays(deviceAddress byte) (state [8]bool, err error) {
	response, err := md.callFunction(
		deviceAddress,
		WaveshareFunctionReadRelay,
		[]byte{
			0x00, 0xFF, // fixed
			0x00, 0x01, // fixed
		},
		4, // device address, command, number, state
	)

	if err != nil {
		return state, fmt.Errorf("cannot read state of realys: %s", err)
	}

	// extract bits of response into boolean state
	for i := 0; i < 8; i += 1 {
		state[i] = (response[3] & (1 << i)) != 0
	}

	return
}

func (md *Modbus) callFunction(
	deviceAddress byte,
	functionCode FunctionCode,
	payload []byte,
	responseLength int,
) (response []byte, err error) {
	// flush any unread bytes in the receive buffer
	md.RecvFlush()

	// send request
	// frame structure:
	// 1 byte Device Address
	// 1 byte Function Code
	// n bytes payload
	// 2 bytes crc16 computeChecksum
	var packet bytes.Buffer

	err = binary.Write(&packet, byteOrder, deviceAddress)
	if err != nil {
		return
	}

	err = binary.Write(&packet, byteOrder, functionCode)
	if err != nil {
		return
	}

	err = binary.Write(&packet, byteOrder, payload)
	if err != nil {
		return
	}

	checksum := computeChecksum(packet.Bytes())
	err = binary.Write(&packet, checksumByteOrder, checksum)
	if err != nil {
		return
	}

	md.debugPrintf(
		"callFunction: request: deviceAddress=%02x, functionCode=%02x, payload=%02x, checksum=%04x",
		deviceAddress, functionCode, payload, checksum,
	)

	_, err = io.Copy(md, &packet)
	if err != nil {
		return
	}

	// read response + computeChecksum
	response = make([]byte, responseLength+2)
	_, err = io.ReadFull(md, response)
	if err != nil {
		return nil, err
	}

	md.debugPrintf(
		"callFunction: response=%02x",
		response,
	)

	// check computeChecksum
	received := checksumByteOrder.Uint16(response[len(response)-2:])
	computed := computeChecksum(response[:len(response)-2])
	if received != computed {
		return nil, fmt.Errorf("computeChecksum missmatch received != computed : %x != %x", received, computed)
	}

	return
}

var crcTable = crc16.MakeTable(crc16.CRC16_MODBUS)

func computeChecksum(data []byte) uint16 {
	return crc16.Checksum(data, crcTable)
}
