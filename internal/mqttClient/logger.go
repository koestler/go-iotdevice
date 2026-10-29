package mqttClient

import (
	"fmt"
	"log"
)

type logger struct {
	prefix string
}

func (l logger) Println(v ...any) {
	log.Print(l.prefix + fmt.Sprintln(v...))
}

func (l logger) Printf(format string, v ...any) {
	if len(format) > 0 && format[len(format)-1] != '\n' {
		format = format + "\n" // some log calls in paho do not add \n
	}
	log.Printf(l.prefix+format, v...)
}
