package conn

import (
	"crypto/tls"
	"net"
)

type Conn struct {
	net.Conn

	*Reader
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
	conn.Reader = NewReader(c)
	conn.Writer = NewWriter(c)
	conn.isTLS = true

	return conn, nil
}
