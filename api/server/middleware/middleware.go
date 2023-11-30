package middleware

import "net/http"

type Middleware interface {
	Apply(http.Handler) http.Handler
}

func Chain(h http.Handler, mids ...Middleware) http.Handler {
	result := h
	for _, mid := range mids {
		result = mid.Apply(result)
	}

	return result
}
