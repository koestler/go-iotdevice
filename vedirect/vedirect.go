package vedirect

import (
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

	return &Vedirect{io: io}, nil
}

func (vd *Vedirect) Read(b []byte) (n int, err error) {
	n, err = vd.io.Read(b)
	if err != nil {
		log.Printf("vedirect.Read error: %v\n", err)
	} else {
		//log.Printf("vedirect.Read n=%v, b=%x = %+q\n", n, b, b)
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
	log.Printf("vedirect.Recv nRequested=%v, b=%x = %+q\n", nRequested, b, b)

	return nil
}

func (vd *Vedirect) RecvFlush() (err error) {
	nBuff := 64
	b := make([]byte, nBuff)

	for {
		n, err := vd.io.Read(b)

		log.Printf("vedirect.RecvFlush n=%v, nBuff=%v", n, nBuff)

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
	log.Printf("vedirect.RecvUntil start, needle=%v\n", needle)

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
			log.Printf("vedirect.RecvUntil %v found\n", needle)
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
		log.Printf("vedirect.Write error: %v\n", err)
	} else {
		log.Printf("vedirect.Write n=%v, b=%v = %+q\n", n, b, b)
	}
	return
}

func (vd *Vedirect) Close() (err error) {
	log.Printf("vedirect.Close\n")
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

	log.Printf("vedirect.SendVeCommand: cmd=%v\n", cmd)

	checksum := computeChecksum(cmd, data)
	str := fmt.Sprintf(":%X%X%X\n", cmd, data, checksum)

	_, err = vd.write([]byte(str))
	return
}

func (vd *Vedirect) VeCommandPing() (err error) {
	log.Printf("vedirect.VeCommandPing begin\n")

	err = vd.SendVeCommand(VeCommandPing, []byte{})
	if err != nil {
		return err
	}

	var responseData []byte
	responseData, err = vd.RecvVeResponse(7)
	if err != nil {
		return err
	}

	log.Printf("vedirect.VeCommandPing end, responseData=%v\n", responseData)

	return nil
}

func (vd *Vedirect) VeCommandGet(address uint16) (value []byte, err error) {
	log.Printf("vedirect.VeCommandGet address=%X begin\n", address)

	id := []byte{byte(address), byte(address >> 8)}
	param := append(id, 0x00)

	//vd.RecvFlush()

	err = vd.SendVeCommand(VeCommandGet, param)
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
		log.Printf("vedirect.VeCommandGet end, error: %v", err)
		return
	}

	// check command
	var responseCommand VeCommand
	if s, err := strconv.ParseUint(string(responseData[0]), 16, 8); err != nil {
		err = errors.New(fmt.Sprintf("cannot parse responseCommand, s=%v, err=%v", s, err))
		log.Printf("vedirect.VeCommandGet end, error: %v", err)
		return nil, err
	} else {
		responseCommand = VeCommand(s)
	}

	if VeCommandGet != responseCommand {
		err = errors.New(fmt.Sprintf("VeCommandGet != responseCommand, VeCommandGet=%v, responseCommand=%v", VeCommandGet, responseCommand))
		log.Printf("vedirect.VeCommandGet end, error: %v", err)
		return nil, err
	}

	// check address
	var responseAddress uint16
	if s, err := strconv.ParseUint(string(responseData[1:5]), 16, 16); err != nil {
		err = errors.New(fmt.Sprintf("cannot parse responseAddress, s=%v, err=%v", s, err))
		log.Printf("vedirect.VeCommandGet end, error: %v", err)
		return nil, err
	} else {
		responseAddress = uint16(((s & 0x00FF) << 8) | ((s & 0xFF00) >> 8))
	}

	if address != responseAddress {
		err = errors.New(fmt.Sprintf("address != responseAddress, address=%v, responseAddress=%v", address, responseAddress))
		log.Printf("vedirect.VeCommandGet end, error: %v", err)
		return nil, err
	}

	// check flag
	var responseFlag VeResponseFlag
	if s, err := strconv.ParseUint(string(responseData[6:7]), 16, 8); err != nil {
		err = errors.New(fmt.Sprintf("cannot parse responseFlag, s=%v, err=%v", s, err))
		log.Printf("vedirect.VeCommandGet end, error: %v", err)
		return nil, err
	} else {
		responseFlag = VeResponseFlag(s)
	}

	if VeResponseFlagOk != responseFlag {
		err = errors.New(fmt.Sprintf("VeResponseFlagOk != responseFlag, responseFlag=%v", responseFlag))
		log.Printf("vedirect.VeCommandGet end, error: %v", err)
		return nil, err
	}

	// get value
	value = responseData[7 : len(responseData)-2]

	// check checksum
	var responseChecksum byte
	if s, err := strconv.ParseUint(string(responseData[len(responseData)-2:]), 16, 8); err != nil {
		err = errors.New(fmt.Sprintf("cannot parse responseChecksum, s=%v, err=%v", s, err))
		log.Printf("vedirect.VeCommandGet end, error: %v", err)
		return nil, err
	} else {
		responseChecksum = byte(s)
	}

	checksum := computeChecksum(responseCommand, responseData[1:len(responseData)-2])
	if checksum != responseChecksum {
		err = errors.New(fmt.Sprintf("checksum != responseChecksum, checksum=%X, responseChecksum=%X", checksum, responseChecksum))
		log.Printf("vedirect.VeCommandGet end, error: %v", err)
		return nil, err
	}

	log.Printf("vedirect.VeCommandGet end, value: %x = %s \n", value, value)

	return
}

func (vd *Vedirect) RecvVeResponse(maxLength int) (data []byte, err error) {
	_, err = vd.RecvUntil(':', 1024)
	if err != nil {
		return
	}

	data, err = vd.RecvUntil('\n', maxLength)
	if err != nil {
		return
	}

	log.Printf("vedirect.RecvVeResponse len(data)=%v, data=%v = %+q\n", len(data), data, data)
	return
}
