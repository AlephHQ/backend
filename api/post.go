package api

import (
	"mime"
	"ncp/backend/imap"
	"strings"
)

type Address struct {
	Name    string `json:"name"`
	Mailbox string `json:"mailbox"`
	Host    string `json:"host"`
}

type BodyPart struct {
	Type     string            `json:"type"`
	Subtype  string            `json:"subtype"`
	Encoding string            `json:"encoding"`
	Size     uint64            `json:"size"`
	Params   map[string]string `json:"params"`
	Content  string            `json:"content"`
}

type Body struct {
	Parts []*BodyPart `json:"parts"`
	Full  string      `json:"full"`
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
	Body         *Body              `json:"body"`
	Preview      string             `json:"preview"`
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

		if strings.HasPrefix(msg.Envelope.Subject, "=?") && strings.HasSuffix(msg.Envelope.Subject, "?=") {
			dec := new(mime.WordDecoder)
			subject, _ := dec.Decode(msg.Envelope.Subject)
			post.Subject = subject
		} else {
			post.Subject = msg.Envelope.Subject
		}
	}

	post.InternalDate = msg.InternalDate
	post.Flags = msg.Flags

	if msg.Body != nil {
		post.Body = &Body{
			Parts: make([]*BodyPart, 0),
		}

		for _, p := range msg.Body.Parts {
			part := &BodyPart{
				Type:     p.Type,
				Subtype:  p.Subtype,
				Encoding: p.Encoding,
				Size:     p.Size,
				Params:   p.ParameterList,
			}

			post.Body.Parts = append(
				post.Body.Parts,
				part,
			)
		}
	}

	post.Preview = msg.Preview

	return post
}
