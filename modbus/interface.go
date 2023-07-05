package modbus

import "time"

type Config interface {
	Name() string
	Device() string
	BaudRate() int
	ReadTimeout() time.Duration
	LogDebug() bool
}
