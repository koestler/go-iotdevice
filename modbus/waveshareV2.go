package modbus

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/sigurn/crc16"
	"io"
)

type FunctionCode uint8
type Address uint16

const (
	WaveshareFunctionReadRelay             FunctionCode = 0x01
	WaveshareFunctionReadAddressAndVersion FunctionCode = 0x03
	WaveshareFunctionWriteRelay            FunctionCode = 0x05
	WaveshareFunctionSetBaudRateAndAddress FunctionCode = 0x06
	WaveshareFunctionWriteAllRelays        FunctionCode = 0x0F
)

type Command uint16

const (
	WaveshareCommandOpen                Command = 0xFF00
	WaveshareCommandClose               Command = 0x0000
	WaveshareCommandFlip                Command = 0x5500
	WaveshareCommandReadDeviceAddress   Command = 0x400
	WavesahreCommandReadSoftwareVersion Command = 0x8000
)

var byteOrder = binary.LittleEndian

func (md *Modbus) ReadStateOfRelays(deviceAddress Address) (state [8]bool, err error) {
	response, err := md.CallFunction(
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

func (md *Modbus) CallFunction(
	deviceAddress Address,
	functionCode FunctionCode,
	payload []byte,
	responseLength int,
) (response []byte, err error) {
	// flush any unread bytes in the receive buffer
	md.RecvFlush()

	// send request
	err = md.SendFunctionCall(deviceAddress, functionCode, payload)
	if err != nil {
		return
	}

	// read response + checksum
	response = make([]byte, responseLength+2)
	_, err = io.ReadFull(md, response)
	if err != nil {
		return nil, err
	}

	// check checksum
	received := byteOrder.Uint16(response[len(response)-2:])
	computed := checksum(response[:len(response)-2])
	if received != computed {
		return nil, fmt.Errorf("checksum missmatch received != computed : %x != %x", received, computed)
	}

	return
}

func (md *Modbus) SendFunctionCall(deviceAddress Address, functionCode FunctionCode, payload []byte) (err error) {
	// frame structure:
	// 1 byte Device Address
	// 1 byte Function Code
	// n bytes payload
	// 2 bytes crc16 checksum

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

	err = binary.Write(&packet, byteOrder, checksum(packet.Bytes()))
	if err != nil {
		return
	}

	_, err = io.Copy(md, &packet)
	if err != nil {
		return
	}

	return
}

var crcTable = crc16.MakeTable(crc16.CRC16_MAXIM)

func checksum(data []byte) uint16 {
	return crc16.Checksum(data, crcTable)
}
