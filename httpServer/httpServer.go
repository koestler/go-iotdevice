package httpServer

import (
	"net/http"
	"strconv"
	"log"
)

func Run(bind string, port int, env *Environment) {
	router := newRouter(env)

	address := bind + ":" + strconv.Itoa(port)

	go func() {
		log.Printf("httpServer: listening on %v", address)
		log.Fatal(router, http.ListenAndServe(address, router))
	}()
}