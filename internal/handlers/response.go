package handlers

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Status string `json:"status"`
	Data   any    `json:"data,omitempty"`
	Error  any    `json:"error,omitempty"`
}

func writeHeader(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
}

func WriteJSON(w http.ResponseWriter, statusCode int, statusMessage string, data any) {
	writeHeader(w, statusCode)

	resp := Response{
		Status: statusMessage,
		Data:   data,
	}

	// Use NewDecoder directly on w to optimize memory better than Marshal
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, `{"status":"error","error":"json encoding failed"}`, http.StatusInternalServerError)
	}
}

// WriteError standard error response for the client
func WriteError(w http.ResponseWriter, statusCode int, message string) {
	writeHeader(w, statusCode)

	resp := Response{
		Status: "error",
		Error:  message,
	}

	_ = json.NewEncoder(w).Encode(resp)
}
