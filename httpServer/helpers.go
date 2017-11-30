package httpServer

import "net/http"

func writeJsonHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Model", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func writeJpegHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}
