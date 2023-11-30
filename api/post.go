package api

import (
	"aleph/backend/imap"
	"io"
	"mime"
	"mime/quotedprintable"
	"strconv"
	"strings"
)

type Address struct {
	Name    string `json:"name"`
	Mailbox string `json:"mailbox"`
	Host    string `json:"host"`
}

type Body struct {
	Type     string            `json:"type"`
	Subtype  string            `json:"subtype"`
	Encoding string            `json:"encoding"`
	Size     uint64            `json:"size"`
	Params   map[string]string `json:"params"`
	Content  string            `json:"content"`
}

type Post struct {
	UID          uint64             `json:"uid"`
	SeqNum       uint64             `json:"seq_num"`
	InternalDate string             `json:"internal_date"`
	From         []*Address         `json:"from"`
	Sender       []*Address         `json:"sender"`
	To           []*Address         `json:"to"`
	MessageID    string             `json:"message_id"`
	Subject      string             `json:"subject"`
	Flags        map[imap.Flag]bool `json:"flags"`
	Body         *Body              `json:"body,omitempty"`
	Preview      string             `json:"preview,omitempty"`
}

func MessageToPost(msg *imap.Message) *Post {
	post := &Post{
		UID:    msg.UID,
		SeqNum: msg.SeqNum,
		From:   make([]*Address, 0),
		Sender: make([]*Address, 0),
		To:     make([]*Address, 0),
	}

	if msg.Envelope != nil {
		for _, addr := range msg.Envelope.From {
			post.From = append(
				post.From,
				&Address{
					Name:    addr.Name,
					Mailbox: addr.Mailbox,
					Host:    addr.Host,
				},
			)
		}

		for _, addr := range msg.Envelope.Sender {
			post.Sender = append(
				post.Sender,
				&Address{
					Name:    addr.Name,
					Mailbox: addr.Mailbox,
					Host:    addr.Host,
				},
			)
		}

		for _, addr := range msg.Envelope.To {
			post.To = append(
				post.To,
				&Address{
					Name:    addr.Name,
					Mailbox: addr.Mailbox,
					Host:    addr.Host,
				},
			)
		}

		post.MessageID = msg.Envelope.MessageID

		post.Subject = msg.Envelope.Subject
		if strings.Contains(msg.Envelope.Subject, "=?") && strings.Contains(msg.Envelope.Subject, "?=") {
			// need to decode part of or the entire subject
			startIndex := strings.Index(msg.Envelope.Subject, "=?")
			endIndex := strings.Index(msg.Envelope.Subject, "?=") + 1
			encoded := msg.Envelope.Subject[startIndex : endIndex+1]
			dec := new(mime.WordDecoder)
			decoded, _ := dec.Decode(encoded)
			post.Subject = strings.Replace(msg.Envelope.Subject, encoded, decoded, -1)
		}
	}

	post.InternalDate = msg.InternalDate
	post.Flags = msg.Flags

	if msg.Body != nil {
		for i, p := range msg.Body.Parts {
			section := strconv.Itoa(i + 1)
			if msg.Body.Sections[section] != "" {
				post.Body = &Body{
					Type:     p.Type,
					Subtype:  p.Subtype,
					Encoding: p.Encoding,
					Size:     p.Size,
					Params:   p.ParameterList,
					Content:  msg.Body.Sections[section],
				}

				switch imap.Encoding(p.Encoding) {
				case imap.EncodingQuotePrintable:
					b, err := io.ReadAll(quotedprintable.NewReader(strings.NewReader(msg.Body.Sections[section])))
					if err != nil {
						panic(err)
					}

					post.Body.Content = string(b)
				}

				if p.Type == "text" && p.Subtype == "plain" {
					// let's use this to set the post's Preview field
					content := post.Body.Content
					if len(content) > 512 {
						content = content[0:512]
					}

					lines := strings.Split(content, "\r\n")
					for _, line := range lines {
						if strings.Contains(line, "http") {
							continue
						}

						post.Preview = line
						break
					}
				}
			}
		}
	}

	return post
}
