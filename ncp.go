package main

import (
	"log"
	"ncp/backend/imap/client"
)

func main() {
	client, err := client.New()
	if err != nil {
		log.Panic(err)
	}

	client.Logout()
}
