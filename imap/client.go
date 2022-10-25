package imap

import (
	"errors"
	"fmt"
	"log"
)

var ErrStatusNotOK = errors.New("status not ok")

type Client struct {
	conn *Conn
}

func New(conn *Conn) *Client {
	return &Client{conn}
}

func (c *Client) execute(cmd string) error {
	return c.conn.Writer.WriteString(getTag() + " " + cmd)
}

func (c *Client) Login(username, password string) error {
	return c.execute(fmt.Sprintf("login %s %s", username, password))
}

func (c *Client) Logout() error {
	err := c.execute("logout")
	if err != nil {
		return err
	}

	c.conn.Close()
	return nil
}

func (c *Client) Read() {
	resp := ""
	for {
		r, _, err := c.conn.ReadRune()
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
