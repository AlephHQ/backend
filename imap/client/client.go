package client

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"ncp/backend/imap"
	"net"
	"sync"
)

var ErrStatusNotOK = errors.New("status not ok")

type Client struct {
	conn   net.Conn
	reader *bufio.Reader

	State imap.ConnectionState
	lock  sync.Mutex
}

func New() (*Client, error) {
	conn, err := net.Dial("tcp", "modsoussi.com:143")
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)

	// did we get a greeting?
	r, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	resp := imap.NewResponse(r).Parse()
	if resp.StatusResp != imap.StatusResponseOK {
		return nil, ErrStatusNotOK
	}
	log.Println(resp)

	return &Client{conn: conn, State: imap.ConnectedState, reader: reader}, nil
}

func (c *Client) Logout() {
	_, err := fmt.Fprintf(c.conn, "a logout")
	if err != nil {
		log.Panic(err)
	}

	r, err := c.reader.ReadString('\n')
	if err != nil {
		log.Panic(err)
	}
	log.Println(r)

	c.conn.Close()
}
