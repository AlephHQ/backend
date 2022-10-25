package imap

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

var ErrStatusNotOK = errors.New("status not ok")

type Client struct {
	conn *Conn

	lock     sync.Mutex
	handlers map[string]HandlerFunc
}

func New(conn *Conn) *Client {
	handlers := make(map[string]HandlerFunc)
	client := &Client{conn: conn, handlers: handlers}

	go client.Read()

	return client
}

func (c *Client) execute(cmd string) error {
	tag := getTag()
	c.registerHandler(tag, func(resp *Response) {
		log.Println(resp.Raw)

		if resp.StatusResp == StatusResponseBYE {
			log.Println("Closing connection ...")
			c.conn.Close()
		}
	})

	return c.conn.Writer.WriteString(tag + " " + cmd)
}

func (c *Client) registerHandler(tag string, f HandlerFunc) {
	c.lock.Lock()
	c.handlers[tag] = f
	c.lock.Unlock()
}

func (c *Client) Login(username, password string) error {
	return c.execute(fmt.Sprintf("login %s %s", username, password))
}

func (c *Client) Logout() error {
	err := c.execute("logout")
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Read() {
	for {
		respRaw := ""
		for {
			r, _, err := c.conn.ReadRune()
			if err != nil {
				log.Panic(err)
			}

			if r == '\n' {
				break
			}

			respRaw = respRaw + string(r)
		}

		if respRaw != "" {
			resp := Parse(respRaw)
			if handler := c.handlers[resp.Tag]; handler != nil {
				handler(resp)
			}
		}
	}
}
