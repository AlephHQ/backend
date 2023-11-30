package conn

import (
	"aleph/backend/imap/response"
	"crypto/tls"
	"net"
)

type Conn struct {
	net.Conn

	*IMAPReader
	*IMAPWriter

	isTLS bool
}

func New(network, addr string, isTLS bool) (*Conn, error) {
	var c net.Conn
	var err error

	if isTLS {
		c, err = tls.Dial(network, addr, nil)
	} else {
		c, err = net.Dial(network, addr)
	}

	if err != nil {
		return nil, err
	}

	conn := &Conn{}
	conn.Conn = c
	conn.IMAPReader = NewIMAPReader(c)
	conn.IMAPWriter = NewIMAPWriter(c)
	conn.isTLS = isTLS

	return conn, nil
}

func (c *Conn) Read() (*response.Response, error) {
	return c.IMAPReader.read()
}

func (c *Conn) Write(cmd string) error {
	return c.IMAPWriter.writeCommand(cmd)
}
