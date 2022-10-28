package main

import (
	"encoding/json"
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

	b, _ := json.Marshal(client.Mailbox())
	log.Println("mailbox", string(b))
}
