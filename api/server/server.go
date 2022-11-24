package server

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"ncp/backend/api/handlers/auth"
	"ncp/backend/api/handlers/inbox"
	"ncp/backend/api/handlers/post"
	"ncp/backend/api/handlers/posts"
	"ncp/backend/api/server/middleware"
)

type paramroute struct {
	route    string
	elements []string
	length   int
}

type ServeMux struct {
	mu sync.RWMutex
	r  *radix
	p  []paramroute
}

func NewServeMux() *ServeMux {
	return &ServeMux{}
}

func (mux *ServeMux) appendParamRoute(route string) {
	elements := strings.Split(
		strings.Trim(route, "/"),
		"/",
	)

	mux.p = append(
		mux.p,
		paramroute{
			elements: elements,
			length:   len(elements),
			route:    route,
		},
	)
}

func (mux *ServeMux) matchParamRoute(path string) (pattern string, params map[string]string) {
	matcher := strings.Split(
		strings.Trim(path, "/"),
		"/",
	)

	for _, r := range mux.p {
		if len(matcher) == r.length {
			match := true
			params = make(map[string]string)

			for i, elem := range r.elements {
				if strings.HasPrefix(elem, ":") {
					params[elem[1:]] = matcher[i]
				} else {
					match = match && elem == matcher[i]
				}
			}

			if match {
				pattern = r.route
				return
			}
		}
	}

	return
}

func (mux *ServeMux) Handle(pattern string, handler http.Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if pattern == "" {
		panic("api/mux: invalid pattern")
	}
	if handler == nil {
		panic("api/mux: nil handler")
	}

	if mux.r == nil {
		mux.r = new(radix)
	}

	if h, _ := mux.r.find(pattern); h != nil {
		panic("api/mux: multiple registrations for " + pattern)
	}

	if strings.Contains(pattern, ":") {
		mux.appendParamRoute(pattern)
	}

	mux.r.insert(pattern, handler)
}

func (mux *ServeMux) match(path string) (h http.Handler, pattern string) {
	h, pattern = mux.r.find(path)
	return
}

func (mux *ServeMux) Handler(r *http.Request) (h http.Handler, pattern string) {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	// try to get an exact match
	h, pattern = mux.match(r.URL.Path)
	if h != nil {
		return
	}

	// does this match with a param route?
	p, vals := mux.matchParamRoute(r.URL.Path)
	if p != "" && vals != nil {
		ctx := context.WithValue(r.Context(), "params", vals)

		*r = *r.Clone(ctx)
		h, pattern = mux.match(p)
	}

	if h == nil {
		h, pattern = http.NotFoundHandler(), ""
	}

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
	mux.Handle("/v1.0/posts/:id/update/:action", post.NewHandler())

	if err := http.ListenAndServe(":"+params.Port, middleware.Logger(mux)); err != nil {
		return err
	}

	return nil
}
