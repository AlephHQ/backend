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

var ErrNotStatusRespCode = errors.New("not a status response code")

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

	return result, nil
}

func readCode(reader *bufio.Reader) (string, error) {
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
	code, err := readCode(reader)

	if err != nil && err != ErrNotStatusRespCode {
		log.Panic(err)
	}

	if err != ErrNotStatusRespCode {
		resp.StatusRespCode = StatusResponseCode(code)
	}

	// no resp status code, read the rest
	rest, err := readTillEOF(reader)
	if err != nil {
		log.Panic(err)
	}
	resp.Information = rest

	return resp
}
