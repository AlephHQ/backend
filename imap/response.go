package imap

import (
	"bufio"
	"log"
	"strings"
)

const (
	space = ' '
	star  = '*'
)

type Response struct {
	// Raw contains the original response in its raw format
	Raw string

	// A tag associated with the imap response
	// can also be a * if this is a untagged response
	Tag string

	// Status Response
	StatusResp StatusResponse

	// StatusResponseCode
	StatusRespCode StatusResponseCode

	// Arguments
	Arguments interface{}
}

func readAtom(reader *bufio.Reader) (string, error) {
	atom := ""

	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return "", err
		}

		if r == space {
			break
		}

		atom += string(r)
	}

	return atom, nil
}

func Parse(raw string) *Response {
	resp := &Response{Raw: raw}
	reader := bufio.NewReader(strings.NewReader(resp.Raw))

	// Attempt to read the Tag
	atom, err := readAtom(reader)
	if err != nil {
		log.Panic(err)
	}
	resp.Tag = atom

	return resp
}
