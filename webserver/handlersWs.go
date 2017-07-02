package webserver

import (
	"github.com/gorilla/websocket"
	"github.com/gorilla/mux"
	"net/http"
	"github.com/koestler/go-ve-sensor/dataflow"
	"log"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:    4096,
	WriteBufferSize:   4096,
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWsRoundedValues(env *Environment, w http.ResponseWriter, r *http.Request) Error {
	// upgrade to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil
	}

	// get device id
	vars := mux.Vars(r)
	device, err := dataflow.DevicesGetByName(vars["DeviceId"])
	if err != nil {
		return StatusError{404, err}
	}

	// setup filter
	filter := dataflow.Filter{Devices: map[*dataflow.Device]bool{device: true}}

	// subscribe to data
	dataChan := env.RoundedStorage.Subscribe(filter)
	sinkJson(conn, dataChan)

	return nil;
}

type Message struct {
	DeviceName string
	ValueName  string
	Value      float64
}

func convertValueToMessage(value dataflow.Value) (Message) {
	return Message{
		DeviceName: value.Device.Name,
		ValueName:  value.Name,
		Value:      value.Value,
	}
}

func sinkJson(conn *websocket.Conn, input <-chan dataflow.Value) {
	go func() {
		log.Printf("SinkJson started")
		for value := range input {
			conn.WriteJSON(convertValueToMessage(value))
		}
		log.Printf("SinkJson stoped")
	}()
}
