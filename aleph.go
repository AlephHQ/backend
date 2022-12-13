package main

import (
	"log"

	"aleph/backend/api/server"
	"aleph/backend/imap"
	"aleph/backend/imap/client"
)

func runIMAP() {
	c, err := client.DialWithTLS("tcp", "modsoussi.com:993")
	if err != nil {
		log.Panic(err)
	}
	defer c.Logout()

	// err = c.Login("4Y4kKNqUH0ZUUw@modsoussi.com", "qEYDj0T4T1_SHhgN")
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
		[]imap.SeqSet{
			&imap.SeqNumber{Val: 1},
			&imap.SeqRange{
				From: 4,
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
			{
				Name: imap.DataItemNamePreview,
			},
		},
		"",
	)
	if err != nil {
		log.Panic(err)
	}

	for _, msg := range messages {
		if len(msg.Envelope.From) > 0 {
			log.Printf("* %d: %v %s %v UID:%d\n", msg.SeqNum, msg.Envelope.From[0], msg.Envelope.Subject, msg.Flags, msg.UID)
		} else {
			log.Printf("No Sender: %s", msg.Envelope.Subject)
		}

		log.Println(msg.Preview)

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

	// err = c.Store(
	// 	[]imap.SeqSet{
	// 		&imap.SeqRange{
	// 			From: 7,
	// 			To:   10,
	// 		},
	// 	},
	// 	imap.DataItemPlusFlag,
	// 	[]imap.Flag{
	// 		imap.Flag(`Physics`),
	// 	},
	// )
	// if err != nil {
	// 	log.Panic(err)
	// }

	// err = c.Expunge()
	// if err != nil {
	// 	log.Panic(err)
	// }
}

func main() {
	if err := server.Serve(&server.Params{Port: "7001"}); err != nil {
		log.Panic(err)
	}
}
