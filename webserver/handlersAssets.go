package webserver

import (
	"net/http"
	"github.com/gorilla/mux"
	"bytes"
	"io"
)

//go:generate ../frontend_to_bindata.sh

func HandleAssetsGet(env *Environment, w http.ResponseWriter, r *http.Request) Error {
	vars := mux.Vars(r)

	path := vars["Path"]
	if path == "" {
		path = "index.html"
	}

	if bs, err := Asset(path); err != nil {
		return StatusError{404, err}
	} else {
		var reader = bytes.NewBuffer(bs)
		io.Copy(w, reader)
	}
	return nil
}
