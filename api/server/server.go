package server

import (
	"net/http"
	"sync"

	"ncp/backend/api/handlers/auth"
	"ncp/backend/api/handlers/inbox"
	"ncp/backend/api/handlers/posts"
	"ncp/backend/api/server/middleware"
)

type ServeMux struct {
	mu sync.RWMutex
	r  *radix
}

func NewServeMux() *ServeMux {
	return &ServeMux{}
}

func (mux *ServeMux) Handle(pattern string, handler http.Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if pattern == "" {
		panic("http: invalid pattern")
	}
	if handler == nil {
		panic("http: nil handler")
	}
	if h, _ := mux.r.find(pattern); h != nil {
		panic("http: multiple registrations for " + pattern)
	}

	if mux.r == nil {
		mux.r = new(radix)
	}

	mux.r.insert(pattern, handler)
}

func (mux *ServeMux) Handler(r *http.Request) (h http.Handler, pattern string) {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	h, pattern = mux.r.find(r.URL.Path)
	return
}

func (mux *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "*" {
		if r.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h, _ := mux.Handler(r)
	h.ServeHTTP(w, r)
}

type ServeParams struct {
	Port string
}

func Serve(params *ServeParams) error {
	mux := NewServeMux()
	mux.Handle("/v1.0/auth", auth.NewHandler())
	mux.Handle("/v1.0/inbox", inbox.NewHandler())
	mux.Handle("/v1.0/posts", posts.NewHandler())

	if err := http.ListenAndServe(":"+params.Port, middleware.Logger(mux)); err != nil {
		return err
	}

	return nil
}
