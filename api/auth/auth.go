package auth

import (
	"fmt"
	"net/http"
)

type AuthHandler struct{}

func NewHandlerAuth() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `{"status":"success"}`)
}
