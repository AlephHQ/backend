package conn

import (
	"bufio"
	"errors"
	"io"
	"log"
	"ncp/backend/imap"
	"strconv"
)

var ErrParseNotMessage = errors.New("not a message response")
var ErrParse = errors.New("message parse error")
var ErrFoundNIL = errors.New("found NIL")
var ErrNotList = errors.New("not a list")
var ErrNotString = errors.New("not a string")
var ErrNotNumber = errors.New("not a number")

type Reader struct {
	*bufio.Reader
}

func NewReader(r io.Reader) *Reader {
	reader := &Reader{}
	reader.Reader = bufio.NewReader(r)

	return reader
}

func (reader *Reader) ReadAtom() (string, error) {
	atom := ""
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return "", err
		}

		if imap.IsSpecialChar(r) {
			switch imap.SpecialCharacter(r) {
			case imap.SpecialCharacterOpenBracket, imap.SpecialCharacterCloseBracket:
				if atom == "" || atom[0:4] != "BODY" {
					reader.UnreadRune()
					return atom, imap.ErrFoundSpecialChar
				}
			default:
				reader.UnreadRune()
				return atom, imap.ErrFoundSpecialChar
			}
		}

		atom += string(r)
	}
}

func (reader *Reader) ReadSpecialChar() (rune, error) {
	r, _, err := reader.ReadRune()
	if err != nil {
		return 0, err
	}

	if imap.IsSpecialChar(r) {
		return r, nil
	}

	reader.UnreadRune()
	return 0, imap.ErrNotSpecialChar
}

func (reader *Reader) ReadString() (string, error) {
	str := ""

	r, _, err := reader.ReadRune()
	if err != nil {
		return str, err
	}

	switch r {
	case rune(imap.SpecialCharacterDoubleQuote):
		for {
			r, _, err := reader.ReadRune()
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
			r, _, err = reader.ReadRune()
			if err != nil {
				return "", err
			}

			str += string(r)
		}

		return str, nil
	case rune(imap.SpecialCharacterOpenCurly):
		num, err := reader.ReadNumber()
		if err != nil && err != imap.ErrFoundSpecialChar {
			return "", ErrParse
		}

		// need to read till CRLF, then read octet data to get the
		// actual string
		var r rune
		for r != rune(imap.SpecialCharacterLF) {
			r, _, err = reader.ReadRune()
			if err != nil {
				return "", err
			}
		}

		str := ""
		for i := uint64(0); i < num; i++ {
			r, _, err = reader.ReadRune()
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

func (reader *Reader) ReadNumber() (uint64, error) {
	numStr := ""

	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			log.Panic(err)
		}

		if r < 48 || r > 57 {
			reader.UnreadRune()
			num, err := strconv.ParseUint(numStr, 10, 64)
			if err != nil {
				return 0, err
			}

			return num, imap.ErrFoundSpecialChar
		}

		numStr += string(r)
	}
}

func (reader *Reader) ReadList() ([]interface{}, error) {
	result := make([]interface{}, 0)

	r, _, err := reader.ReadRune()
	if err != nil {
		log.Panic(err)
	}

	var current string
	switch r {
	case 'N':
		// read and return nil
		current += string(r)
		for i := 0; i < 2; i++ {
			r, _, err = reader.ReadRune()
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
			r, _, err = reader.ReadRune()
			if err != nil {
				log.Panic(err)
			}

			if r == rune(imap.SpecialCharacterListStart) {
				reader.UnreadRune()
				nested, err := reader.ReadList()
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
				reader.UnreadRune()
				str, err := reader.ReadString()
				if err != nil {
					return nil, ErrParse
				}

				result = append(result, str)
			case 'N':
				current = "N"
				for i := 0; i < 2; i++ {
					r, _, err = reader.ReadRune()
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
				reader.UnreadRune()
				num, err := reader.ReadNumber()
				if err != nil && err != imap.ErrFoundSpecialChar {
					return nil, ErrParse
				}

				result = append(result, num)
			case rune(imap.SpecialCharacterSpace):
				// do nothing
			default:
				reader.UnreadRune()
				atom, err := reader.ReadAtom()
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
