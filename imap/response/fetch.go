package response

import (
	"fmt"
	"log"
	"ncp/backend/imap"
)

type Fetch struct {
	Tag      string
	Messages []*imap.Message
	Done     chan bool
}

func NewHandlerFetch(tag string) *Fetch {
	return &Fetch{
		Tag:      tag,
		Messages: make([]*imap.Message, 0),
		Done:     make(chan bool),
	}
}

func (f *Fetch) Handle(resp *Response) (bool, error) {
	status := imap.StatusResponse(resp.Fields[1])
	switch status {
	case imap.StatusResponseOK:
		go func() { f.Done <- true }() // go channels are so damn cool
		return true, nil
	case imap.StatusResponseNO:
		return true, fmt.Errorf("error Fetching: %s", resp.Fields[2])
	}

	msgStatusRespCode := imap.MessageStatusResponseCode(resp.Fields[2])
	if msgStatusRespCode == imap.MessageStatusResponseCodeFetch {
		msg, err := ParseMessage(resp)
		if err != nil {
			log.Panic(err)
		}

		f.Messages = append(f.Messages, msg)
		return false, nil
	}

	return false, imap.ErrUnhandled
}
