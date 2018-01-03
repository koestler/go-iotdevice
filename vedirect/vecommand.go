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

	_, err = vd.Write([]byte(str))
	return
}

func (vd *Vedirect) VeCommandPing() (err error) {
	debugPrintf("vedirect: VeCommandPing begin")

	err = vd.SendVeCommand(VeCommandPing, []byte{})
	if err != nil {
		debugPrintf("vedirect: VeCommandPing end err=%v", err)
		return err
	}

	_, err = vd.RecvVeResponse()
	if err != nil {
		debugPrintf("vedirect: VeCommandPing end err=%v", err)
		return err
	}

	debugPrintf("vedirect: VeCommandPing end")

	return nil
}

func (vd *Vedirect) VeCommand(command VeCommand, address uint16) (values []byte, err error) {
	debugPrintf("vedirect: VeCommand begin command=%v, address=%x", command, address)

	id := []byte{byte(address), byte(address >> 8)}
	param := append(id, 0x00)

	err = vd.SendVeCommand(command, param)
	if err != nil {
		debugPrintf("vedirect: VeCommand end err=%v", err)
		return
	}

	var responseData []byte
	responseData, err = vd.RecvVeResponse()
	if err != nil {
		return
	}

	if len(responseData) < 8 {
		err = errors.New(fmt.Sprintf("responseData too short, len(resposneData)=%v", len(responseData)))
		debugPrintf("vedirect: VeCommand end err=%v", err)
		return nil, err
	}

	// extract and check command
	var responseCommand VeCommand
	if s, err := strconv.ParseUint(string(responseData[0]), 16, 8); err != nil {
		err = errors.New(fmt.Sprintf("cannot parse responseCommand, s=%v, err=%v", s, err))
		debugPrintf("vedirect: VeCommand end err=%v", err)
		return nil, err
	} else {
		responseCommand = VeCommand(s)
	}

	if VeCommandGet != responseCommand {
		err = errors.New(fmt.Sprintf("VeCommandGet != responseCommand, VeCommandGet=%v, responseCommand=%v", VeCommandGet, responseCommand))
		debugPrintf("vedirect: VeCommand end err=%v", err)
		return nil, err
	}

	// extract data
	hexData := responseData[1:]
	if len(hexData)%2 != 0 {
		err = errors.New(fmt.Sprintf("received an odd number of hex bytes, len(hexData)=%v", len(hexData)))
		debugPrintf("vedirect: VeCommand end err=%v", err)
		return nil, err
	}

	numbBytes := len(hexData) / 2
	binData := make([]byte, numbBytes)

	if n, err := hex.Decode(binData, hexData); err != nil || n != numbBytes {
		err = errors.New(fmt.Sprintf("hex to bin conversion failed: n=%v, err=%v", n, err))
		debugPrintf("vedirect: VeCommand end err=%v", err)
		return nil, err
	}

	// extract and check checksum
	values = binData[:len(binData)-1]
	responseChecksum := binData[len(binData)-1]

	checksum := computeChecksum(responseCommand, values)
	if checksum != responseChecksum {
		err = errors.New(fmt.Sprintf("checksum != responseChecksum, checksum=%X, responseChecksum=%X", checksum, responseChecksum))
		debugPrintf("vedirect: VeCommand end err=%v", err)
		return nil, err
	}

	debugPrintf("vedirect: VeCommand end")
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
	debugPrintf("vedirect: VeCommandGet begin address=%x", address)

	var rawValues []byte
	rawValues, err = vd.VeCommand(VeCommandGet, address)
	if err != nil {
		debugPrintf("vedirect: VeCommandGet end err=%v", err)
		return
	}

	// check address
	responseAddress := uint16(littleEndianBytesToUint(rawValues[0:2]))
	if address != responseAddress {
		err = errors.New(fmt.Sprintf("address != responseAddress, address=%x, responseAddress=%x", address, responseAddress))
		debugPrintf("vedirect: VeCommandGet end err=%v", err)
		return nil, err
	}

	// check flag
	responseFlag := VeResponseFlag(littleEndianBytesToUint(rawValues[2:3]))
	if VeResponseFlagOk != responseFlag {
		err = errors.New(fmt.Sprintf("VeResponseFlagOk != responseFlag, responseFlag=%v", responseFlag))
		debugPrintf("vedirect: VeCommandGet end err=%v", err)
		return nil, err
	}

	// extract value

	debugPrintf("vedirect: VeCommandGet end")
	return rawValues[3:], nil
}

func (vd *Vedirect) VeCommandGetUint(address uint16) (value uint64, err error) {
	debugPrintf("vedirect: VeCommandGetUint begin")

	rawValue, err := vd.VeCommandGet(address)
	if err != nil {
		debugPrintf("vedirect: VeCommandGetUint end err=%v", err)
		return
	}

	value = littleEndianBytesToUint(rawValue)
	debugPrintf("vedirect: VeCommandGetUint end value=%v", value)
	return
}

func (vd *Vedirect) VeCommandGetInt(address uint16) (value int64, err error) {
	debugPrintf("vedirect: VeCommandGetInt begin")

	rawValue, err := vd.VeCommandGet(address)
	if err != nil {
		debugPrintf("vedirect: VeCommandGetInt end err=%v", err)
		return
	}
	value = littleEndianBytesToInt(rawValue)

	debugPrintf("vedirect: VeCommandGetInt end value=%v", value)
	return
}

func (vd *Vedirect) RecvVeResponse() (data []byte, err error) {
	debugPrintf("vedirect: RecvVeResponse begin")

	for {
		// search start marker
		_, err = vd.RecvUntil(':', 1024)
		if err != nil {
			debugPrintf("vedirect: RecvVeResponse end err=%v", err)
			return nil, err
		}

		// search end marker
		data, err = vd.RecvUntil('\n', 64)
		if err != nil {
			debugPrintf("vedirect: RecvVeResponse end err=%v", err)
			return nil, err
		}

		if len(data) > 0 && data[0] == 'A' {
			debugPrintf("vedirect: RecvVeResponse async message received; ignore and read next response")
		} else {
			break;
		}
	}

	debugPrintf("vedirect: RecvVeResponse end")
	return
}
