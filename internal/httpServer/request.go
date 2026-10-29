package httpServer

import (
	"encoding/json"
	"errors"
	"net/http"
)

const maxRequestBodyBytes = 1 << 16 // 64 KiB

// decodeJsonBody decodes the request body into v and rejects bodies larger than maxRequestBodyBytes.
func decodeJsonBody(w http.ResponseWriter, r *http.Request, v any) error {
	return json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)).Decode(v)
}

func respondInvalidBody(w http.ResponseWriter, err error) {
	if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
		jsonErrorResponse(w, http.StatusRequestEntityTooLarge, errors.New("request body too large"))
		return
	}
	jsonErrorResponse(w, http.StatusUnprocessableEntity, errors.New("invalid json body provided"))
}
