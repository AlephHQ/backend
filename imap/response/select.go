package response

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"ncp/backend/imap"
)

type Select struct {
	Tag     string
	Mailbox *imap.MailboxStatus
}

func NewHandlerSelect(mbox, tag string) *Select {
	return &Select{
		Tag:     tag,
		Mailbox: imap.NewMailboxStatus().SetName(mbox),
	}
}

func (s *Select) Handle(resp *Response) (bool, error) {
	// first, let's attempted to handle the cases where
	// we have an un untagged OK response or result response
	status := imap.StatusResponse(resp.Fields[1])
	switch status {
	case imap.StatusResponseOK:
		// set read and write permissions
		statusRespCode := imap.StatusResponseCode(resp.Fields[3])
		switch statusRespCode {
		case imap.StatusResponseCodeReadOnly, imap.StatusResponseCodeReadWrite:
			if s.Mailbox != nil {
				s.Mailbox.SetReadOnly(statusRespCode == imap.StatusResponseCodeReadOnly)
			}

			return s.Tag == resp.Fields[0], nil
		case imap.StatusResponseCodePermanentFlags:
			s.Mailbox.SetPermanentFlags(strings.Split(strings.Trim(resp.Fields[4], "()"), " "))
		case imap.StatusResponseCodeUnseen, imap.StatusResponseCodeUIDNext, imap.StatusResponseCodeUIDValidity:
			num, err := strconv.ParseUint(resp.Fields[4], 10, 64)
			if err != nil {
				log.Panic(err)
			}

			switch imap.StatusResponseCode(statusRespCode) {
			case imap.StatusResponseCodeUnseen:
				s.Mailbox.SetUnseen(num)
			case imap.StatusResponseCodeUIDNext:
				s.Mailbox.SetUIDNext(num)
			case imap.StatusResponseCodeUIDValidity:
				s.Mailbox.SetUIDValidity(num)
			}
		}

		return false, imap.ErrUnhandled
	case imap.StatusResponseNO:
		return true, errors.New(strings.Join(resp.Fields[2:], " "))
	}

	// now let's handle the untagged responses FLAGS, EXISTS, and RECENT
	code := imap.DataResponseCode(resp.Fields[1])
	switch code {
	case imap.DataResponseCodeFlags:
		flags := strings.Split(resp.Fields[3], " ")
		if s.Mailbox != nil {
			s.Mailbox.SetFlags(flags)
		}

		return false, nil
	}

	code = imap.DataResponseCode(resp.Fields[2])
	switch code {
	case imap.DataResponseCodeExists, imap.DataResponseCodeRecent:
		num, err := strconv.ParseUint(resp.Fields[1], 10, 64)
		if err != nil {
			log.Panic(err)
		}

		if s.Mailbox != nil {
			switch code {
			case imap.DataResponseCodeExists:
				s.Mailbox.SetExists(num)
				return false, nil
			case imap.DataResponseCodeRecent:
				s.Mailbox.SetRecent(num)
				return false, nil
			}
		}
	}

	return false, imap.ErrUnhandled
}
