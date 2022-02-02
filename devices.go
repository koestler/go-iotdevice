package main

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/vedevices"
	"github.com/pkg/errors"
	"log"
)

func runDevices(
	cfg *config.Config,
	target dataflow.Fillable,
	initiateShutdown chan<- error,
) *vedevices.DevicePool {
	devicePoolInstance := vedevices.RunPool()

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

		if device, err := vedevices.RunDevice(cfgDev, target); err != nil {
			log.Printf("deviceClient[%s]: start failed: %s", cfgDev.Name(), err)
		} else {
			devicePoolInstance.AddDevice(device)
			countStarted += 1
			if cfg.LogWorkerStart() {
				log.Printf(
					"deviceClient[%s]: started",
					device.Name(),
				)
			}
		}
	}

	if countStarted < 1 {
		initiateShutdown <- errors.New("no cfgDev was started")
	}

	return devicePoolInstance
}
