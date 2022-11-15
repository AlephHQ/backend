package response

import "ncp/backend/imap"

type Store struct {
	Tag string
}

func NewHandlerStore(tag string) *Store {
	return &Store{
		Tag: tag,
	}
}

func (s *Store) Handle(resp *Response) (bool, error) {
	status := imap.StatusResponse(resp.Fields[1].(string))

	switch status {
	case imap.StatusResponseOK:
		return s.Tag == resp.Fields[0].(string), nil
	case imap.StatusResponseBAD, imap.StatusResponseNO:
		return true, resp.Error()
	}

	return false, nil
}
