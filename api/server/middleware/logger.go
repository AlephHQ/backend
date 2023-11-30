package middleware

import (
	"log"
	"net/http"
)

type logger struct{}

func (logger) Apply(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log.Printf("aleph/api - %s %s", r.Method, r.URL.Path)

			next.ServeHTTP(w, r)
		},
	)
}

func Logger() logger {
	return logger{}
}
