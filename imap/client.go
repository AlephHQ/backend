package imap

import (
	"errors"
	"fmt"
)

var ErrStatusNotOK = errors.New("status not ok")

type Client struct {
	conn *Conn
}

func New(conn *Conn) *Client {
	return &Client{conn}
}

func (c *Client) execute(cmd string) error {
	return c.conn.Writer.WriteString(cmd)
}

func (c *Client) Login(username, password string) error {
	return c.execute(fmt.Sprintf("login %s %s", username, password))
}

func (c *Client) Logout() error {
	return c.execute("logout")
}

func (c *Client) Read() {
	c.conn.Read()
}
