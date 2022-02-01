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

	for _, cfgDev := range cfg.Devices() {
		if cfg.LogWorkerStart() {
			log.Printf(
				"deviceClient[%s]: start %s, cfgDev='%s'",
				cfgDev.Name(),
				cfgDev.Kind().String(),
				cfgDev.Device(),
			)
		}

		deviceConfig := deviceConfig{
			DeviceConfig: *cfgDev,
			logDebug:     cfg.LogDebug(),
		}

		if device, err := vedevices.RunDevice(&deviceConfig, target); err != nil {
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

type deviceConfig struct {
	config.DeviceConfig
	logDebug bool
}

func (cc *deviceConfig) LogDebug() bool {
	return cc.logDebug
}
