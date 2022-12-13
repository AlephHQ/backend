package response

import (
	"aleph/backend/imap"
)

type Login struct {
	Tag          string
	Capabilities []string
}

func NewHandlerLogin(tag string) *Login {
	return &Login{
		Tag:          tag,
		Capabilities: make([]string, 0),
	}
}

func (l *Login) Handle(resp *Response) (bool, error) {
	if l.Tag == resp.Fields[0].(string) {
		status := imap.StatusResponse(resp.Fields[1].(string))
		switch status {
		case imap.StatusResponseNO, imap.StatusResponseBAD:
			return true, resp.Error()
		case imap.StatusResponseOK:
			if statusRespCode, ok := resp.Fields[2].([]interface{}); ok {
				code := statusRespCode[0].(string)

				if code == string(imap.StatusResponseCodeCapability) && len(statusRespCode) > 1 {
					l.Capabilities = make([]string, 0)
					for _, arg := range statusRespCode[1:] {
						l.Capabilities = append(l.Capabilities, arg.(string))
					}
				}
			}

			return true, nil
		}
	}

	return false, imap.ErrUnhandled
}
