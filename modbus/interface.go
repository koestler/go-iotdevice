package modbus

import "time"

type Config interface {
	Name() string
	Device() string
	BaudRate() int
	ReadTimeout() time.Duration
	LogDebug() bool
}

type Modbus interface {
	Name() string
	Shutdown()
	WriteRead(request []byte, responseBuf []byte) error
}
