package main

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/hassDiscovery"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/pool"
	"log"
)

func runHassDisovery(
	cfg *config.Config,
	stateStorage *dataflow.ValueStorage,
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
		stateStorage,
		mqttClientPool,
	)
	hd.Run()
	return hd
}
