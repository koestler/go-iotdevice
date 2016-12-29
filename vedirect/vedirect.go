package vedirect

import (
	"errors"
	"fmt"
	"github.com/jacobsa/go-serial/serial"
	"io"
	"log"
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
		InterCharacterTimeout: 500,
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
		log.Printf("vedirect.Read n=%v, b=%x = %+q\n", n, b, b)
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

func (vd *Vedirect) RecvUntil(needle byte, maxLength int) (data []byte, err error) {
	log.Printf("vedirect.RecvUntil start, needle=%v\n", needle)

	b := make([]byte, 1)

	data = make([]byte, 0, 32)

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
	VeResponseFlagUnknownId      VeResponseFlag = 0x01
	VeResponseFlagNotSupported   VeResponseFlag = 0x02
	VeResponseFlagParameterError VeResponseFlag = 0x04
)

func (vd *Vedirect) SendVeCommand(cmd VeCommand, data []byte) (err error) {

	log.Printf("vedirect.SendVeCommand: cmd=%v\n", cmd)

	// compute and add checksum
	checksum := byte(0x55)
	checksum -= byte(cmd)
	for _, v := range data {
		checksum -= v
	}

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

	responseData := make([]byte, 0, 7)
	err = vd.RecvVeResponse(responseData)
	if err != nil {
		return err
	}

	log.Printf("vedirect.VeCommandPing end, responseData=%v\n", responseData)

	return nil
}

func (vd *Vedirect) RecvVeResponse(data []byte) (err error) {
	_, err = vd.RecvUntil(':', 1024)
	if err != nil {
		return err
	}

	data, err = vd.RecvUntil('\n', 7)
	if err != nil {
		return err
	}

	log.Printf("vedirect.RecvVeResponse len(data)=%v, data=%v = %+q\n", len(data), data, data)
	return err
}
