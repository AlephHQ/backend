package conn

import (
	"bufio"
	"errors"
	"io"
	"log"
	"ncp/backend/imap"
	"ncp/backend/imap/response"
	"strconv"
)

var ErrParseNotMessage = errors.New("not a message response")
var ErrParse = errors.New("message parse error")
var ErrFoundNIL = errors.New("found NIL")
var ErrNotList = errors.New("not a list")
var ErrNotString = errors.New("not a string")
var ErrNotNumber = errors.New("not a number")

type Reader struct {
	r *bufio.Reader
}

func NewReader(r io.Reader) *Reader {
	reader := &Reader{}
	reader.r = bufio.NewReader(r)

	return reader
}

func (reader *Reader) readAtom() (string, error) {
	atom := ""
	for {
		r, _, err := reader.r.ReadRune()
		if err != nil {
			return "", err
		}

		if imap.IsSpecialChar(r) {
			switch imap.SpecialCharacter(r) {
			case imap.SpecialCharacterOpenBracket, imap.SpecialCharacterCloseBracket:
				if atom == "" || atom[0:4] != "BODY" {
					reader.r.UnreadRune()
					return atom, nil
				}
			default:
				reader.r.UnreadRune()
				return atom, nil
			}
		}

		atom += string(r)
	}
}

func (reader *Reader) readSpecialChar() (rune, error) {
	r, _, err := reader.r.ReadRune()
	if err != nil {
		return 0, err
	}

	if imap.IsSpecialChar(r) {
		return r, nil
	}

	reader.r.UnreadRune()
	return 0, imap.ErrNotSpecialChar
}

func (reader *Reader) readString() (string, error) {
	str := ""

	r, _, err := reader.r.ReadRune()
	if err != nil {
		return str, err
	}

	switch r {
	case rune(imap.SpecialCharacterDoubleQuote):
		for {
			r, _, err := reader.r.ReadRune()
			if err != nil {
				log.Panic(err)
			}

			if r == rune(imap.SpecialCharacterDoubleQuote) {
				break
			}

			str += string(r)
		}

		return str, nil
	case 'N':
		str += string(r)

		for i := 0; i < 2; i++ {
			r, _, err = reader.r.ReadRune()
			if err != nil {
				return "", err
			}

			str += string(r)
		}

		return str, nil
	case rune(imap.SpecialCharacterOpenCurly):
		num, err := reader.readNumber()
		if err != nil && err != imap.ErrFoundSpecialChar {
			return "", ErrParse
		}

		// need to read till CRLF, then read octet data to get the
		// actual string
		var r rune
		for r != rune(imap.SpecialCharacterLF) {
			r, _, err = reader.r.ReadRune()
			if err != nil {
				return "", err
			}
		}

		str := ""
		for i := uint64(0); i < num; i++ {
			r, _, err = reader.r.ReadRune()
			if err != nil {
				return "", err
			}

			str += string(r)
		}

		return str, nil
	default:
		return "", ErrNotString
	}
}

func (reader *Reader) readNumber() (uint64, error) {
	numStr := ""

	for {
		r, _, err := reader.r.ReadRune()
		if err != nil {
			log.Panic(err)
		}

		if r < 48 || r > 57 {
			reader.r.UnreadRune()
			num, err := strconv.ParseUint(numStr, 10, 64)
			if err != nil {
				return 0, err
			}

			return num, imap.ErrFoundSpecialChar
		}

		numStr += string(r)
	}
}

func (reader *Reader) readList() ([]interface{}, error) {
	result := make([]interface{}, 0)

	r, _, err := reader.r.ReadRune()
	if err != nil {
		log.Panic(err)
	}

	var current string
	switch r {
	case 'N':
		// read and return nil
		current += string(r)
		for i := 0; i < 2; i++ {
			r, _, err = reader.r.ReadRune()
			if err != nil {
				return nil, err
			}

			current += string(r)
		}

		if current == "NIL" {
			result = append(result, nil)
			return result, nil
		}

		return nil, ErrParse
	case rune(imap.SpecialCharacterListStart):
		for {
			r, _, err = reader.r.ReadRune()
			if err != nil {
				log.Panic(err)
			}

			if r == rune(imap.SpecialCharacterListStart) {
				reader.r.UnreadRune()
				nested, err := reader.readList()
				if err != nil {
					return nil, err
				}

				result = append(result, nested)
				continue
			}

			if r == rune(imap.SpecialCharacterListEnd) {
				return result, nil
			}

			switch r {
			case rune(imap.SpecialCharacterDoubleQuote), rune(imap.SpecialCharacterOpenCurly):
				reader.r.UnreadRune()
				str, err := reader.readString()
				if err != nil {
					return nil, ErrParse
				}

				result = append(result, str)
			case 'N':
				current = "N"
				for i := 0; i < 2; i++ {
					r, _, err = reader.r.ReadRune()
					if err != nil {
						return nil, err
					}

					current += string(r)
				}

				if current == "NIL" {
					result = append(result, nil)
				} else {
					return nil, ErrParse
				}
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				reader.r.UnreadRune()
				num, err := reader.readNumber()
				if err != nil && err != imap.ErrFoundSpecialChar {
					return nil, ErrParse
				}

				result = append(result, num)
			case rune(imap.SpecialCharacterSpace):
				// do nothing
			default:
				reader.r.UnreadRune()
				atom, err := reader.readAtom()
				if err != nil && err != imap.ErrFoundSpecialChar {
					return nil, ErrParse
				}

				result = append(result, atom)
			}
		}
	default:
		return nil, ErrNotList
	}
}

func (reader *Reader) read() (*response.Response, error) {
	resp := response.NewResponse()

	// read the first char with the assumption that
	// it's the star special char (*)
	if sp, err := reader.readSpecialChar(); err == nil {
		switch sp {
		case rune(imap.SpecialCharacterStar), rune(imap.SpecialCharacterPlus):
			resp.AddField(string(sp))
		}
	} else {
		log.Panic(err)
	}

	var err error
	for err != io.EOF {
		var atom string
		var sp rune
		// this will read the next atom in the response. If the response is tagged,
		// this would be
		atom, err = reader.readAtom()
		if err == nil {
			if atom != "" {
				resp.AddField(atom)
			}

			sp, err = reader.readSpecialChar()
			if err != nil {
				log.Panic(err)
			}

			if sp != rune(imap.SpecialCharacterSpace) {
				switch sp {
				case rune(imap.SpecialCharacterListStart):
					// this is either a list or a regular info string
					// that contains open and close parentheses such
					// as (Ubuntu) or (0.001 + 0.000 s). We must be able
					// to handle both cases.
					//
					// Only cases where this isn't the start of an actual
					// list are when this is an OK, NO, BYE, or BAD status
					// response.
					status := imap.StatusResponse(resp.Fields[1].(string))
					switch status {
					case imap.StatusResponseBAD, imap.StatusResponseNO, imap.StatusResponseOK, imap.StatusResponseBYE, imap.StatusResponsePREAUTH:
						str, err := reader.readAtom()
						if err != nil && err != imap.ErrFoundSpecialChar {
							log.Panic(err)
						}

						resp.AddField("(" + str)
					default:
						reader.r.UnreadRune()

						list, err := reader.readList()
						if err != nil {
							log.Panic(err)
						}

						resp.AddField(list)
					}
				case rune(imap.SpecialCharacterOpenBracket):
					// this a status response code, read and store
					// code and rguments, then pass everything off
					// to be handled later by the appropriate
					// handler
					statusRespCodeFields := make([]interface{}, 0)
					for {
						atom, err := reader.readAtom()
						if err != nil {
							return nil, err
						}

						statusRespCodeFields = append(statusRespCodeFields, atom)
						sp, err := reader.readSpecialChar()
						if err != nil {
							return nil, err
						}

						if sp == rune(imap.SpecialCharacterCloseBracket) {
							resp.AddField(statusRespCodeFields)
							break
						}
					}
				case rune(imap.SpecialCharacterCR):
					sp, err = reader.readSpecialChar()
					if err == io.EOF {
						break
					}

					if err != nil {
						log.Panic(err)
					}

					if sp != rune(imap.SpecialCharacterLF) {
						log.Panic("expected \"\\n\", found " + string(sp))
					}
				}
			}
		}
	}

	return resp, nil
}
