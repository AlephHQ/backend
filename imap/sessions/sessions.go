package sessions

import (
	"errors"
	"ncp/backend/imap"
	"ncp/backend/imap/client"
	"sync"
)

type store struct {
	clients map[string]*client.Client
	mu      sync.RWMutex
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
		err = ErrParamsInvalid
		return
	}

	if defaultStore.clients == nil {
		defaultStore.clients = make(map[string]*client.Client)
	}

	defaultStore.mu.RLock()
	c = defaultStore.clients[p.Username]
	defaultStore.mu.RUnlock()

	// a valid is found
	if c != nil && c.State() == imap.SelectedState {
		return
	}

	if c == nil {
		c, err = client.DialWithTLS("tcp", "modsoussi.com:993")
		if err != nil {
			return
		}
	}

	if c.State() == imap.NotAuthenticatedState {
		err = c.Login(p.Username+"@modsoussi.com", p.Password)
		if err != nil {
			c.Logout()
			c = nil
			return
		}
	}

	if c.State() == imap.AuthenticatedState {
		err = c.Select("INBOX")
		if err != nil {
			c.Logout()
			c = nil
			return
		}
	}

	defaultStore.mu.Lock()
	defaultStore.clients[p.Username] = c
	defaultStore.mu.Unlock()

	return
}
