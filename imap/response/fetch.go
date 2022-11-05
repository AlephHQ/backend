package response

import (
	"fmt"
	"ncp/backend/imap"
)

type Fetch struct {
	Messages chan string
	Done     chan bool
}

func (f *Fetch) Handle(resp *Response) (bool, error) {
	status := imap.StatusResponse(resp.Fields[1])
	switch status {
	case imap.StatusResponseOK:
		close(f.Messages)
		go func() { f.Done <- true }() // go channels are so damn cool
		return true, nil
	case imap.StatusResponseNO:
		return true, fmt.Errorf("error Fetching: %s", resp.Fields[2])
	}

	msgStatusRespCode := imap.MessageStatusResponseCode(resp.Fields[2])
	if msgStatusRespCode == imap.MessageStatusResponseCodeFetch {
		f.Messages <- resp.Raw
		return false, nil
	}

	return false, imap.ErrUnhandled
}
