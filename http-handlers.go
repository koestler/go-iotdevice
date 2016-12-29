package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to the go-ve-sensor server!\n")
}

func writeJsonHeaders(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(status)
}

func HttpHandleBmv(w http.ResponseWriter, r *http.Request) {
	articles := []string{"foo", "bar"}

	writeJsonHeaders(w, http.StatusOK)

	b, err := json.MarshalIndent(articles, "", "    ")
	if err != nil {
		panic(err)
	}
	w.Write(b)
}
