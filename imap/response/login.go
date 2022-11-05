package response

import (
	"errors"
	"ncp/backend/imap"
	"strings"
)

type Login struct {
	Tag string
}

func NewHandlerLogin(tag string) *Login {
	return &Login{tag}
}

func (l *Login) Handle(resp *Response) (bool, error) {
	status := imap.StatusResponse(resp.Fields[1])
	switch status {
	case imap.StatusResponseNO:
		return true, errors.New(strings.Join(resp.Fields[5:], " "))
	case imap.StatusResponseOK:
		if resp.Fields[2] == string(imap.SpecialCharacterRespCodeStart) {
			// code := imap.StatusResponseCode(resp.Fields[3])
			// fields := strings.Split(resp.Fields[4], " ")

			// switch code {
			// case imap.StatusResponseCodeCapability:
			// 	c.capabilities = make([]string, 0)
			// 	c.capabilities = append(c.capabilities, fields[1:]...)
			// }

			return true, nil
		}
	case imap.StatusResponseBAD:
		return true, errors.New(strings.Join(resp.Fields[1:], " "))
	}

	return false, imap.ErrUnhandled
}
