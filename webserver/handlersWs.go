package webserver

import (
	"github.com/gorilla/websocket"
	"net/http"
	"github.com/koestler/go-ve-sensor/dataflow"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:    4096,
	WriteBufferSize:   4096,
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWsRoundedValues(env *Environment, w http.ResponseWriter, r *http.Request) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return StatusError{500, err}
	}

	go func() {
		for _ = range time.Tick(time.Second) {
			roundedValues := env.RoundedStorage.GetMap(dataflow.Filter{})
			roundedValuesEssential := roundedValues.ConvertToEssential()
			conn.WriteJSON(roundedValuesEssential)
		}
	}()

	return nil;
}
