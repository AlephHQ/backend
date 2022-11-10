package response

import (
	"errors"
	"log"
	"strconv"

	"ncp/backend/imap"
)

type Select struct {
	Tag     string
	Mailbox *imap.Mailbox
}

func NewHandlerSelect(mbox, tag string) *Select {
	return &Select{
		Tag:     tag,
		Mailbox: imap.NewMailbox().SetName(mbox),
	}
}

func (s *Select) Handle(resp *Response) (bool, error) {
	// first, let's attempted to handle the cases where
	// we have an un untagged OK response or result response
	status := imap.StatusResponse(resp.Fields[1].(string))
	switch status {
	case imap.StatusResponseOK:
		// set read and write permissions
		if statusRespCodeFields, ok := resp.Fields[2].([]interface{}); ok {
			statusRespCode := imap.StatusResponseCode(statusRespCodeFields[0].(string))
			switch statusRespCode {
			case imap.StatusResponseCodeReadOnly, imap.StatusResponseCodeReadWrite:
				if s.Mailbox != nil {
					s.Mailbox.SetReadOnly(statusRespCode == imap.StatusResponseCodeReadOnly)
				}

				return s.Tag == resp.Fields[0], nil
			case imap.StatusResponseCodePermanentFlags:
				permflags := make([]string, 0)
				if list, ok := statusRespCodeFields[1].([]interface{}); ok {
					for _, flag := range list {
						permflags = append(permflags, flag.(string))
					}
				}

				s.Mailbox.SetPermanentFlags(permflags)
				return false, nil
			case imap.StatusResponseCodeUnseen, imap.StatusResponseCodeUIDNext, imap.StatusResponseCodeUIDValidity:
				num, err := strconv.ParseUint(statusRespCodeFields[1].(string), 10, 64)
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

				return false, nil
			}
		}

		return false, imap.ErrUnhandled
	case imap.StatusResponseNO:
		return true, errors.New( /* strings.Join(resp.Fields[2:], " " */ "error")
	}

	// now let's handle the untagged responses FLAGS, EXISTS, and RECENT
	code := imap.DataResponseCode(resp.Fields[1].(string))
	switch code {
	case imap.DataResponseCodeFlags:
		flags := make([]string, 0)
		if f, ok := resp.Fields[2].([]interface{}); ok {
			for _, flag := range f {
				flags = append(flags, flag.(string))
			}

			s.Mailbox.SetFlags(flags)
		}

		return false, nil
	}

	code = imap.DataResponseCode(resp.Fields[2].(string))
	switch code {
	case imap.DataResponseCodeExists, imap.DataResponseCodeRecent:
		num, err := strconv.ParseUint(resp.Fields[1].(string), 10, 64)
		if err != nil {
			log.Panic(err)
		}

		switch code {
		case imap.DataResponseCodeExists:
			s.Mailbox.SetExists(num)
		case imap.DataResponseCodeRecent:
			s.Mailbox.SetRecent(num)
		}

		return false, nil
	}

	return false, imap.ErrUnhandled
}
