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
		PortName:        portName,
		BaudRate:        19200,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
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

func (vd *Vedirect) RecvSyncHex() (err error) {
	b := make([]byte, 1)

	for max := 1024; max > 0; max -= 1 {
		n, err := vd.Read(b)
		if b[0] == 0x58 {
			return
		}
	}

	return errors.New("no : found after 1024 tries")
}
