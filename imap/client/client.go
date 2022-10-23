package client

import (
	"ncp/backend/imap"
	"net"
	"sync"
)

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

}
