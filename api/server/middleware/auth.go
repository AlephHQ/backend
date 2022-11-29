package middleware

import (
	"context"
	"ncp/backend/api"
	"net/http"
)

type auth struct{}

// Apply loads authentication data, verifies and validates it,
// then stores it in the request's context
func (auth) Apply(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		csrf := r.Header.Get("X-Csrf-Token")
		if csrf == "" {
			api.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		t, err := api.GetAuthToken(r)
		if err != nil {
			api.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if t.Valid {
			if c, ok := t.Claims.(*api.JWTClaims); ok {
				if c.CsrfToken == csrf {
					ctx := context.WithValue(r.Context(), api.ContextKeyNameUserID, c.Subject)
					*r = *r.Clone(ctx)

					next.ServeHTTP(w, r)
					return
				}
			}
		}

		api.Error(w, "unauthorized", http.StatusUnauthorized)
	})
}

func Auth() auth {
	return auth{}
}
