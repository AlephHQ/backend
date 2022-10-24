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

	conn.Read()

	_, err = conn.Writer.WriteString("a01 noop")
	if err != nil {
		log.Panic(err)
	}
	conn.Writer.Flush()
	conn.Read()
}
