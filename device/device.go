package device

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"time"
)

type Config interface {
	Name() string
	SkipFields() []string
	SkipCategories() []string
	TelemetryViaMqttClients() []string
	RealtimeViaMqttClients() []string
	LogDebug() bool
	LogComDebug() bool
}

type Device interface {
	Config() Config
	Registers() dataflow.Registers
	GetRegister(registerName string) dataflow.Register
	LastUpdated() time.Time
	Model() string
	Shutdown()
	ShutdownChan() chan struct{}
}
