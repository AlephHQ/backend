package imap

import (
	"bufio"
	"crypto/tls"
	"net"
)

type Conn struct {
	net.Conn

	*bufio.Reader
	*Writer

	isTLS bool
}

func Dial(network, addr string) (*Client, error) {
	c, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	conn := &Conn{}
	conn.Conn = c
	conn.Reader = bufio.NewReader(c)
	conn.Writer = NewWriter(c)

	return New(conn)
}

func DialWithTLS(network, addr string) (*Client, error) {
	c, err := tls.Dial(network, addr, nil)
	if err != nil {
		return nil, err
	}

	conn := &Conn{}
	conn.Conn = c
	conn.Reader = bufio.NewReader(c)
	conn.Writer = NewWriter(c)
	conn.isTLS = true

	return New(conn)
}
