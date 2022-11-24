package post

import (
	"fmt"
	"log"
	"ncp/backend/api"
	"net/http"
)

type Handler struct{}

func NewHandler() *Handler {
	return new(Handler)
}

func (Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		log.Printf("Params: %v\n", r.Context().Value(api.ContextKeyNameParams))

		fmt.Fprint(w, `{"status":"success"}`)
	}
}
