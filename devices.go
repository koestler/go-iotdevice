package main

import (
	"github.com/koestler/go-victron-to-mqtt/config"
	"github.com/koestler/go-victron-to-mqtt/dataflow"
	"github.com/koestler/go-victron-to-mqtt/vedevices"
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

	for _, device := range cfg.Devices() {
		if cfg.LogWorkerStart() {
			log.Printf(
				"deviceClient[%s]: start %s, device='%s'",
				device.Name(),
				device.Kind().String(),
				device.Device(),
			)
		}

		deviceConfig := deviceConfig{
			DeviceConfig: *device,
			logDebug:     cfg.LogDebug(),
		}

		if device, err := vedevices.RunDevice(&deviceConfig, target); err != nil {
			log.Printf("deviceClient[%s]: start failed: %s", device.Name(), err)
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
		initiateShutdown <- errors.New("no device was started")
	}

	return devicePoolInstance
}

type deviceConfig struct {
	config.DeviceConfig
	logDebug bool
}

func (cc *deviceConfig) LogDebug() bool {
	return cc.logDebug
}
