package response

import (
	"ncp/backend/imap"
)

type Expunge struct {
	Tag string
}

func NewHandlerExpunge(tag string) *Expunge {
	return &Expunge{
		Tag: tag,
	}
}

func (e *Expunge) Handle(resp *Response) (bool, error) {
	status := resp.Fields[1].(string)
	switch imap.StatusResponse(status) {
	case imap.StatusResponseOK:
		return e.Tag == resp.Fields[0].(string), nil
	case imap.StatusResponseBAD, imap.StatusResponseNO:
		return true, resp.Error()
	}

	return false, imap.ErrUnhandled
}
