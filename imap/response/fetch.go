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
	status := imap.StatusResponse(resp.Fields[1].(string))
	switch status {
	case imap.StatusResponseOK:
		go func() { f.Done <- true }() // go channels are so damn cool
		return true, nil
	case imap.StatusResponseNO:
		return true, fmt.Errorf("FETCH error: %s", resp.Error())
	case imap.StatusResponseBAD:
		return true, fmt.Errorf("FETCH error: %s", resp.Error())
	}

	msgStatusRespCode := imap.ResponseCode(resp.Fields[2].(string))
	if msgStatusRespCode == imap.ResponseCodeFetch {
		msg, err := ParseMessage(resp)
		if err != nil {
			log.Panic(err)
		}

		f.Messages = append(f.Messages, msg)
		return false, nil
	}

	return false, imap.ErrUnhandled
}
