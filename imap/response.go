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
)

type ResponseType string

const (
	ResponseTypeStatusResp              ResponseType = "status"
	ResponseTypeServerMailBoxStatusResp ResponseType = "server status"
	ResponseTypeMessageStatus           ResponseType = "message status"
	ResponseTypeCommandContinuationReq  ResponseType = "continuation request"
)

var ErrNotStatusRespCode = errors.New("not a status response code")
var ErrStatusNotOK = errors.New("status not ok")

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

	// Capabilities
	Capabilities []string

	// Information
	Information string
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

		result += string(r)
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
	resp := &Response{Raw: raw}
	reader := bufio.NewReader(strings.NewReader(resp.Raw))

	// Attempt to read the Tag
	atom, err := readAtom(reader)
	if err != nil {
		log.Panic(err)
	}
	resp.Tag = atom

	// Attempt to read status response
	atom, err = readAtom(reader)
	if err != nil {
		log.Panic(err)
	}
	resp.StatusResp = StatusResponse(atom)

	// Attempt to read status response code
	code, err := readStatusRespCode(reader)

	if err != nil && err != ErrNotStatusRespCode {
		log.Panic(err)
	}

	// Only way we get here is if err is nil or we actually
	// have a status response code.
	if err != ErrNotStatusRespCode {
		code = strings.Trim(code, "[]")
		codeArray := strings.Split(code, " ")
		resp.StatusRespCode = StatusResponseCode(codeArray[0])

		if resp.StatusRespCode == StatusResponseCodeCapability {
			resp.Capabilities = append(resp.Capabilities, codeArray[1:]...)
		}
	}

	// no resp status code, read the rest
	rest, err := readTillEOF(reader)
	if err != nil {
		log.Panic(err)
	}
	resp.Information = rest

	return resp
}
