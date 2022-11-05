package response

import (
	"errors"
	"ncp/backend/imap"
	"strings"
)

type Select struct {
	Tag     string
	Mailbox chan *imap.MailboxStatus
}

func NewHandlerSelect(tag string) *Select {
	return &Select{
		Tag:     tag,
		Mailbox: make(chan *imap.MailboxStatus),
	}
}

func (s *Select) Handle(resp *Response) (bool, error) {
	status := imap.StatusResponse(resp.Fields[1])
	switch status {
	case imap.StatusResponseOK:
		// c.setState(imap.SelectedState)

		// set read and write permissions
		statusRespCode := imap.StatusResponseCode(resp.Fields[3])
		switch statusRespCode {
		case imap.StatusResponseCodeReadOnly, imap.StatusResponseCodeReadWrite:
			// if c.mbox != nil {
			// 	c.mbox.SetReadOnly(statusRespCode == imap.StatusResponseCodeReadOnly)
			// }

			return true, nil
		}

		return false, imap.ErrUnhandled
	case imap.StatusResponseNO:
		return true, errors.New(strings.Join(resp.Fields[2:], " "))
	}

	return false, imap.ErrUnhandled
}
