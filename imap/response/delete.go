package response

import "aleph/backend/imap"

type Delete struct {
	Tag string
}

func NewHandlerDelete(tag string) *Delete {
	return &Delete{
		Tag: tag,
	}
}

func (d *Delete) Handle(resp *Response) (bool, error) {
	status := imap.StatusResponse(resp.Fields[1].(string))
	switch status {
	case imap.StatusResponseOK:
		return d.Tag == resp.Fields[0].(string), nil
	case imap.StatusResponseBAD, imap.StatusResponseNO:
		return true, resp.Error()
	}

	return false, imap.ErrUnhandled
}
