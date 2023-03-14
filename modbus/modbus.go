package modbus

import (
	"bufio"
	"fmt"
	"github.com/tarm/serial"
	"log"
	"time"
)

type Modbus struct {
	ioPort         *serial.Port
	reader         *bufio.Reader
	logDebug       bool
	logDebugIndent int
}

func Open(portName string, logDebug bool) (*Modbus, error) {
	if logDebug {
		log.Printf("modbus: Open portName=%v", portName)
	}

	options := serial.Config{
		Name:        portName,
		Baud:        9600,
		ReadTimeout: time.Millisecond * 200,
	}

	ioHandle, err := serial.OpenPort(&options)
	if err != nil {
		return nil, fmt.Errorf("cannot open port: %v", portName)
	}

	if logDebug {
		log.Printf("modbus: Open succeeded portName=%v", portName)
	}

	return &Modbus{ioHandle, bufio.NewReader(ioHandle), logDebug, 0}, nil
}

func (md *Modbus) Close() (err error) {
	md.debugPrintf("modbus: Close begin")
	err = md.ioPort.Close()
	md.debugPrintf("modbus: Close end err=%v", err)
	return
}

func (md *Modbus) Read(b []byte) (n int, err error) {
	n, err = md.ioPort.Read(b)
	if err != nil {
		log.Printf("modbus: Read error: %v\n", err)
	}
	return
}

func (md *Modbus) Write(b []byte) (n int, err error) {
	md.debugPrintf("modbus: Write b=%x len=%v", b, len(b))
	for _, t := range b {
		md.debugPrintf("byte: %02x", t)
	}
	n, err = md.ioPort.Write(b)
	if err != nil {
		log.Printf("modbus: Write error: %v\n", err)
		return 0, err
	}
	return
}

func (md *Modbus) RecvFlush() {
	md.debugPrintf("modbus: RecvFlush begin")

	if err := md.ioPort.Flush(); err != nil {
		md.debugPrintf("modbus: RecvFlush err=%v", err)
	}
	md.reader.Reset(md.ioPort)

	md.debugPrintf("modbus: RecvFlush end")
}
