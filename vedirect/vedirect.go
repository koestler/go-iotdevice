package vedirect

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/jacobsa/go-serial/serial"
	"io"
	"log"
	"strconv"
)

type Vedirect struct {
	io io.ReadWriteCloser
}

func Open(portName string) (*Vedirect, error) {
	log.Printf("vedirect: Open portName=%v", portName)

	options := serial.OpenOptions{
		PortName:              portName,
		BaudRate:              19200,
		DataBits:              8,
		StopBits:              1,
		MinimumReadSize:       4,
		InterCharacterTimeout: 100,
	}

	io, err := serial.Open(options)
	if err != nil {
		log.Fatalf("vedirect.Open: %v\n", err)
		return nil, errors.New("cannot open port")
	}

	log.Printf("vedirect: Open succeeded portName=%v, io=%v", portName, io)

	return &Vedirect{io: io}, nil
}

func (vd *Vedirect) Read(b []byte) (n int, err error) {
	n, err = vd.io.Read(b)
	if err != nil {
		log.Printf("vedirect: Read error: %v\n", err)
	}
	return
}

func (vd *Vedirect) Recv(b []byte) (err error) {
	nRequested := len(b)

	for n := 0; n < nRequested; {
		nRead, err := vd.Read(b[n:])
		if err != nil {
			return err
		}
		n += nRead
	}
	return nil
}

func (vd *Vedirect) RecvFlush() (err error) {
	nBuff := 64
	b := make([]byte, nBuff)

	for {
		n, err := vd.io.Read(b)

		if err == io.EOF {
			return nil
		}

		if n < nBuff {
			break
		}
	}

	return nil
}

func (vd *Vedirect) RecvUntil(needle byte, maxLength int) (data []byte, err error) {
	b := make([]byte, 1)
	data = make([]byte, 0, maxLength)

	for i := 0; i <= maxLength; i += 1 {
		n, err := vd.Read(b)

		if err != nil {
			return nil, err
		}

		if err == io.EOF || n < 1 {
			// no answer yet -> wait
			continue
		}

		if n == 1 && b[0] == needle {
			return data, nil
		}

		data = append(data, b[0])
	}

	return nil, errors.New(
		fmt.Sprintf("vedirect.RecvUntil gave up after reaching maxLength=%v\n", maxLength),
	)
}

func (vd *Vedirect) write(b []byte) (n int, err error) {
	n, err = vd.io.Write(b)
	if err != nil {
		log.Printf("vedirect: Write error: %v\n", err)
		return 0, err
	}
	return
}

func (vd *Vedirect) Close() (err error) {
	return vd.io.Close()
}

type VeCommand byte

const (
	VeCommandPing       VeCommand = 0x01
	VeCommandAppVersion VeCommand = 0x03
	VeCommandDeviceId   VeCommand = 0x04
	VeCommandRestart    VeCommand = 0x06
	VeCommandGet        VeCommand = 0x07
	VeCommandSet        VeCommand = 0x08
	VeCommandAsync      VeCommand = 0x0A
)

type VeResponse byte

const (
	VeResponseDone    VeResponse = 0x01
	VeResponseUnknown VeResponse = 0x03
	VeResponsePing    VeResponse = 0x05
	VeResponseGet     VeResponse = 0x07
	VeResponseSet     VeResponse = 0x08
	VeResponseAsync   VeResponse = 0x0A
)

type VeResponseFlag byte

const (
	VeResponseFlagOk             VeResponseFlag = 0x00
	VeResponseFlagUnknownId      VeResponseFlag = 0x01
	VeResponseFlagNotSupported   VeResponseFlag = 0x02
	VeResponseFlagParameterError VeResponseFlag = 0x04
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

	//vd.RecvFlush()

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

func littleEndianBytesToInt(bytes []byte) (res int64) {
	for i, b := range bytes {
		res |= int64(b) << uint(i*8)
		if i >= 7 {
			break
		}
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
