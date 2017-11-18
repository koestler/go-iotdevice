package webserver

import (
	"net/http"
	"strconv"
	"log"
)

func Run(bind string, port int, env *Environment) {
	router := newRouter(env)

	address := bind + ":" + strconv.Itoa(port)

	log.Printf("webserver: listening on port %v", address)
	go log.Fatal(router, http.ListenAndServe(address, router))
}
