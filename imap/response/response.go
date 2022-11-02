package response

import (
	"bufio"
	"io"
	"log"
	"strings"

	"ncp/backend/imap"
)

type Response struct {
	// Raw contains the original response in its raw format
	Raw string

	// Fields contains all the different fields received
	// in the response
	Fields []string

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

		if imap.IsSpecialChar(r) {
			reader.UnreadRune()
			return atom, imap.ErrFoundSpecialChar
		}

		atom += string(r)
	}
}

// readSpecialChar reads a single special character
func readSpecialChar(reader *bufio.Reader) (rune, error) {
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
func readRespStatusCodeArgs(reader *bufio.Reader) (string, error) {
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

// readList will read till end of list
func readList(reader *bufio.Reader) (string, error) {
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
