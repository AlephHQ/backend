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
		imap.NewSeqSet(c.Mailbox().Exists-4, c.Mailbox().Exists),
		imap.FetchMacroAll,
	)
	if err != nil {
		log.Panic(err)
	}

	for _, msg := range messages {
		log.Printf("(%v) %s\n", msg.Envelope.From, msg.Envelope.Subject)
	}
}
