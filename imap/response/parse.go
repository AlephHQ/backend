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
var ErrFoundNIL = errors.New("found NIL")
var ErrNotList = errors.New("not a list")
var ErrNotString = errors.New("not a string")
var ErrNotNumber = errors.New("not a number")

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
func readList(reader io.RuneScanner) ([]interface{}, error) {
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
				nested, err := readList(reader)
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
			case rune(imap.SpecialCharacterDoubleQuote):
				reader.UnreadRune()
				str, err := readString(reader)
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
				num, err := readNumber(reader)
				if err != nil && err != imap.ErrFoundSpecialChar {
					return nil, ErrParse
				}

				result = append(result, num)
			case rune(imap.SpecialCharacterSpace):
				// do nothing
			default:
				reader.UnreadRune()
				atom, err := readAtom(reader)
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

// readString reads an entire string without assuming the first double quotes were read
func readString(reader io.RuneScanner) (string, error) {
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
	default:
		return "", ErrNotString
	}
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
						str, err := readAtom(reader)
						if err != nil && err != imap.ErrFoundSpecialChar {
							log.Panic(err)
						}

						resp.AddField("(" + str)
					default:
						reader.UnreadRune()

						list, err := readList(reader)
						if err != nil {
							log.Panic(err)
						}

						resp.AddField(list)
					}
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

func getAddresses(list interface{}) ([]*imap.Address, error) {
	result := make([]*imap.Address, 0)
	if list == nil {
		return result, nil
	}

	l := list.([]interface{})
	for _, addr := range l {
		if addrFields, ok := addr.([]interface{}); !ok {
			return nil, ErrParse
		} else {
			var name, mailbox, host string

			if name, ok = addrFields[0].(string); !ok {
				name = ""
			}

			if mailbox, ok = addrFields[2].(string); !ok {
				mailbox = ""
			}

			if host, ok = addrFields[3].(string); !ok {
				host = ""
			}

			result = append(result, imap.NewAddress(name, mailbox, host))
		}
	}

	return result, nil
}

func getBodyPart(fields []interface{}) *imap.BodyStructure {
	part := imap.NewBodyStrcuture()

	if fields[0] != nil {
		part.SetType(fields[0].(string))
	}

	if fields[1] != nil {
		part.SetSubtype(fields[1].(string))
	}

	if fields[3] != nil {
		part.SetID(fields[3].(string))
	}

	if fields[4] != nil {
		part.SetDescription(fields[4].(string))
	}

	if fields[5] != nil {
		part.SetEncoding(fields[5].(string))
	}

	if fields[6] != nil {
		part.SetSize(fields[6].(uint64))
	}

	if paramList, ok := fields[2].([]interface{}); ok {
		for i := 0; i < len(paramList)-1; i += 2 {
			part.AddKeyValParam(paramList[i].(string), paramList[i+1].(string))
		}
	}

	return part
}

func ParseMessage(resp *Response) (*imap.Message, error) {
	uid, err := strconv.ParseUint(resp.Fields[1].(string), 10, 64)
	if err != nil {
		return nil, err
	}
	message := imap.NewMessage(uid)

	fields := resp.Fields[3].([]interface{})
	i := 0
	for i < len(fields)-1 {
		attribute, ok := fields[i].(string)
		if !ok {
			return nil, ErrParse
		}

		switch imap.MessageAttribute(attribute) {
		case imap.MessageAttributeFlags:
			flags := fields[i+1].([]interface{})
			for _, f := range flags {
				if flag, ok := f.(string); ok {
					message.SetFlag(flag)
				} else {
					return nil, ErrParse
				}
			}

			i += 2
		case imap.MessageAttributeInternalDate:
			if date, ok := fields[i+1].(string); ok {
				message.SetInternalDate(date)
			} else {
				return nil, ErrParse
			}

			i += 2
		case imap.MessageAttributeRFC822Size:
			if size, ok := fields[i+1].(uint64); ok {
				message.SetSize(size)
			} else {
				return nil, ErrParse
			}

			i += 2
		case imap.MessageAttributeEnvelope:
			envelope := imap.NewEnvelope()
			envelopeFields := fields[i+1].([]interface{})

			if date, ok := envelopeFields[0].(string); ok {
				envelope.SetDate(date)
			}

			if subject, ok := envelopeFields[1].(string); ok {
				envelope.SetSubject(subject)
			} else {
				envelope.SetSubject("(No Subject)")
			}

			from, err := getAddresses(envelopeFields[2])
			if err != nil {
				return nil, err
			}
			envelope.SetFrom(from)

			sender, err := getAddresses(envelopeFields[3])
			if err != nil {
				return nil, err
			}
			envelope.SetSender(sender)

			replyTo, err := getAddresses(envelopeFields[4])
			if err != nil {
				return nil, err
			}
			envelope.SetReplyTo(replyTo)

			to, err := getAddresses(envelopeFields[5])
			if err != nil {
				return nil, err
			}
			envelope.SetTo(to)

			cc, err := getAddresses(envelopeFields[6])
			if err != nil {
				return nil, err
			}
			envelope.SetCC(cc)

			bcc, err := getAddresses(envelopeFields[7])
			if err != nil {
				return nil, err
			}
			envelope.SetBCC(bcc)

			inReplyto, err := getAddresses(envelopeFields[8])
			if err != nil {
				return nil, err
			}
			envelope.SetInReplyTo(inReplyto)

			if messageID, ok := envelopeFields[9].(string); ok {
				envelope.SetMessageID(messageID)
			}

			message.SetEnvelope(envelope)
			i += 2
		case imap.MessageAttributeBody:
			// when dealing with body, the first element is either
			// a string (simple non-multipart message) or a list
			// (multipart message), so we need to handle both cases
			bodyFields := fields[1].([]interface{})
			body := imap.NewBody()
			if _, ok := bodyFields[0].(string); ok {
				part := getBodyPart(bodyFields)
				body.AddPart(part).SetMultipart(false)
			}

			if firstPart, ok := bodyFields[0].([]interface{}); ok {
				body.AddPart(getBodyPart(firstPart)).SetMultipart(true)

				for _, elem := range bodyFields[1:] {
					if partFields, ok := elem.([]interface{}); ok {
						body.AddPart(getBodyPart(partFields))
					}

					if multiSubtype, ok := elem.(string); ok {
						body.SetMultipartSubtype(multiSubtype)
					}
				}
			}

			message.SetBody(body)
			i += 2
		}
	}

	return message, nil
}
