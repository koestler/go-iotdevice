package main

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
)

func runStorage(cfg *config.Config) *dataflow.ValueStorageInstance {
	storageInstance := dataflow.ValueStorageCreate()

	if cfg.LogStorageDebug() {
		subscription := storageInstance.Subscribe(dataflow.Filter{})
		dataflow.SinkLog("storage", subscription.GetOutput())
	}

	return storageInstance
}
