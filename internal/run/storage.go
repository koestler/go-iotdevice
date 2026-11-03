package run

import (
	"context"

	"github.com/koestler/go-iotdevice/v3/internal/dataflow"
)

func Storage(logPrefix string) *dataflow.ValueStorage {
	valueStorage := dataflow.NewValueStorage()

	if len(logPrefix) > 0 {
		subscription := valueStorage.SubscribeSendInitial(context.Background(), func(curr, prev dataflow.Value) bool {
			return !prev.Equals(curr)
		})
		go dataflow.SinkLog(logPrefix, subscription.Drain())
	}

	return valueStorage
}
