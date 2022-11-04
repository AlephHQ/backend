package main

import (
	"log"

	"ncp/backend/imap/client"
)

func main() {
	c, err := client.DialWithTLS("tcp", "modsoussi.com:993")
	if err != nil {
		log.Panic(err)
	}

	err = c.Login("mo@modsoussi.com", "alohomora")
	if err != nil {
		log.Panic(err)
	}

	err = c.Select("inbox")
	if err != nil {
		log.Panic(err)
	}
	log.Println(c.Mailbox())

	// err = c.Fetch()
	// if err != nil {
	// 	log.Panic(err)
	// }

	err = c.Logout()
	if err != nil {
		log.Panic(err)
	}
}
