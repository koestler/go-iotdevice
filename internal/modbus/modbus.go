package modbus

import (
	"fmt"
	"github.com/tarm/serial"
	"io"
	"log"
	"sync"
)

type ModbusStruct struct {
	cfg Config

	ioPort *serial.Port

	mutex sync.Mutex
}

func New(cfg Config) (*ModbusStruct, error) {
	if cfg.LogDebug() {
		log.Printf("modbus[%s]: create device=%v", cfg.Name(), cfg.Device())
	}

	options := serial.Config{
		Name:        cfg.Device(),
		Baud:        cfg.BaudRate(),
		ReadTimeout: cfg.ReadTimeout(),
	}

	ioHandle, err := serial.OpenPort(&options)
	if err != nil {
		return nil, fmt.Errorf("cannot open device %s: %w", cfg.Device(), err)
	}

	if cfg.LogDebug() {
		log.Printf("modbus[%s]: Open succeeded", cfg.Name())
	}

	return &ModbusStruct{
		cfg:    cfg,
		ioPort: ioHandle,
	}, nil
}

func (md *ModbusStruct) Name() string {
	return md.cfg.Name()
}

func (md *ModbusStruct) Shutdown() {
	if err := md.ioPort.Close(); err != nil {
		md.debugPrintf("Shutdown err=%v", err)
	} else {
		md.debugPrintf("Shutdown successful")
	}
}

func (md *ModbusStruct) WriteRead(request []byte, responseBuf []byte) error {
	md.mutex.Lock()
	defer md.mutex.Unlock()

	// flush receiver
	md.RecvFlush()

	// send request
	if _, err := md.Write(request); err != nil {
		return err
	}

	// read response or return error
	_, err := io.ReadFull(md, responseBuf)
	return err
}

func (md *ModbusStruct) Read(b []byte) (n int, err error) {
	n, err = md.ioPort.Read(b)
	if err != nil {
		md.debugPrintf("Read error: %v", err)
	} else {
		md.debugPrintf("Read b=%x len=%v", b[:n], n)
	}
	return
}

func (md *ModbusStruct) Write(b []byte) (n int, err error) {
	md.debugPrintf("Write b=%x len=%v", b, len(b))
	n, err = md.ioPort.Write(b)
	if err != nil {
		log.Printf("modbus[%s]: Write error: %v", md.cfg.Name(), err)
		return n, err
	}
	return
}

func (md *ModbusStruct) RecvFlush() {
	err := md.ioPort.Flush()
	md.debugPrintf("Flush err=%v", err)
}
