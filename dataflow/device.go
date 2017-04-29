package dataflow

import (
	"sync"
	"errors"
)

// todo: this should be refactored in a proper DeviceStorage without global DevicesGet() method

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

func DevicesGetByName(name string) (*Device, error) {
	devices := DevicesGet()

	for _, device := range devices {
		if device.Name == name {
			return device, nil
		}
	}

	return nil, errors.New("device not found: " + name)
}
