package modbus

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/sigurn/crc16"
	"io"
)

type FunctionCode uint8
type Address uint8

const (
	WaveshareFunctionReadRelay             FunctionCode = 0x01
	WaveshareFunctionReadAddressAndVersion FunctionCode = 0x03
	WaveshareFunctionWriteRelay            FunctionCode = 0x05
	WaveshareFunctionSetBaudRateAndAddress FunctionCode = 0x06
	WaveshareFunctionWriteAllRelays        FunctionCode = 0x0F
)

type Command uint16

const (
	RelayOpen  Command = 0xFF00
	RelayClose Command = 0x0000
	RelayFlip  Command = 0x5500
)

var byteOrder = binary.LittleEndian

func (md *Modbus) WriteRelay(deviceAddress Address, relayNr int, command Command) (err error) {
	if relayNr > 7 {
		return fmt.Errorf("invalid relayNr: %d, it must be between 0 and 7", relayNr)
	}

	// payload strucutre:
	// 2 bytes for register address of controlled relay, 0x0000 - 0x0007
	// 2 bytes for command: 0xFF00 open relay, 0x0000 close relay, 0x5500 flip relay
	var payload bytes.Buffer

	err = binary.Write(&payload, byteOrder, uint16(relayNr))
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

func (md *Modbus) ReadRelays(deviceAddress Address) (state [8]bool, err error) {
	response, err := md.callFunction(
		deviceAddress,
		WaveshareFunctionReadRelay,
		[]byte{
			0x00, 0x00, // relay start address 0x0000
			0x00, 0x08, // number of relays 0x0008
		},
		2, // 1 byte for number of bytes returned, 1 byte with the state
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

func (md *Modbus) callFunction(
	deviceAddress Address,
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
	err = binary.Write(&packet, byteOrder, checksum)
	if err != nil {
		return
	}

	md.debugPrintf(
		"sendFunctionCall: deviceAddress=%02x, functionCode=%02x, payload=%02x, checksum=%04x",
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

	// check computeChecksum
	received := byteOrder.Uint16(response[len(response)-2:])
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
