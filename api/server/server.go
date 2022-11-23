package server

import (
	"net/http"

	"ncp/backend/api/handlers/auth"
	"ncp/backend/api/handlers/inbox"
	"ncp/backend/api/handlers/posts"
	"ncp/backend/api/server/middleware"
)

type Params struct {
	Port string
}

func Serve(params *Params) error {
	mux := http.NewServeMux()
	mux.Handle("/v1.0/auth", auth.NewHandler())
	mux.Handle("/v1.0/inbox", inbox.NewHandler())
	mux.Handle("/v1.0/posts", posts.NewHandler())

	if err := http.ListenAndServe(":"+params.Port, middleware.Logger(mux)); err != nil {
		return err
	}

	return nil
}
