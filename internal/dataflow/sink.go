package dataflow

import (
	"log"
)

func SinkLog(prefix string, input <-chan Value) {
	for value := range input {
		log.Printf(
			"%s: %s: %s",
			prefix,
			value.DeviceName(),
			value.String(),
		)
	}
}
