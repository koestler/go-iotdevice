package main

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/hassDiscovery"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/pool"
	"github.com/koestler/go-iotdevice/restarter"
	"log"
)

func runHassDisovery(
	cfg *config.Config,
	devicePool *pool.Pool[*restarter.Restarter[device.Device]],
	mqttClientPool *pool.Pool[mqttClient.Client],
) *hassDiscovery.HassDiscovery {
	hdConfig := cfg.HassDiscovery()

	if len(hdConfig) < 1 {
		return nil
	}

	if cfg.LogWorkerStart() {
		log.Printf("hassDiscovery: start discovery service")
	}

	hd := hassDiscovery.Create(
		hdConfig,
		devicePool,
		mqttClientPool,
	)
	hd.Run()
	return hd
}
