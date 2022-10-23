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
	conn net.Conn

	State imap.ConnectionState
	lock  sync.Mutex
}

func New() (*Client, error) {
	conn, err := net.Dial("tcp", "modsoussi.com:143")
	if err != nil {
		return nil, err
	}

	// did we get a greeting?
	r, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return nil, err
	}

	resp := imap.NewResponse(r).Parse()
	if resp.StatusResp != imap.StatusResponseOK {
		return nil, ErrStatusNotOK
	}
	log.Println(resp)

	return &Client{conn: conn, State: imap.ConnectedState}, nil
}

func (c *Client) Logout() {
	fmt.Fprintf(c.conn, "A99 Logout")

	c.conn.Close()
}
