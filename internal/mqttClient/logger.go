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
	log.Printf(l.prefix+format, v...)
}
