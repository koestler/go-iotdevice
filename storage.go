package main

import (
	"context"
	"github.com/koestler/go-iotdevice/dataflow"
)

func runStorage(logPrefix string) *dataflow.ValueStorageInstance {
	storageInstance := dataflow.NewValueStorage()

	if len(logPrefix) > 0 {
		subscription := storageInstance.Subscribe(context.Background(), dataflow.Filter{})
		dataflow.SinkLog(logPrefix, subscription.GetOutput())
	}

	return storageInstance
}
