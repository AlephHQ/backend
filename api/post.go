package api

import (
	"ncp/backend/imap"
)

type Address struct {
	Name    string `json:"name"`
	Mailbox string `json:"mailbox"`
	Host    string `json:"host"`
}

type BodyPart struct {
	Type     string `json:"type"`
	Subtype  string `json:"subtype"`
	Encoding string `json:"encoding"`
	Size     uint64 `json:"size"`
}

type Body struct {
	Parts []*BodyPart `json:"parts"`
	Text  string      `json:"text"`
}

type Post struct {
	UID       uint64     `json:"uid"`
	SeqNum    uint64     `json:"seq_num"`
	From      []*Address `json:"from"`
	Sender    []*Address `json:"sender"`
	To        []*Address `json:"to"`
	MessageID string     `json:"message_id"`
	Subject   string     `json:"subject"`
	Body      *Body      `json:"body"`
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
	}

	if msg.Body != nil {
		post.Body = &Body{
			Parts: make([]*BodyPart, 0),
		}

		for _, part := range msg.Body.Parts {
			post.Body.Parts = append(
				post.Body.Parts,
				&BodyPart{
					Type:     part.Type,
					Subtype:  part.Subtype,
					Encoding: part.Encoding,
					Size:     part.Size,
				},
			)
		}

		post.Body.Text = msg.Body.Sections[string(imap.BodySectionText)]
	}

	return post
}
