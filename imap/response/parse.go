package response

import (
	"errors"
	"ncp/backend/imap"
	"strconv"
)

var ErrParseNotMessage = errors.New("not a message response")
var ErrParse = errors.New("message parse error")
var ErrFoundNIL = errors.New("found NIL")
var ErrNotList = errors.New("not a list")
var ErrNotString = errors.New("not a string")
var ErrNotNumber = errors.New("not a number")

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
	seqnum, err := strconv.ParseUint(resp.Fields[1].(string), 10, 64)
	if err != nil {
		return nil, err
	}
	message := imap.NewMessage(seqnum)

	fields := resp.Fields[3].([]interface{})
	i := 0
	for i < len(fields)-1 {
		attribute, ok := fields[i].(string)
		if !ok {
			return nil, ErrParse
		}

		compAttr := imap.NewCompoundMessageAttribute(attribute)
		switch compAttr.Attribute {
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
		case imap.MessageAttributeBody, imap.MessageAttributeBodyPeek:
			body := message.Body
			if body == nil {
				body = imap.NewBody()
			}

			if compAttr.Section == "" { // BODY case
				bodyFields := fields[i+1].([]interface{})
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
			} else {
				// BODY[section] case
				bodySection := fields[i+1].(string)
				body.SetSection(compAttr.Section, bodySection)
			}

			message.SetBody(body)
			i += 2
		case imap.MessageAttributeUID:
			if uid, ok := fields[i+1].(uint64); ok {
				message.SetUID(uid)
			}
			i += 2
		}
	}

	return message, nil
}
