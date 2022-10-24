package main

import (
	"log"
	"ncp/backend/imap"
)

func main() {
	conn, err := imap.DialWithTLS("tcp", "modsoussi.com:993")
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()

	conn.Read()

	_, err = conn.Writer.WriteString("a01 login mo@modsoussi.com alohomora")
	if err != nil {
		log.Panic(err)
	}
	conn.Writer.Flush()
	conn.Read()

	_, err = conn.Writer.WriteString("a02 logout")
	if err != nil {
		log.Panic(err)
	}
	conn.Writer.Flush()
	conn.Read()
}
