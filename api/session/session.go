package session

import (
	"ncp/backend/api"
	"ncp/backend/imap/client"
	"ncp/backend/utils"
	"net/http"
	"time"
)

type Session struct {
	ID string

	Client *client.Client
}

func NewSession() *Session {
	return &Session{
		ID: utils.RandStr(20),
	}
}

var sessions = make(map[string]*Session)

// SetCookie starts a new session and sets
// a cookie with the corresponding session id
func SetCookie(w http.ResponseWriter) {
	s := NewSession()

	sessions[s.ID] = s
	http.SetCookie(
		w,
		&http.Cookie{
			Name:    string(api.CookieNameSession),
			Value:   s.ID,
			Path:    "/",
			Expires: time.Now().AddDate(0, 0, 1),
		},
	)
}

// Session returns the corresponding session
// to this request's cookie
func GetSession(r http.Request) *Session {
	c, err := r.Cookie(string(api.CookieNameSession))
	if err != nil {
		panic("session cookie not found")
	}

	return sessions[c.Value]
}
