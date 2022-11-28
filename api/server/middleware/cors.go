package middleware

import (
	"fmt"
	"net/http"
)

type corsHeaderName string

const (
	corsHeaderNameOrigin      corsHeaderName = "Access-Control-Allow-Origin"
	corsHeaderNameHeaders     corsHeaderName = "Access-Control-Allow-Headers"
	corsHeaderNameMethods     corsHeaderName = "Access-Control-Allow-Methods"
	corsHeaderNameCredentials corsHeaderName = "Access-Control-Allow-Credentials"
)

type cors struct{}

func (cors) Apply(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if o := r.Header.Get("Origin"); o != "" {
			w.Header().Add(string(corsHeaderNameOrigin), o)
			w.Header().Add(string(corsHeaderNameCredentials), "true")

			if r.Method == http.MethodOptions {
				w.Header().Set(
					string(corsHeaderNameHeaders),
					r.Header.Get("Access-Control-Request-Headers"),
				)
				w.Header().Set("Access-Control-Allow-Methods", "POST, GET, DELETE, PUT, OPTIONS")

				fmt.Fprint(w, "")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func CORS() cors {
	return cors{}
}
