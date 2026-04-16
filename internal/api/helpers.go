package api

import (
	"encoding/json"
	"net/http"
	"technical-challenge/internal/api/openapi"
)

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, openapi.ErrorResponse{
		Code:    code,
		Message: message,
	})
}
