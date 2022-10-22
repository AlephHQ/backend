package imap

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func Run() {
	conn, err := net.Dial("tcp", "modsoussi.com:143")
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()

	fmt.Fprintf(conn, "abcd CAPABILITY")

	resp, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Println(err)
	}

	log.Println(resp)
}
