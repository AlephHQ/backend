package conn

import (
	"crypto/tls"
	"ncp/backend/imap/response"
	"net"
)

type Conn struct {
	net.Conn

	*IMAPReader
	*Writer

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
	conn.Writer = NewWriter(c)
	conn.isTLS = isTLS

	return conn, nil
}

func (c *Conn) ReadResponse() (*response.Response, error) {
	return c.IMAPReader.read()
}
