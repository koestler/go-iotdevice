package httpServer

import (
	"net/http"
	"github.com/gorilla/mux"
	"bytes"
	"io"
	"strings"
)

//go:generate ../frontend_to_bindata.sh

func HandleAssetsGet(env *Environment, w http.ResponseWriter, r *http.Request) Error {
	vars := mux.Vars(r)

	path := vars["Path"]
	if path == "" {
		path = "index.html"
	}

	// cache static files and 404 for one day
	w.Header().Set("Cache-Control", "public, max-age=86400")

	if bs, err := Asset(path); err != nil {
		return StatusError{404, err}
	} else {
		if strings.HasSuffix(path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		} else if strings.HasSuffix(path, ".css") {
			w.Header().Set("Content-Type", "text/css")
		}

		w.WriteHeader(http.StatusOK)
		var reader = bytes.NewBuffer(bs)
		io.Copy(w, reader)
	}

	return nil
}
