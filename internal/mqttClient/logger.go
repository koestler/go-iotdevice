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

func logWrapper(clientName string, messageHandler MessageHandler) MessageHandler {
	return func(message Message) {
		// only log first 80 chars of payload
		const maxLen = 80
		pl := make([]byte, 0, maxLen+3)
		if len(message.Payload()) > maxLen {
			pl = append(pl, message.Payload()[:maxLen]...)
			pl = append(pl, "..."...)
		} else {
			pl = append(pl, message.Payload()...)
		}

		log.Printf("mqttClient[%s]: received: %s %s", clientName, message.Topic(), pl)
		messageHandler(message)
	}
}
