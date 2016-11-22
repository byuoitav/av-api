package jsonresp

import (
	"encoding/json"
	"net/http"
)

type response struct {
	Response string `json:"response"`
}

// New returns an error message in proper JSON format
func New(writer http.ResponseWriter, httpStatus int, message string) {
	response := &response{
		Response: message,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(httpStatus)
	writer.Write(jsonResponse)
}
