package imap

import (
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
)

var ErrStatusNotOK = errors.New("status not ok")

type Client struct {
	conn *Conn

	hlock    sync.Mutex
	wg       sync.WaitGroup
	handlers map[string]HandlerFunc

	capabilities []string

	state ConnectionState
	slock sync.Mutex
}

func New(conn *Conn) *Client {
	handlers := make(map[string]HandlerFunc)
	client := &Client{conn: conn, handlers: handlers}

	greeting := ""
	var err error
	for greeting == "" {
		greeting, err = client.readOne()
		if err != nil {
			log.Panic(err)
		}
	}

	resp := Parse(greeting)
	if resp.StatusResp != StatusResponseOK {
		log.Panic("not a greeting")
	} else {
		log.Println("Connected")
	}

	if resp.StatusRespCode == StatusResponseCodeCapability && len(resp.Capabilities) > 0 {
		client.capabilities = append(client.capabilities, resp.Capabilities...)
	}

	client.slock.Lock()
	if resp.StatusResp == StatusResponseOK {
		client.state = NotAuthenticatedState
	}

	if resp.StatusResp == StatusResponsePREAUTH {
		client.state = AuthenticatedState
	}
	client.slock.Unlock()

	go client.Read()

	return client
}

func (c *Client) execute(cmd string) error {
	tag := getTag()
	c.registerHandler(tag, func(resp *Response) {
		log.Println(resp.Tag, resp.StatusResp, resp.StatusRespCode, resp.Information)

		c.wg.Done()
	})

	c.wg.Add(1)
	return c.conn.Writer.WriteString(tag + " " + cmd)
}

func (c *Client) registerHandler(tag string, f HandlerFunc) {
	c.hlock.Lock()
	c.handlers[tag] = f
	c.hlock.Unlock()
}

func (c *Client) handleUnsolicitedResp(resp *Response) {
	log.Println(resp.Tag, resp.StatusResp, resp.StatusRespCode, resp.Information)
}

func (c *Client) Login(username, password string) error {
	return c.execute(fmt.Sprintf("login %s %s", username, password))
}

func (c *Client) Logout() error {
	err := c.execute("logout")
	if err != nil {
		return err
	}

	c.wg.Wait()
	c.conn.Close()
	return nil
}

func (c *Client) Read() {
	for {
		respRaw, err := c.readOne()
		if err != nil {
			log.Panic(err)
		}

		if respRaw != "" {
			resp := Parse(respRaw)
			if handler := c.handlers[resp.Tag]; handler != nil {
				handler(resp)
			} else {
				c.handleUnsolicitedResp(resp)
			}
		}
	}
}

func (c *Client) readOne() (string, error) {
	respRaw := ""
	for {
		r, _, err := c.conn.ReadRune()
		if err == io.EOF || r == lf {
			break
		}

		if err != nil {
			return "", err
		}

		respRaw += string(r)
	}

	return respRaw, nil
}
