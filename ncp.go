package main

import (
	"log"
	"ncp/backend/imap"
)

func main() {
	client, err := imap.DialWithTLS("tcp", "modsoussi.com:993")
	if err != nil {
		log.Panic(err)
	}
	defer client.Logout()

	err = client.Login("mo@modsoussi.com", "alohomora")
	if err != nil {
		log.Panic(err)
	}

	err = client.Select("inbox")
	if err != nil {
		log.Panic(err)
	}
	log.Println(client.Mailbox())
}
