package imap

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
)

type Client struct {
	C net.Conn
	L sync.Mutex
}

func (c *Client) ReadLine() (string, error) {
	return bufio.NewReader(c.C).ReadString('\n')
}

func Run() {
	c, err := net.Dial("tcp", "modsoussi.com:143")
	if err != nil {
		log.Panic(err)
	}
	defer c.Close()

	client := Client{C: c}

	resp, err := client.ReadLine()
	if err != nil {
		log.Println(err)
	}

	log.Println(resp)

	client.L.Lock()
	fmt.Fprintf(client.C, "abcd LOGOUT")

	resp, err = client.ReadLine()
	if err != nil {
		log.Println(err)
	}

	log.Println(resp)
	client.L.Unlock()
}
