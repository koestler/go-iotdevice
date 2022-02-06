package dataflow

import (
	"log"
)

func SinkLog(prefix string, input <-chan Value) {
	go func() {
		for value := range input {
			log.Printf(
				"%s: %s: %s = %v",
				prefix,
				value.DeviceName(),
				value.Register().Name(),
				value.String(),
			)
		}
	}()
}
