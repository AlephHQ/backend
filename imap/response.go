package imap

import (
	"bufio"
	"errors"
	"io"
	"log"
	"strings"
)

const (
	space         = ' '
	star          = '*'
	cr            = '\r'
	lf            = '\n'
	doubleQuote   = '"'
	respCodeStart = '['
	respCodeEnd   = ']'
	plus          = '+'
	listStart     = '('
	listEnd       = ')'
)

type ResponseType string

const (
	ResponseTypeStatusResp              ResponseType = "status"
	ResponseTypeServerMailBoxStatusResp ResponseType = "server mailbox status"
	ResponseTypeMessageStatus           ResponseType = "message status"
	ResponseTypeCommandContinuationReq  ResponseType = "continuation request"
)

var ErrNotStatusRespCode = errors.New("not a status response code")
var ErrStatusNotOK = errors.New("status not ok")

type Response struct {
	// Raw contains the original response in its raw format
	Raw string

	// Fields contains all the different fields received
	// in the response
	Fields []string

	// Type indicates what type of response we're dealing with.
	Type ResponseType

	// Tagged indicates whether this is a tagged response
	Tagged bool
}

func NewResponse(raw string) *Response {
	resp := &Response{}
	resp.Raw = raw
	resp.Fields = make([]string, 0)

	return resp
}

func (resp *Response) AddField(field string) {
	resp.Fields = append(resp.Fields, field)
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

func readTillEOF(reader *bufio.Reader) (string, error) {
	result := ""
	for {
		r, _, err := reader.ReadRune()
		if err == io.EOF {
			break
		}

		if err != nil {
			return "", err
		}

		if r != cr && r != lf {
			result += string(r)
		}
	}

	return strings.Trim(result, " "), nil
}

func readStatusRespCode(reader *bufio.Reader) (string, error) {
	code := ""

	// is this a status response code?
	r, _, err := reader.ReadRune()
	if err != nil {
		return "", err
	}

	if r != respCodeStart {
		reader.UnreadRune()
		return "", ErrNotStatusRespCode
	}

	err = reader.UnreadRune()
	if err != nil {
		return "", err
	}

	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return "", err
		}

		if r == respCodeEnd {
			code += string(r)
			break
		}

		code += string(r)
	}

	return code, nil
}

func Parse(raw string) *Response {
	resp := NewResponse(raw)
	reader := bufio.NewReader(strings.NewReader(resp.Raw))

	// Read the first element in the response,
	// whether that be a star (*) or a tag
	atom, err := readAtom(reader)
	if err != nil {
		log.Panic(err)
	}
	resp.AddField(atom)
	resp.Tagged = (atom != string(star)) && (atom != string(plus))
	if atom == string(plus) {
		resp.Type = ResponseTypeCommandContinuationReq
	}

	// Read the second element. This could be a status response
	// or a piece of data
	atom, err = readAtom(reader)
	if err != nil {
		log.Panic(err)
	}
	resp.AddField(atom)

	// Read the next element in line
	code, err := readStatusRespCode(reader)
	if err != nil && err != ErrNotStatusRespCode {
		log.Panic(err)
	}

	// Only way we get here is if err is nil or we actually
	// have a status response code.
	if err != ErrNotStatusRespCode {
		resp.AddField(code)
	}

	// no resp status code, read the rest
	rest, err := readTillEOF(reader)
	if err != nil {
		log.Panic(err)
	}
	resp.AddField(rest)

	return resp
}
