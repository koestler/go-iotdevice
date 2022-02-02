package vedirect

import (
	"errors"
	"fmt"
	"github.com/jacobsa/go-serial/serial"
	"io"
	"log"
)

type Vedirect struct {
	ioHandle       io.ReadWriteCloser
	logDebug       bool
	logDebugIndent int
}

func Open(portName string, logDebug bool) (*Vedirect, error) {
	if logDebug {
		log.Printf("vedirect: Open portName=%v", portName)
	}

	options := serial.OpenOptions{
		PortName:              portName,
		BaudRate:              19200,
		DataBits:              8,
		StopBits:              1,
		MinimumReadSize:       4,
		InterCharacterTimeout: 100,
	}

	ioHandle, err := serial.Open(options)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot open port: %v", portName))
	}

	if logDebug {
		log.Printf("vedirect: Open succeeded portName=%v, ioHandle=%v", portName, ioHandle)
	}

	return &Vedirect{ioHandle, logDebug, 0}, nil
}

func (vd *Vedirect) Close() (err error) {
	vd.debugPrintf("vedirect: Close begin")
	err = vd.ioHandle.Close()
	vd.debugPrintf("vedirect: Close end err=%v", err)
	return
}

func (vd *Vedirect) read(b []byte) (n int, err error) {
	n, err = vd.ioHandle.Read(b)
	if err != nil {
		log.Printf("vedirect: Read error: %v\n", err)
	}
	return
}

func (vd *Vedirect) Write(b []byte) (n int, err error) {
	vd.debugPrintf("vedirect: Write b=%s len=%v", b, len(b))
	n, err = vd.ioHandle.Write(b)
	if err != nil {
		log.Printf("vedirect: Write error: %v\n", err)
		return 0, err
	}
	return
}

func (vd *Vedirect) RecvFlush() {
	vd.debugPrintf("vedirect: RecvFlush begin")

	nBuff := 64
	b := make([]byte, nBuff)
	flushed := 0

	for {
		n, err := vd.ioHandle.Read(b)
		flushed += n

		if err == io.EOF || n < nBuff {
			// n < nBuff: read buffer empty -> we are done
			vd.debugPrintf("vedirect: RecvFlush end flushed=%v", flushed)
			return
		}
	}
}

func (vd *Vedirect) RecvUntil(needle byte, maxLength int) (data []byte, err error) {
	vd.debugPrintf("vedirect: RecvUntil begin needle=%c maxLength=%v", needle, maxLength)

	b := make([]byte, 1)
	data = make([]byte, 0, maxLength)

	for i := 0; i <= maxLength; i += 1 {
		n, err := vd.read(b)

		if err != nil {
			vd.debugPrintf("vedirect: RecvUntil end err=%v", err)
			return nil, err
		}

		if err == io.EOF || n < 1 {
			// no answer yet -> wait
			continue
		}

		if n == 1 && b[0] == needle {
			vd.debugPrintf("vedirect: RecvUntil end data=%s size=%v", data, len(data))
			return data, nil
		}

		data = append(data, b[0])
	}

	vd.debugPrintf("vedirect: RecvUntil end gave up after reaching maxLength=%v, data=%s size=%v",
		maxLength, data, len(data))

	err = errors.New(
		fmt.Sprintf("vedirect: RecvUntil end gave up after reaching maxLength=%v", maxLength),
	)

	vd.debugPrintf("vedirect: RecvUntil end err=%v", err)
	return nil, err
}
