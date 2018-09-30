package handlers

import (
	"net/http"

	"github.com/helloeave/json"
)

type (
	Response struct {
		Meta     map[string]string `json:"meta"`
		Response interface{}       `json:"response"`
	}
)

func writeHTTPResponse(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	res := Response{Response: body}
	bytes, err := json.MarshalSafeCollections(res)
	if err != nil {
		writeHTTPResponse(w, http.StatusInternalServerError, err)
		return
	}
	// No errors marshalling. Write response.
	w.WriteHeader(status)
	w.Write(bytes)
}

func writeHTTPError(w http.ResponseWriter, status int, err error) {
	body := map[string]string{"message": err.Error()}
	writeHTTPResponse(w, status, body)
}
