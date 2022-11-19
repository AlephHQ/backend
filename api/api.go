package api

import (
	"net/http"

	"ncp/backend/api/auth"
)

type Mux struct {
	*http.ServeMux
}

func NewAPIMux() *Mux {
	mux := &Mux{
		ServeMux: http.NewServeMux(),
	}

	mux.Handle("/v1.0/auth", auth.NewHandlerAuth())

	return mux
}
