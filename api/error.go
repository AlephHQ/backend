package api

import (
	"fmt"
	"net/http"
)

func Error(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"status":"error", "message": "%s"}`, msg)
}
