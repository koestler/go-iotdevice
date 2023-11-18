package modbusDevice

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/sigurn/crc16"
)

type WriteReadBusFunc func(request []byte, responseBuf []byte) error
type FunctionCode byte

var byteOrder = binary.BigEndian
var checksumByteOrder = binary.LittleEndian

func callFunction(
	writeRead WriteReadBusFunc,
	deviceAddress byte,
	functionCode FunctionCode,
	payload []byte,
	responsePayloadLength int,
) (responsePayload []byte, err error) {
	// frame structure of request and response
	// 1 byte Device Address
	// 1 byte Function Code
	// n bytes payload
	// 2 bytes crc16 computeChecksum
	var request bytes.Buffer

	err = binary.Write(&request, byteOrder, deviceAddress)
	if err != nil {
		return
	}

	err = binary.Write(&request, byteOrder, functionCode)
	if err != nil {
		return
	}

	err = binary.Write(&request, byteOrder, payload)
	if err != nil {
		return
	}

	checksum := computeChecksum(request.Bytes())
	err = binary.Write(&request, checksumByteOrder, checksum)
	if err != nil {
		return
	}

	// slave address, function code, payload, 16bit crc
	responseLength := 1 + 1 + responsePayloadLength + 2
	response := make([]byte, responseLength+2)

	err = writeRead(request.Bytes(), response)
	if err != nil {
		return
	}

	// check computeChecksum
	received := checksumByteOrder.Uint16(response[len(response)-2:])
	computed := computeChecksum(response[:len(response)-2])
	if received != computed {
		return nil, fmt.Errorf("computeChecksum missmatch received != computed : %x != %x", received, computed)
	}

	// check slave address
	if received := response[0]; received != deviceAddress {
		return nil, fmt.Errorf("device address in response != address in request: %x != %x", received, deviceAddress)
	}

	// check function code
	if received := response[2]; received != byte(functionCode) {
		return nil, fmt.Errorf("function code in response != function code in request: %x != %x", received, functionCode)
	}

	return response[2 : len(response)-4], nil
}

var crcTable = crc16.MakeTable(crc16.CRC16_MODBUS)

func computeChecksum(data []byte) uint16 {
	return crc16.Checksum(data, crcTable)
}
