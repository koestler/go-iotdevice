package main

import (
	"github.com/koestler/go-victron-to-mqtt/config"
	"github.com/koestler/go-victron-to-mqtt/vedevices"
	"github.com/pkg/errors"
	"log"
)

func runDevices(
	cfg *config.Config,
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

		if device, err := vedevices.RunDevice(&deviceConfig); err != nil {
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

/*
func setupBmvDevices() {
	log.Printf("main: setup Bmv Devices")

	configs := config.GetVedeviceConfigs()

	sources := make([]dataflow.Drainable, 0, len(configs))

	// get devices from database and create them
	for _, c := range configs {
		log.Printf(
			"bmvDevices: setup name=%v model=%v device=%v",
			c.Name, c.Model, c.Device,
		)

		// register device in storage
		device := storage.DeviceCreate(c.Name, c.Model, c.FrontendConfig)

		// setup the datasource
		if "dummy" == c.Device {
			sources = append(sources, vedevices.CreateDummySource(device, c))
		} else {
			if err, source := vedevices.CreateSource(device, c); err == nil {
				sources = append(sources, source)
			} else {
				log.Printf("bmvDevices: error during CreateSource: %v", err)
			}
		}
	}

	// append them as sources to the raw storage
	for _, source := range sources {
		source.Append(rawStorage)
	}
}
*/
