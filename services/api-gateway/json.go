package main

import (
	"encoding/json"
	"net/http"
)

// writeJSON sets the content type on writter, writes the status code in the header and then encode data as JSON.
func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}
