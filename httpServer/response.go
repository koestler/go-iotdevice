package httpServer

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type ErrorResponse struct {
	Message string `json:"message" example:"status bad request"`
}

func jsonErrorResponse(w http.ResponseWriter, status int, err error) {
	er := ErrorResponse{
		Message: err.Error(),
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(er)
}

func sendNotModified(w http.ResponseWriter, r *http.Request, etag string) (sent bool) {
	ifNonMatch := r.Header.Get("If-None-Match")
	if len(ifNonMatch) < 1 {
		return false
	}

	if strings.TrimPrefix(ifNonMatch, "W/") == etag {
		w.WriteHeader(http.StatusNotModified)
		return true
	}

	return false
}

func jsonGetResponse(w http.ResponseWriter, r *http.Request, obj interface{}) {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	hash := md5.Sum(jsonBytes)
	etag := hex.EncodeToString(hash[:])

	if sendNotModified(w, r, etag) {
		return
	}

	w.Header().Set("ETag", "W/"+etag)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func setCacheControlPublic(w http.ResponseWriter, maxAge time.Duration) {
	if maxAge < 0 {
		maxAge = 0
	}
	maxAgeSeconds := int(maxAge.Seconds()) // floor given duration to next lower second
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAgeSeconds))
}
