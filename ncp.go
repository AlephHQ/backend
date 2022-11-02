package main

import (
	"log"

	"ncp/backend/imap/client"
)

// func runEmersion() {
// 	log.Println("Connecting to server...")

// 	// Connect to server
// 	c, err := client.DialTLS("mail.example.org:993", nil)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Println("Connected")

// 	// Don't forget to logout
// 	defer c.Logout()

// 	// Login
// 	if err := c.Login("username", "password"); err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Println("Logged in")

// 	// List mailboxes
// 	mailboxes := make(chan *emimap.MailboxInfo, 10)
// 	done := make(chan error, 1)
// 	go func() {
// 		done <- c.List("", "*", mailboxes)
// 	}()

// 	log.Println("Mailboxes:")
// 	for m := range mailboxes {
// 		log.Println("* " + m.Name)
// 	}

// 	if err := <-done; err != nil {
// 		log.Fatal(err)
// 	}

// 	// Select INBOX
// 	mbox, err := c.Select("INBOX", false)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Println("Flags for INBOX:", mbox.Flags)

// 	// Get the last 4 messages
// 	from := uint32(1)
// 	to := mbox.Messages
// 	if mbox.Messages > 3 {
// 		// We're using unsigned integers here, only subtract if the result is > 0
// 		from = mbox.Messages - 3
// 	}
// 	seqset := new(emimap.SeqSet)
// 	seqset.AddRange(from, to)

// 	messages := make(chan *emimap.Message, 10)
// 	done = make(chan error, 1)
// 	go func() {
// 		done <- c.Fetch(seqset, []emimap.FetchItem{emimap.FetchEnvelope}, messages)
// 	}()

// 	log.Println("Last 4 messages:")
// 	for msg := range messages {
// 		log.Println("* " + msg.Envelope.Subject)
// 	}

// 	if err := <-done; err != nil {
// 		log.Fatal(err)
// 	}

// 	log.Println("Done!")
// }

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

	err = c.Fetch()
	if err != nil {
		log.Panic(err)
	}
}
