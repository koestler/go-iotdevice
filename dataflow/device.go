package dataflow

import (
	"log"
)

type Device struct {
	Name  string
	Model string
}

var deviceDb []*Device

func init() {
	deviceDb = make([]*Device, 0)
}

// this function must not be used concurrently
func DeviceCreate(name, model string) (device *Device) {
	device = &Device{
		Name:  name,
		Model: model,
	}

	deviceDb = append(deviceDb, device)

	return
}

func DevicePrintToLog() {
	log.Printf("deviceDb holds the current devices:")
	for _, device := range deviceDb {
		log.Printf("- %v: %v", device.Name, device.Model)
	}
}
