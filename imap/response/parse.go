package response

import (
	"bufio"
	"errors"
	"io"
	"log"
	"ncp/backend/imap"
	"strconv"
	"strings"
)

var ErrParseNotMessage = errors.New("not a message response")
var ErrParse = errors.New("message parse error")

// readAtom reads until it finds a special character, and when it does
// so, it returns the atom that precedes the special character and an
// error indicating it found a special character
func readAtom(reader io.RuneScanner) (string, error) {
	atom := ""
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return "", err
		}

		if imap.IsSpecialChar(r) {
			reader.UnreadRune()
			return atom, imap.ErrFoundSpecialChar
		}

		atom += string(r)
	}
}

// readSpecialChar reads a single special character
func readSpecialChar(reader io.RuneScanner) (rune, error) {
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

// readRespStatusCodeArgs reads a status response code's
// arguments, and returns when it finds the "]" special
// character
func readRespStatusCodeArgs(reader io.RuneScanner) (string, error) {
	args := ""
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return "", err
		}

		if r == rune(imap.SpecialCharacterRespCodeEnd) {
			reader.UnreadRune()
			return args, imap.ErrFoundSpecialChar
		}

		args += string(r)
	}
}

// readList will read till end of list, and assumes the first "(" has already
// been read
func readList(reader io.RuneScanner) (string, error) {
	list := ""
	nonClosedOpens := 0
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			log.Panic(err)
		}

		if r == rune(imap.SpecialCharacterListStart) {
			nonClosedOpens++
		}

		if r == rune(imap.SpecialCharacterListEnd) {
			if nonClosedOpens == 0 {
				reader.UnreadRune()
				return list, nil
			}

			nonClosedOpens--
		}

		list += string(r)
	}
}

// readString reads an entire string without assuming the first double quotes were read
func readString(reader io.RuneScanner) (string, error) {
	str := ""

	r, _, err := reader.ReadRune()
	if err != nil {
		return str, err
	}

	if r != rune(imap.SpecialCharacterDoubleQuote) {
		return str, ErrParse
	}

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
}

// readNumber reads a number until it finds a non digit rune
func readNumber(reader io.RuneScanner) (uint64, error) {
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

func parseEnvelope(raw string) (*imap.Envelope, error) {
	log.Println("ENVELOPE", raw)
	envelope := imap.NewEnvelope()
	reader := strings.NewReader(raw)

	date, err := readString(reader)
	if err != nil {
		return nil, err
	}
	envelope.SetDate(date)

	sp, err := readSpecialChar(reader)
	if err != nil {
		return nil, err
	}

	if sp != rune(imap.SpecialCharacterSpace) {
		return nil, ErrParse
	}

	subject, err := readString(reader)
	if err != nil {
		return nil, err
	}
	envelope.SetSubject(subject)

	return envelope, nil
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
		case rune(imap.SpecialCharacterStar), rune(imap.SpecialCharacterPlus):
			resp.AddField(string(sp))
			// resp.Tagged is already false, so no need to set this
		}
	} else if err == imap.ErrNotSpecialChar {
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
		if err == imap.ErrFoundSpecialChar {
			if atom != "" {
				resp.AddField(atom)
			}

			sp, err = readSpecialChar(reader)
			if err != nil {
				log.Panic(err)
			}

			if sp != rune(imap.SpecialCharacterSpace) {
				switch sp {
				case rune(imap.SpecialCharacterListStart):
					// this is a list, read till end of list
					resp.AddField(string(imap.SpecialCharacterListStart))
					list, err := readList(reader)
					if err != nil {
						log.Panic(err)
					}

					resp.AddField(list)
					sp, err = readSpecialChar(reader)
					if err != nil {
						log.Panic(err)
					}
					resp.AddField(string(sp))
				case rune(imap.SpecialCharacterRespCodeStart):
					// this a status response code, read and store
					// code, then read and store arguments, which
					// will be handled later by the appropriate
					// handler
					resp.AddField(string(imap.SpecialCharacterRespCodeStart))
					code, err := readAtom(reader)
					if err == imap.ErrFoundSpecialChar {
						resp.AddField(code)

						// read special character and make sure it
						// is a space or "]"
						sp, _ = readSpecialChar(reader)
						if sp == rune(imap.SpecialCharacterSpace) {
							args, err := readRespStatusCodeArgs(reader)
							if err == imap.ErrFoundSpecialChar {
								resp.AddField(args)

								sp, _ = readSpecialChar(reader)
								if sp != rune(imap.SpecialCharacterRespCodeEnd) {
									log.Panic("expected \"]\", found " + "\"" + string(sp) + "\"")
								}
								resp.AddField(string(sp))
							}
						}

						if sp == rune(imap.SpecialCharacterRespCodeEnd) {
							resp.AddField(string(sp))
						}
					}
				case rune(imap.SpecialCharacterCR):
					sp, err = readSpecialChar(reader)
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

	return resp
}

func ParseMessage(resp *Response) (*imap.Message, error) {
	log.Println(resp.Raw)

	uid, err := strconv.ParseUint(resp.Fields[1], 10, 64)
	if err != nil {
		return nil, err
	}
	message := imap.NewMessage(uid)

	reader := strings.NewReader(resp.Fields[4])
	for err != io.EOF {
		var atom string
		var sp rune

		atom, err = readAtom(reader)
		if err == imap.ErrFoundSpecialChar {
			sp, err = readSpecialChar(reader)
			if err != nil {
				return nil, err
			}

			if sp == rune(imap.SpecialCharacterSpace) {
				switch imap.MessageAttribute(atom) {
				case imap.MessageAttributeFlags:
					sp, err = readSpecialChar(reader)
					if err != nil {
						return nil, err
					}

					if sp != rune(imap.SpecialCharacterListStart) {
						return nil, ErrParse
					}

					flagsStr, err := readList(reader)
					if err != nil {
						return nil, err
					}

					flags := strings.Split(flagsStr, " ")
					message.SetFlags(flags)
				case imap.MessageAttributeInternalDate:
					var date string
					date, err = readString(reader)
					if err != nil {
						return nil, err
					}

					message.SetInternalDate(date)
				case imap.MessageAttributeRFC822Size:
					size, err := readNumber(reader)
					if err == imap.ErrFoundSpecialChar {
						message.SetSize(size)
					} else {
						return nil, err
					}
				case imap.MessageAttributeEnvelope:
					sp, err = readSpecialChar(reader)
					if err != nil {
						return nil, err
					}

					if sp != rune(imap.SpecialCharacterListStart) {
						return nil, ErrParse
					}

					envelopeRaw, err := readList(reader)
					if err != nil {
						return nil, err
					}

					envelope, err := parseEnvelope(envelopeRaw)
					if err != nil {
						return nil, err
					}

					message.SetEnvelope(envelope)
				}
			}
		}
	}

	return message, nil
}
