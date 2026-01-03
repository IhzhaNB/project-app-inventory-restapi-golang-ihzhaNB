package utils

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Status  bool   `json:"status"`           // true untuk success, false untuk error
	Message string `json:"message"`          // Pesan untuk client
	Data    any    `json:"data,omitempty"`   // Data payload (jika success)
	Errors  any    `json:"errors,omitempty"` // Error details (jika error)
	Meta    any    `json:"meta,omitempty"`   // Metadata tambahan (pagination, dll)
}

// ResponseSuccess mengirim response success dengan data
func ResponseSuccess(w http.ResponseWriter, code int, message string, data any) {
	response := Response{
		Status:  true,
		Message: message,
		Data:    data,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// ResponseError mengirim response error dengan detail errors
func ResponseError(w http.ResponseWriter, code int, message string, errors any) {
	response := Response{
		Status:  false,
		Message: message,
		Errors:  errors,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// ResponsePagination mengirim response success dengan data dan pagination info
func ResponsePagination(w http.ResponseWriter, code int, message string, data any, pagination any) {
	response := Response{
		Status:  true,
		Message: message,
		Data:    data,
		Meta:    pagination,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// ResponseJSON mengirim response generic dengan custom status
func ResponseJSON(w http.ResponseWriter, code int, status bool, message string, data any) {
	response := Response{
		Status:  status,
		Message: message,
	}

	if status {
		response.Data = data
	} else {
		response.Errors = data
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}
