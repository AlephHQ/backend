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
var ErrFoundSpecialChar = errors.New("found a special char")
var ErrNotSpecialChar = errors.New("found a non-special char")

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

// readAtom reads until it finds a special character, and when it does
// so, it returns the atom that precedes the special character and an
// error indicating it found a special character
func readAtom(reader *bufio.Reader) (string, error) {
	atom := ""

	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return "", err
		}

		switch r {
		case space, star, cr, lf, doubleQuote, plus, respCodeStart, respCodeEnd, listStart, listEnd:
			reader.UnreadRune()
			return atom, ErrFoundSpecialChar
		default:
			atom += string(r)
		}
	}
}

// readSpecialChar reads a single special character
func readSpecialChar(reader *bufio.Reader) (rune, error) {
	r, _, err := reader.ReadRune()
	if err != nil {
		return 0, err
	}

	switch r {
	case space, star, cr, lf, doubleQuote, plus, respCodeStart, respCodeEnd, listStart, listEnd:
		return r, nil
	default:
		reader.UnreadRune()
		return 0, ErrNotSpecialChar
	}
}

// readRespStatusCodeArgs reads a status response code's
// arguments, and returns when it finds the "]" special
// character
func readRespStatusCodeArgs(reader *bufio.Reader) (string, error) {
	args := ""
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return "", err
		}

		if r == respCodeEnd {
			reader.UnreadRune()
			return args, ErrFoundSpecialChar
		}

		args += string(r)
	}
}

func Parse(raw string) *Response {
	resp := NewResponse(raw)
	reader := bufio.NewReader(strings.NewReader(resp.Raw))

	if resp.Raw == "" {
		return resp
	}

	// read the first char with the assumption that
	// it's the star special char (*)
	if sp, err := readSpecialChar(reader); err == nil {
		switch sp {
		case star, plus:
			resp.AddField(string(sp))
			// resp.Type is already false, so no need to set this
			if sp == plus {
				resp.Type = ResponseTypeCommandContinuationReq
			}
		}
	} else if err == ErrNotSpecialChar {
		resp.Tagged = true
	} else {
		log.Panic(err)
	}

	var err error
	for err != io.EOF {
		var atom string
		var sp rune
		// this will read the next atom in the response. If the response is tagged,
		// this would be
		atom, err = readAtom(reader)
		if err == ErrFoundSpecialChar {
			if atom != "" {
				log.Println("Atom: ", atom)
				resp.AddField(atom)
			}

			sp, err = readSpecialChar(reader)
			if err != nil {
				log.Panic(err)
			}

			if sp != space {
				switch sp {
				case listStart:
					log.Println("list start")
					// this is a list, read till end of list
				case respCodeStart:
					log.Println("status response code start")
					// this a status response code, read and store
					// code, then read and store arguments, which
					// will be handled later by the appropriate
					// handler
					// resp.AddField(string(respCodeStart))
					code, err := readAtom(reader)
					if err == ErrFoundSpecialChar {
						resp.AddField(code)

						// read special character and make sure it
						// is a space
						sp, _ = readSpecialChar(reader)
						if sp != space {
							log.Panic("expected a space, found " + "\"" + string(sp) + "\"")
						}

						args, err := readRespStatusCodeArgs(reader)
						if err == ErrFoundSpecialChar {
							resp.AddField(args)

							sp, _ = readSpecialChar(reader)
							if sp != respCodeEnd {
								log.Panic("expected \"]\", found " + "\"" + string(sp) + "\"")
							}
						}
					}
				case cr:
					sp, err = readSpecialChar(reader)
					if err == io.EOF {
						break
					}

					if err != nil {
						log.Panic(err)
					}

					if sp != lf {
						log.Panic("expected \"\\n\", found " + string(sp))
					}
				}
			}
		}
	}

	return resp
}
