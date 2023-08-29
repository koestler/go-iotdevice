package main

import (
	"context"
	"github.com/koestler/go-iotdevice/dataflow"
)

func runStorage(logPrefix string) *dataflow.ValueStorage {
	valueStorage := dataflow.NewValueStorage()

	if len(logPrefix) > 0 {
		subscription := valueStorage.Subscribe(context.Background(), dataflow.NullFilter)
		dataflow.SinkLog(logPrefix, subscription.Drain())
	}

	return valueStorage
}
