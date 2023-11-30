package sessions

import (
	"aleph/backend/imap"
	"aleph/backend/imap/client"
	"errors"
	"sync"
)

type session struct {
	c        *client.Client
	messages []*imap.Message
}

type store struct {
	sessions map[string]*session
	mu       sync.Mutex
}

var defaultStore store

type Params struct {
	Username string
	Password string
}

var ErrParamsInvalid = errors.New("invalid params")

// Session returns a client for the (username, password)
// pair. Both username and password need to be set whenever
// Session is called to make sure a session is created
// when a valid one isn't found.
//
// if no valid session is found, a new one is created and
// added to the store.
func Session(p *Params) (c *client.Client, err error) {
	if p.Username == "" && p.Password == "" {
		panic("username and password can't be empty")
	}

	defaultStore.mu.Lock()
	defer defaultStore.mu.Unlock()

	if defaultStore.sessions == nil {
		defaultStore.sessions = make(map[string]*session)
	}

	s := defaultStore.sessions[p.Username]
	if s == nil {
		s = &session{
			c:        nil,
			messages: make([]*imap.Message, 0),
		}
	}

	if s.c != nil && s.c.State() == imap.SelectedState {
		c = s.c
		return
	}

	if s.c == nil {
		s.c, err = client.DialWithTLS("tcp", "modsoussi.com:993")
		if err != nil {
			return
		}
	}

	if s.c.State() == imap.NotAuthenticatedState {
		err = s.c.Login(p.Username+"@modsoussi.com", p.Password)
		if err != nil {
			s.c.Logout()
			s.c = nil
			return
		}
	}

	if s.c.State() == imap.AuthenticatedState {
		err = s.c.Select("INBOX")
		if err != nil {
			s.c.Logout()
			s.c = nil
			return
		}
	}

	defaultStore.sessions[p.Username] = s
	c = s.c
	return
}
