package webserver

// this setup is based on the following, awesome article:
//https://elithrar.github.io/article/custom-handlers-avoiding-globals/

import (
	"net/http"
	"log"
	"github.com/koestler/go-ve-sensor/dataflow"
)

// Our application wide data containers
type Environment struct {
	RoundedStorage *dataflow.ValueStorageInstance
	Devices        []*dataflow.Device
}

// Error represents a handler error. It provides methods for a HTTP status
// code and embeds the built-in error interface.
type Error interface {
	error
	Status() int
}

// StatusError represents an error with an associated HTTP status code.
type StatusError struct {
	Code int
	Err  error
}

// Allows StatusError to satisfy the error interface.
func (statusError StatusError) Error() string {
	return statusError.Err.Error()
}

// Returns our HTTP status code.
func (statusError StatusError) Status() int {
	return statusError.Code
}

// define an extended version of http.HandlerFunc
type HandlerHandleFunc func(e *Environment, w http.ResponseWriter, r *http.Request) error

// The Handler struct that takes a configured Environment and a function matching
// our useful signature.
type Handler struct {
	Env    *Environment
	Handle HandlerHandleFunc
}

// ServeHTTP allows our Handler type to satisfy webserver.Handler.
func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := handler.Handle(handler.Env, w, r)
	log.Printf("ServeHTTP err=%v", err)

	if err != nil {
		switch e := err.(type) {
		case Error:
			// We can retrieve the status here and write out a specific
			// HTTP status code.
			log.Printf("HTTP %d - %s", e.Status(), e)
			http.Error(w, "", e.Status())
			return
		default:
			// Any error types we don't specifically look out for default
			// to serving a HTTP 500
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}
	}
}
