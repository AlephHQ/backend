package imap

import (
	"bufio"
	"crypto/tls"
	"log"
	"net"
)

type Conn struct {
	net.Conn

	*bufio.Reader
	*Writer

	isTLS bool
}

func (c *Conn) Read() {
	resp := ""
	for {
		r, _, err := c.ReadRune()
		if err != nil {
			log.Panic(err)
		}

		if r == '\n' {
			break
		}

		resp = resp + string(r)
	}

	log.Println(resp)
}

func Dial(network, addr string) (*Conn, error) {
	c, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	conn := &Conn{}
	conn.Conn = c
	conn.Reader = bufio.NewReader(c)
	conn.Writer = NewWriter(c)

	return conn, nil
}

func DialWithTLS(network, addr string) (*Conn, error) {
	c, err := tls.Dial(network, addr, nil)
	if err != nil {
		return nil, err
	}

	conn := &Conn{}
	conn.Conn = c
	conn.Reader = bufio.NewReader(c)
	conn.Writer = NewWriter(c)
	conn.isTLS = true

	return conn, nil
}
