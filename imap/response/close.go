package response

import "ncp/backend/imap"

type Close struct {
	Tag string
}

func NewHandlerClose(tag string) *Close {
	return &Close{tag}
}

func (c *Close) Handle(resp *Response) (bool, error) {
	status := imap.StatusResponse(resp.Fields[1])
	switch status {
	case imap.StatusResponseOK:
		// c.setState(imap.AuthenticatedState)
		// c.mbox = nil

		return true, nil
	}

	return false, imap.ErrUnhandled
}
