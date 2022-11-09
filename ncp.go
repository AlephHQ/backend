package main

import (
	"log"

	"ncp/backend/imap/client"
	"ncp/backend/imap/command"
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

	messages, err := c.Fetch(command.FetchMacroAll)
	if err != nil {
		log.Panic(err)
	}

	for _, msg := range messages {
		log.Printf("%v\n", msg.Envelope.Subject)
	}
}
