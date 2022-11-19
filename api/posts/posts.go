package posts

import "net/http"

type HandlerPosts struct{}

func (HandlerPosts) ServeHTTP(w http.ResponseWriter, r *http.Request) {}
