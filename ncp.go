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
	client.Read()

	err = client.Login("mo@modsoussi.com", "alohomora")
	if err != nil {
		log.Panic(err)
	}
	client.Read()
}
