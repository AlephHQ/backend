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

	err = c.Select("INBOX")
	if err != nil {
		log.Panic(err)
	}
	log.Println(c.Mailbox())

	messages, err := c.Fetch(
		&imap.SeqSet{
			{
				From: 1,
				To:   c.Mailbox().Exists,
			},
		},
		[]*imap.DataItem{
			{
				Name: imap.DataItemNameEnvelope,
			},
			{
				Name: imap.DataItemNameUID,
			},
			{
				Name: imap.DataItemNameFlags,
			},
		},
		"",
	)
	if err != nil {
		log.Panic(err)
	}

	for _, msg := range messages {
		if len(msg.Envelope.From) > 0 {
			log.Printf("* %d: %v %s %v %d\n", msg.SeqNum, msg.Envelope.From[0], msg.Envelope.Subject, msg.Flags, msg.UID)
		} else {
			log.Printf("No Sender: %s", msg.Envelope.Subject)
		}

		// if len(msg.Body.Parts) > 0 {
		// 	msg, err := c.Fetch(
		// 		&imap.SeqSet{
		// 			{
		// 				From: msg.UID,
		// 				To:   msg.UID,
		// 			},
		// 		},
		// 		[]*imap.DataItem{
		// 			{
		// 				Name:    imap.DataItemNameBody,
		// 				Section: imap.BodySectionHeader,
		// 			},
		// 			{
		// 				Name:    imap.DataItemNameBody,
		// 				Section: imap.BodySectionText,
		// 			},
		// 		},
		// 		"",
		// 	)
		// 	if err != nil {
		// 		log.Panic(err)
		// 	}

		// 	log.Println(msg[0].Body.Sections[string(imap.BodySectionText)])
		// }
	}

	results, err := c.Search(
		[]*imap.SearchItem{
			{
				Key: imap.SearchKeyText,
				Val: "astro",
			},
		},
	)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Results: %v\n", results)
}
