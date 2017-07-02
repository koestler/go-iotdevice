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
				value.Device.Name,
				value.Name,
				value.Value,
			)
		}
	}()
}
