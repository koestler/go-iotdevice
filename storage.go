package main

import (
	"context"
	"github.com/koestler/go-iotdevice/dataflow"
)

func runStorage(logPrefix string) *dataflow.ValueStorage {
	storageInstance := dataflow.NewValueStorage()

	if len(logPrefix) > 0 {
		subscription := storageInstance.Subscribe(context.Background(), dataflow.NullFilter)
		dataflow.SinkLog(logPrefix, subscription.Drain())
	}

	return storageInstance
}
