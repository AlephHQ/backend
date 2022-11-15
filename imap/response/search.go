package response

import (
	"ncp/backend/imap"
	"strconv"
)

type Search struct {
	Tag     string
	Results []uint64
}

func NewHandlerSearch(tag string) *Search {
	return &Search{
		Tag: tag,
	}
}

func (s *Search) Handle(resp *Response) (bool, error) {
	status := imap.StatusResponse(resp.Fields[1].(string))
	switch status {
	case imap.StatusResponseOK:
		return s.Tag == resp.Fields[0].(string), nil
	}

	responseCode := imap.ResponseCode(resp.Fields[1].(string))
	switch responseCode {
	case imap.ResponseCodeSearch:
		for _, strseqnum := range resp.Fields[2:] {
			seqnum, _ := strconv.ParseUint(strseqnum.(string), 10, 64)
			s.Results = append(s.Results, seqnum)
		}

		return false, nil
	}

	return false, nil
}
