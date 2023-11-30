package response

import "aleph/backend/imap"

type Create struct {
	Tag string
}

func NewHandlerCreate(tag string) *Create {
	return &Create{
		Tag: tag,
	}
}

func (c *Create) Handle(resp *Response) (bool, error) {
	status := imap.StatusResponse(resp.Fields[1].(string))
	switch status {
	case imap.StatusResponseOK:
		return c.Tag == resp.Fields[0].(string), nil
	case imap.StatusResponseBAD, imap.StatusResponseNO:
		return true, resp.Error()
	}

	return false, imap.ErrUnhandled
}
