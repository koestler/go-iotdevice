package device

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"time"
)

type Device interface {
	Name() string
	Registers() dataflow.Registers
	LastUpdated() time.Time
	Model() string
	Shutdown()
}
