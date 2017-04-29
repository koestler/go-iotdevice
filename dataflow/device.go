package dataflow

import (
	"sync"
)

type Device struct {
	Name  string
	Model string
}

var deviceDbMutex sync.RWMutex
var deviceDb []*Device

func init() {
	deviceDb = make([]*Device, 0, 2)
}

// this function must not be used concurrently
func DeviceCreate(name, model string) (device *Device) {
	deviceDbMutex.Lock()
	defer deviceDbMutex.Unlock()

	device = &Device{
		Name:  name,
		Model: model,
	}

	deviceDb = append(deviceDb, device)

	return
}

func DevicesGet() (devices []*Device) {
	deviceDbMutex.RLock()
	defer deviceDbMutex.RUnlock()

	// copy only the slice, not the actual values such that pointers to Devices remain the same
	// -> never ever mutate a Device object
	return deviceDb
}
