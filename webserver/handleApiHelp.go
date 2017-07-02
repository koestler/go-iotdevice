package webserver

import (
	"net/http"
	"errors"
	"fmt"
)

func HandleApiNotFound(env *Environment, w http.ResponseWriter, r *http.Request) error {
	err := errors.New("api method not found")
	fmt.Fprint(w, "api method not found")
	return StatusError{404, err}
}
