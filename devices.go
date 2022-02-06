package main

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/pkg/errors"
	"log"
)

func runDevices(
	cfg *config.Config,
	target dataflow.Fillable,
	initiateShutdown chan<- error,
) *device.DevicePool {
	devicePoolInstance := device.RunPool()

	countStarted := 0

	for _, cfgDev := range cfg.Devices() {
		if cfg.LogWorkerStart() {
			log.Printf(
				"deviceClient[%s]: start %s, cfgDev='%s'",
				cfgDev.Name(),
				cfgDev.Kind().String(),
				cfgDev.Device(),
			)
		}

		if device, err := device.RunDevice(cfgDev, target); err != nil {
			log.Printf("deviceClient[%s]: start failed: %s", cfgDev.Name(), err)
		} else {
			devicePoolInstance.AddDevice(device)
			countStarted += 1
			if cfg.LogWorkerStart() {
				log.Printf(
					"deviceClient[%s]: started",
					device.Config().Name(),
				)
			}
		}
	}

	if countStarted < 1 {
		initiateShutdown <- errors.New("no cfgDev was started")
	}

	return devicePoolInstance
}
