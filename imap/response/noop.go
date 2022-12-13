package response

import "ncp/backend/imap"

type NOOP struct {
	Tag string
}

func NewHandlerNoop(tag string) *NOOP {
	return &NOOP{
		Tag: tag,
	}
}

func (n *NOOP) Handle(resp *Response) (bool, error) {
	if s, ok := resp.Fields[1].(string); ok {
		status := imap.StatusResponse(s)
		switch status {
		case imap.StatusResponseOK:
			return n.Tag == resp.Fields[0].(string), nil
		case imap.StatusResponseBAD, imap.StatusResponseNO:
			return true, resp.Error()
		}
	}

	return false, imap.ErrUnhandled
}
