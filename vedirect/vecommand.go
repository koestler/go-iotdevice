package vedirect

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strconv"
)

func computeChecksum(cmd VeCommand, data []byte) (checksum byte) {
	checksum = byte(0x55)
	checksum -= byte(cmd)
	for _, v := range data {
		checksum -= v
	}
	return
}

func (vd *Vedirect) SendVeCommand(cmd VeCommand, data []byte) (err error) {
	checksum := computeChecksum(cmd, data)
	str := fmt.Sprintf(":%X%X%X\n", cmd, data, checksum)

	_, err = vd.write([]byte(str))
	return
}

func (vd *Vedirect) VeCommandPing() (err error) {
	err = vd.SendVeCommand(VeCommandPing, []byte{})
	if err != nil {
		return err
	}

	_, err = vd.RecvVeResponse(7)
	if err != nil {
		return err
	}

	return nil
}

func (vd *Vedirect) VeCommand(command VeCommand, address uint16) (values []byte, err error) {
	id := []byte{byte(address), byte(address >> 8)}
	param := append(id, 0x00)

	err = vd.SendVeCommand(command, param)
	if err != nil {
		return
	}

	var responseData []byte
	responseData, err = vd.RecvVeResponse(64)
	if err != nil {
		return
	}

	if len(responseData) < 8 {
		err = errors.New(fmt.Sprintf("responseData too short, len(resposneData)=%v\n", len(responseData)))
		log.Printf("vedirect: VeCommandGet, error: %v", err)
		return nil, err
	}

	// extract and check command
	var responseCommand VeCommand
	if s, err := strconv.ParseUint(string(responseData[0]), 16, 8); err != nil {
		err = errors.New(fmt.Sprintf("cannot parse responseCommand, s=%v, err=%v", s, err))
		log.Printf("vedirect: VeCommandGet, error: %v", err)
		return nil, err
	} else {
		responseCommand = VeCommand(s)
	}

	if VeCommandGet != responseCommand {
		err = errors.New(fmt.Sprintf("VeCommandGet != responseCommand, VeCommandGet=%v, responseCommand=%v", VeCommandGet, responseCommand))
		log.Printf("vedirect: VeCommandGet, error: %v", err)
		return nil, err
	}

	// extract data
	hexData := responseData[1:]
	if len(hexData)%2 != 0 {
		err = errors.New(fmt.Sprintf("received an odd number of hex bytes, len(hexData)=%v", len(hexData)))
		log.Printf("vedirect: VeCommandGet, error: %v", err)
		return nil, err
	}

	numbBytes := len(hexData) / 2
	binData := make([]byte, numbBytes)

	if n, err := hex.Decode(binData, hexData); err != nil || n != numbBytes {
		err = errors.New(fmt.Sprintf("hex to bin conversion failed: n=%v, err=%v", n, err))
		log.Printf("vedirect: VeCommandGet, error: %v", err)
		return nil, err
	}

	// extract and check checksum
	values = binData[:len(binData)-1]
	responseChecksum := binData[len(binData)-1]

	checksum := computeChecksum(responseCommand, values)
	if checksum != responseChecksum {
		err = errors.New(fmt.Sprintf("checksum != responseChecksum, checksum=%X, responseChecksum=%X", checksum, responseChecksum))
		log.Printf("vedirect: VeCommandGet, error: %v", err)
		return nil, err
	}

	return
}

func littleEndianBytesToUint(bytes []byte) (res uint64) {
	for i, b := range bytes {
		res |= uint64(b) << uint(i*8)
		if i >= 7 {
			break
		}
	}
	return
}

func littleEndianBytesToInt(input []byte) (res int64) {
	length := len(input)
	buf := bytes.NewReader(input)
	var err error

	switch length {
	case 1:
		var v int8
		err = binary.Read(buf, binary.LittleEndian, &v)
		res = int64(v)
	case 2:
		var v int16
		err = binary.Read(buf, binary.LittleEndian, &v)
		res = int64(v)
	case 4:
		var v int32
		err = binary.Read(buf, binary.LittleEndian, &v)
		res = int64(v)
	case 8:
		var v int64
		err = binary.Read(buf, binary.LittleEndian, &v)
		res = int64(v)
	default:
		log.Printf("vecommand: littleEndianBytesToInt: unhandled length=%v, input=%x", length, input)
		return 0
	}

	if err != nil {
		log.Printf("vecommand: littleEndianBytesToInt: binary.Read failed: %v", err)
		return 0
	}

	return
}

func (vd *Vedirect) VeCommandGet(address uint16) (value []byte, err error) {
	var rawValues []byte
	rawValues, err = vd.VeCommand(VeCommandGet, address)
	if err != nil {
		return
	}

	// check address
	responseAddress := uint16(littleEndianBytesToUint(rawValues[0:2]))
	if address != responseAddress {
		err = errors.New(fmt.Sprintf("address != responseAddress, address=%v, responseAddress=%v", address, responseAddress))
		log.Printf("vedirect: VeCommandGet, error: %v", err)
		return nil, err
	}

	// check flag
	responseFlag := VeResponseFlag(littleEndianBytesToUint(rawValues[2:3]))
	if VeResponseFlagOk != responseFlag {
		err = errors.New(fmt.Sprintf("VeResponseFlagOk != responseFlag, responseFlag=%v", responseFlag))
		log.Printf("vedirect: VeCommandGet, error: %v", err)
		return nil, err
	}

	// extract value
	value = rawValues[3:]

	return
}

func (vd *Vedirect) VeCommandGetUint(address uint16) (value uint64, err error) {

	rawValue, err := vd.VeCommandGet(address)
	if err != nil {
		return
	}

	value = littleEndianBytesToUint(rawValue)

	return
}

func (vd *Vedirect) VeCommandGetInt(address uint16) (value int64, err error) {

	rawValue, err := vd.VeCommandGet(address)
	if err != nil {
		return
	}

	value = littleEndianBytesToInt(rawValue)

	return
}

func (vd *Vedirect) RecvVeResponse(maxLength int) (data []byte, err error) {
	_, err = vd.RecvUntil(':', 1024)
	if err != nil {
		return
	}

	data, err = vd.RecvUntil('\n', maxLength)
	if err != nil {
		return nil, err
	}

	return
}
