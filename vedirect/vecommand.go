package vedirect

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
)

func computeChecksum(cmd byte, data []byte) (checksum byte) {
	checksum = byte(0x55)
	checksum -= cmd
	for _, v := range data {
		checksum -= v
	}
	return
}

func (vd *Vedirect) SendVeCommand(cmd VeCommand, data []byte) (err error) {
	vd.debugPrintf("vedirect: SendVeCommand begin")

	checksum := computeChecksum(byte(cmd), data)
	str := fmt.Sprintf(":%X%X%X\n", cmd, data, checksum)

	_, err = vd.Write([]byte(str))

	vd.debugPrintf("vedirect: SendVeCommand end")
	return
}

func (vd *Vedirect) VeCommandPing() (err error) {
	vd.debugPrintf("vedirect: VeCommandPing begin")

	err = vd.SendVeCommand(VeCommandPing, []byte{})
	if err != nil {
		vd.debugPrintf("vedirect: VeCommandPing end err=%v", err)
		return err
	}

	_, err = vd.RecvVeResponse()
	if err != nil {
		vd.debugPrintf("vedirect: VeCommandPing end err=%v", err)
		return err
	}

	vd.debugPrintf("vedirect: VeCommandPing end")
	return nil
}

func (vd *Vedirect) VeCommandDeviceId() (deviceId VeProduct, err error) {
	vd.debugPrintf("vedirect: VeCommandDeviceId begin")

	rawValue, err := vd.VeCommand(VeCommandDeviceId, 0)
	if err != nil {
		vd.debugPrintf("vedirect: VeCommandDeviceId end err=%v", err)
		return 0, err
	}

	deviceId = VeProduct(littleEndianBytesToUint(rawValue))
	if len(deviceId.String()) < 1 {
		vd.debugPrintf("vedirect: VeCommandDeviceId end unknown deviceId=%x", rawValue)
		return 0, fmt.Errorf("unknownw deviceId=%x", rawValue)
	}

	vd.debugPrintf("vedirect: VeCommandDeviceId end deviceId=%x", deviceId)
	return deviceId, nil
}

func (vd *Vedirect) VeCommand(command VeCommand, address uint16) (values []byte, err error) {
	vd.debugPrintf("vedirect: VeCommand begin command=%v, address=%x", command, address)

	var param []byte
	if command == VeCommandGet || command == VeCommandSet {
		id := []byte{byte(address), byte(address >> 8)}
		param = append(id, 0x00)
	}

	err = vd.SendVeCommand(command, param)
	if err != nil {
		vd.debugPrintf("vedirect: VeCommand end err=%v", err)
		return
	}

	var responseData []byte
	responseData, err = vd.RecvVeResponse()
	if err != nil {
		vd.debugPrintf("vedirect: VeCommand end err=%v", err)
		return
	}

	if len(responseData) < 7 {
		err = fmt.Errorf("responseData too short, len(responseData)=%v", len(responseData))
		vd.debugPrintf("vedirect: VeCommand end err=%v", err)
		return nil, err
	}

	// extract and check command
	var response VeResponse
	s := string(responseData[0])
	if i, err := strconv.ParseUint(s, 16, 8); err != nil {
		err = fmt.Errorf("cannot parse response, address=%x, s=%v, err=%v", address, s, err)
		vd.debugPrintf("vedirect: VeCommand end err=%v", err)
		return nil, err
	} else {
		response = VeResponse(i)
	}

	expectedResponse := ResponseForCommand(command)
	if expectedResponse != response {
		err = fmt.Errorf("expectedResponse != response, expectedResponse=%v, response=%v",
			expectedResponse, response)
		vd.debugPrintf("vedirect: VeCommand end err=%v", err)
		return nil, err
	}

	// extract data
	hexData := responseData[1:]
	if len(hexData)%2 != 0 {
		err = fmt.Errorf("received an odd number of hex bytes, len(hexData)=%v", len(hexData))
		vd.debugPrintf("vedirect: VeCommand end err=%v", err)
		return nil, err
	}

	numbBytes := len(hexData) / 2
	binData := make([]byte, numbBytes)

	if n, err := hex.Decode(binData, hexData); err != nil || n != numbBytes {
		err = fmt.Errorf("hex to bin conversion failed: n=%v, err=%v", n, err)
		vd.debugPrintf("vedirect: VeCommand end err=%v", err)
		return nil, err
	}

	// extract and check checksum
	values = binData[:len(binData)-1]
	responseChecksum := binData[len(binData)-1]

	checksum := computeChecksum(byte(response), values)
	if checksum != responseChecksum {
		err = fmt.Errorf("checksum != responseChecksum, checksum=%X, responseChecksum=%X", checksum, responseChecksum)
		vd.debugPrintf("vedirect: VeCommand end err=%v", err)
		return nil, err
	}

	vd.debugPrintf("vedirect: VeCommand end")
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
		res = v
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
	vd.debugPrintf("vedirect: VeCommandGet begin address=%x", address)

	var rawValues []byte

	err = func() error {
		rawValues, err = vd.VeCommand(VeCommandGet, address)
		if err != nil {
			return err
		}

		// check address
		responseAddress := uint16(littleEndianBytesToUint(rawValues[0:2]))
		if address != responseAddress {
			return fmt.Errorf("address != responseAddress, address=%x, responseAddress=%x", address, responseAddress)
		}

		// check flag
		responseFlag := VeResponseFlag(littleEndianBytesToUint(rawValues[2:3]))
		if VeResponseFlagOk != responseFlag {
			return fmt.Errorf("VeResponseFlagOk != responseFlag, responseFlag=%v", responseFlag)
		}

		return nil
	}()
	if err != nil {
		vd.debugPrintf("vedirect: VeCommandGet end err=%s", address, err)
		return nil, err
	}

	// extract value
	vd.debugPrintf("vedirect: VeCommandGet end")
	return rawValues[3:], nil
}

func (vd *Vedirect) VeCommandGetUint(address uint16) (value uint64, err error) {
	vd.debugPrintf("vedirect: VeCommandGetUint begin")

	rawValue, err := vd.VeCommandGet(address)
	if err != nil {
		vd.debugPrintf("vedirect: VeCommandGetUint end err=%v", err)
		return
	}

	value = littleEndianBytesToUint(rawValue)
	vd.debugPrintf("vedirect: VeCommandGetUint end value=%v", value)
	return
}

func (vd *Vedirect) VeCommandGetInt(address uint16) (value int64, err error) {
	vd.debugPrintf("vedirect: VeCommandGetInt begin")

	rawValue, err := vd.VeCommandGet(address)
	if err != nil {
		vd.debugPrintf("vedirect: VeCommandGetInt end err=%v", err)
		return
	}
	value = littleEndianBytesToInt(rawValue)

	vd.debugPrintf("vedirect: VeCommandGetInt end value=%v", value)
	return
}

func (vd *Vedirect) VeCommandGetString(address uint16) (value string, err error) {
	vd.debugPrintf("vedirect: VeCommandGetString begin")

	rawValue, err := vd.VeCommandGet(address)
	if err != nil {
		vd.debugPrintf("vedirect: VeCommandGetString end err=%v", err)
		return
	}

	value = string(bytes.TrimRightFunc(rawValue, func(r rune) bool { return r == 0 }))

	vd.debugPrintf("vedirect: VeCommandGetString end value=%v", value)
	return
}

func (vd *Vedirect) RecvVeResponse() (data []byte, err error) {
	vd.debugPrintf("vedirect: RecvVeResponse begin")

	for {
		// search start marker
		_, err = vd.RecvUntil(':')
		if err != nil {
			vd.debugPrintf("vedirect: RecvVeResponse end err=%v", err)
			return nil, err
		}

		// search end marker
		data, err = vd.RecvUntil('\n')
		if err != nil {
			vd.debugPrintf("vedirect: RecvVeResponse end err=%v", err)
			return nil, err
		}

		if len(data) > 0 && data[0] == 'A' {
			vd.debugPrintf("vedirect: RecvVeResponse async message received; ignore and read next response")
		} else {
			break
		}
	}

	vd.debugPrintf("vedirect: RecvVeResponse end")
	return
}
