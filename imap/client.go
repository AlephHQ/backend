package imap

import (
	"fmt"
	"io"
	"log"
	"sync"
)

type Client struct {
	conn *Conn

	hlock    sync.Mutex
	wg       sync.WaitGroup
	handlers map[string]HandlerFunc

	capabilities []string

	state ConnectionState
	slock sync.Mutex
}

// BEGIN UNEXPORTED

func (c *Client) waitForAndHandleGreeting() error {
	greeting := ""
	var err error
	for greeting == "" {
		greeting, err = c.readOne()
		if err != nil {
			log.Panic(err)
		}
	}

	resp := Parse(greeting)
	if resp.StatusResp != StatusResponseOK {
		return ErrStatusNotOK
	} else {
		log.Println("Connected")
	}

	if resp.StatusRespCode == StatusResponseCodeCapability && len(resp.Capabilities) > 0 {
		c.capabilities = append(c.capabilities, resp.Capabilities...)
	}

	c.slock.Lock()
	if resp.StatusResp == StatusResponseOK {
		c.state = NotAuthenticatedState
	}

	if resp.StatusResp == StatusResponsePREAUTH {
		c.state = AuthenticatedState
	}
	c.slock.Unlock()

	return nil
}

func (c *Client) execute(cmd string, handler HandlerFunc) error {
	tag := getTag()
	c.registerHandler(tag, handler)
	c.wg.Add(1)

	return c.conn.Writer.WriteString(tag + " " + cmd)
}

func (c *Client) registerHandler(tag string, f HandlerFunc) {
	c.hlock.Lock()
	c.handlers[tag] = f
	c.hlock.Unlock()
}

func (c *Client) handleUnsolicitedResp(resp *Response) {
	log.Println(resp.Raw)
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

func (c *Client) setState(state ConnectionState) {
	c.slock.Lock()
	c.state = state
	c.slock.Unlock()
}

// BEGIN EXPORTED

func New(conn *Conn) *Client {
	handlers := make(map[string]HandlerFunc)
	client := &Client{conn: conn, handlers: handlers}

	err := client.waitForAndHandleGreeting()
	if err != nil {
		log.Panic(err)
	}

	go client.Read()

	return client
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

func (c *Client) Capabilities() []string {
	return c.capabilities
}

func (c *Client) Login(username, password string) error {
	handler := func(resp *Response) {
		log.Println(resp.Raw)

		if resp.StatusRespCode == StatusResponseCodeCapability && len(resp.Capabilities) > 0 {
			c.capabilities = make([]string, 0)
			c.capabilities = append(c.capabilities, resp.Capabilities...)
		}

		c.wg.Done()
	}

	return c.execute(fmt.Sprintf("login %s %s", username, password), handler)
}

func (c *Client) Logout() error {
	handler := func(resp *Response) {
		log.Println(resp.Raw)

		c.setState(LogoutState)

		c.wg.Done()
	}

	err := c.execute("logout", handler)
	if err != nil {
		return err
	}

	c.wg.Wait()
	c.conn.Close()

	return nil
}

func (c *Client) Select() error {
	handler := func(resp *Response) {
		log.Println(resp.Raw)

		if resp.StatusResp == StatusResponseOK {
			c.setState(SelectedState)
		}

		c.wg.Done()
	}

	return c.execute("select inbox", handler)
}
