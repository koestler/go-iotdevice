package main

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/victron"
	"github.com/pkg/errors"
	"log"
)

func runDevices(
	cfg *config.Config,
	target dataflow.Fillable,
	initiateShutdown chan<- error,
) (devicePoolInstance *device.DevicePool) {
	// run ppool
	devicePoolInstance = device.RunPool()

	countStarted := 0
	for _, deviceConfig := range cfg.VictronDevices() {
		if cfg.LogWorkerStart() {
			log.Printf("device[%s]: start", deviceConfig.Name())
		}

		if dev, err := victron.RunDevice(deviceConfig, target); err != nil {
			log.Printf("device[%s]: start failed: %s", deviceConfig.Name(), err)
		} else {
			devicePoolInstance.AddDevice(dev)
			countStarted += 1
		}
	}

	if countStarted < 1 {
		initiateShutdown <- errors.New("no device was started")
	}

	return
}
