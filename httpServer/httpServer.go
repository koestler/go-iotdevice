package httpServer

import (
	"log"
	"net/http"
	"os"
	"strconv"
)

func Run(bind string, port int, logFilePath string, env *Environment) {
	go func() {
		router := newRouter(getLogger(logFilePath), env)
		address := bind + ":" + strconv.Itoa(port)

		log.Printf("httpServer: listening on %v", address)
		log.Fatal(router, http.ListenAndServe(address, router))
	}()
}

func getLogger(logFilePath string) (writer *log.Logger) {
	if len(logFilePath) < 1 {
		// disable logging
		log.Print("httpServer: log disabled")
		return nil
	}

	if logFilePath == "-" {
		// use stdout
		log.Print("httpServer: log to stdout")
		return log.New(os.Stdout, "httpServer: ", log.LstdFlags)
	}

	file, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("httpServer: cannot open logfile: %s", err.Error())
	}
	log.Printf("httpServer: log to file=%s", logFilePath)
	return log.New(file, "", log.LstdFlags)
}
