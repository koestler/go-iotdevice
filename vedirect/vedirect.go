package vedirect

import (
	"bufio"
	"fmt"
	"github.com/tarm/serial"
	"log"
	"time"
)

type Vedirect struct {
	ioPort         *serial.Port
	reader         *bufio.Reader
	logDebug       bool
	logDebugIndent int
}

func Open(portName string, logDebug bool) (*Vedirect, error) {
	if logDebug {
		log.Printf("vedirect: Open portName=%v", portName)
	}

	options := serial.Config{
		Name:        portName,
		Baud:        19200,
		ReadTimeout: time.Millisecond * 200,
	}

	ioHandle, err := serial.OpenPort(&options)
	if err != nil {
		return nil, fmt.Errorf("cannot open port: %v", portName)
	}

	if logDebug {
		log.Printf("vedirect: Open succeeded portName=%v", portName)
	}

	return &Vedirect{ioHandle, bufio.NewReader(ioHandle), logDebug, 0}, nil
}

func (vd *Vedirect) Close() (err error) {
	vd.debugPrintf("vedirect: Close begin")
	err = vd.ioPort.Close()
	vd.debugPrintf("vedirect: Close end err=%v", err)
	return
}

func (vd *Vedirect) Write(b []byte) (n int, err error) {
	vd.debugPrintf("vedirect: Write b=%s len=%v", b, len(b))
	n, err = vd.ioPort.Write(b)
	if err != nil {
		log.Printf("vedirect: Write error: %v\n", err)
		return 0, err
	}
	return
}

func (vd *Vedirect) RecvFlush() {
	vd.debugPrintf("vedirect: RecvFlush begin")

	if err := vd.ioPort.Flush(); err != nil {
		vd.debugPrintf("vedirect: RecvFlush err=%v", err)
	}
	vd.reader.Reset(vd.ioPort)

	vd.debugPrintf("vedirect: RecvFlush end")
}

func (vd *Vedirect) RecvUntil(needle byte) (data []byte, err error) {
	vd.debugPrintf("vedirect: RecvUntil needle=%c", needle)
	data, err = vd.reader.ReadBytes(needle)
	if err == nil {
		data = data[:len(data)-1] // exclude delimiter
	}
	return
}
