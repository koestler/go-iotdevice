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
