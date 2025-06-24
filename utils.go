package main

import (
	"encoding/json"
	"net/http"
)

func writeErrorResponse(w http.ResponseWriter, message string, code int, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	
	errorResp := ErrorResponse{
		Error:   message,
		Code:    code,
		Details: details,
	}
	
	json.NewEncoder(w).Encode(errorResp)
}