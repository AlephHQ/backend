package main

import (
	"log"

	"ncp/backend/imap"
	"ncp/backend/imap/client"
)

func main() {
	c, err := client.DialWithTLS("tcp", "modsoussi.com:993")
	if err != nil {
		log.Panic(err)
	}
	defer c.Logout()

	err = c.Login("mo@modsoussi.com", "alohomora")
	if err != nil {
		log.Panic(err)
	}

	err = c.Select("inbox")
	if err != nil {
		log.Panic(err)
	}
	log.Println(c.Mailbox())

	messages, err := c.Fetch(
		imap.NewSeqSet(c.Mailbox().Exists, c.Mailbox().Exists),
		nil,
		imap.FetchMacroFull,
	)
	if err != nil {
		log.Panic(err)
	}

	for _, msg := range messages {
		log.Printf("(%v) %s\n", msg.Envelope.From[0], msg.Envelope.Subject)

		if len(msg.Body.Parts) > 0 {
			msg, err := c.Fetch(
				imap.NewSeqSet(c.Mailbox().Exists, c.Mailbox().Exists),
				[]*imap.DataItem{
					{
						Name:    imap.DataItemNameBody,
						Section: imap.BodySectionHeader,
					},
					{
						Name:    imap.DataItemNameBody,
						Section: imap.BodySectionText,
					},
				},
				"",
			)
			if err != nil {
				log.Panic(err)
			}

			log.Println(msg[0].Body)
		}
	}
}
