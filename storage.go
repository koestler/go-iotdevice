package main

import (
	"github.com/koestler/go-iotdevice/dataflow"
)

func runStorage(logPrefix string) *dataflow.ValueStorageInstance {
	storageInstance := dataflow.ValueStorageCreate()

	if len(logPrefix) > 0 {
		subscription := storageInstance.Subscribe(dataflow.Filter{})
		dataflow.SinkLog(logPrefix, subscription.GetOutput())
	}

	return storageInstance
}
