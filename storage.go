package main

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"log"
)

func runStorage(cfg *config.Config) *dataflow.ValueStorageInstance {
	if cfg.LogWorkerStart() {
		log.Printf("storage: setup rawStorage")
	}

	storageInstance := dataflow.ValueStorageCreate()

	if cfg.LogStorageDebug() {
		subscription := storageInstance.Drain()
		dataflow.SinkLog("storage", subscription.GetOutput())
	}

	return storageInstance
}
