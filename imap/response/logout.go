package response

import (
	"ncp/backend/imap"
)

type Logout struct {
	Tag string
}

func NewHandlerLogout(tag string) *Logout {
	return &Logout{tag}
}

func (l *Logout) Handle(resp *Response) (bool, error) {
	if s, ok := resp.Fields[1].(string); ok {
		status := imap.StatusResponse(s)
		switch status {
		case imap.StatusResponseOK:
			return l.Tag == resp.Fields[0].(string), nil
		case imap.StatusResponseBAD, imap.StatusResponseNO:
			return true, resp.Error()
		}
	}

	return false, imap.ErrUnhandled
}
