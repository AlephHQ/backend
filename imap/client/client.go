package client

import (
	"errors"
	"io"
	"log"
	"strings"
	"sync"

	"ncp/backend/imap"
	"ncp/backend/imap/command"
	"ncp/backend/imap/conn"
	"ncp/backend/imap/response"
)

type Client struct {
	conn *conn.Conn

	hlock    sync.Mutex
	handlers []response.Handler

	capabilities map[string]bool

	state imap.ConnectionState
	slock sync.Mutex

	mbox *imap.Mailbox

	updates chan string
}

type Handled struct {
	Unregister bool
	Err        error
}

var ErrNotSelectedState = errors.New("not in selected state")

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

	resp := response.Parse(greeting)
	status := imap.StatusResponse(resp.Fields[1])
	switch status {
	case imap.StatusResponseOK:
		c.setState(imap.NotAuthenticatedState)
	case imap.StatusResponsePREAUTH:
		c.setState(imap.AuthenticatedState)
	case imap.StatusResponseBAD, imap.StatusResponseBYE, imap.StatusResponseNO:
		return imap.ErrStatusNotOK
	}

	if resp.Fields[2] == string(imap.SpecialCharacterRespCodeStart) {
		code := imap.StatusResponseCode(resp.Fields[3])
		fields := strings.Split(resp.Fields[5], " ")

		switch code {
		case imap.StatusResponseCodeCapability:
			c.capabilities = make(map[string]bool)
			for _, cap := range fields[1:] {
				c.capabilities[cap] = true
			}
		}
	}

	return nil
}

func (c *Client) execute(cmd string, handler response.Handler) error {
	tag := getTag()
	done := make(chan Handled)
	handlerFunc := response.NewHandlerFunc(func(resp *response.Response) (bool, error) {
		unregister, err := handler.Handle(resp)
		done <- Handled{Unregister: unregister, Err: err}
		return unregister, err
	})

	c.registerHandler(tag, handlerFunc)

	err := c.conn.Writer.WriteCommand(cmd)
	if err != nil {
		log.Panic(err)
	}

	for {
		select {
		case d := <-done:
			if d.Unregister && d.Err != imap.ErrUnhandled {
				return d.Err
			}
		}
	}
}

func (c *Client) registerHandler(tag string, h response.Handler) {
	c.hlock.Lock()
	c.handlers = append(c.handlers, h)
	c.hlock.Unlock()
}

func (c *Client) handleUnsolicitedResp(resp *response.Response) {
	if resp.Fields[0] == string(imap.SpecialCharacterPlus) {
		return
	}

	status := imap.StatusResponse(resp.Fields[1])
	switch status {
	case imap.StatusResponseBAD, imap.StatusResponseBYE, imap.StatusResponseNO:
		log.Println(resp.Raw)
		return
	case imap.StatusResponseOK:
		if resp.Fields[2] == string(imap.SpecialCharacterRespCodeStart) {
			code := resp.Fields[3]
			log.Printf("*** Unsolicited - %s\n", code)
		}
	}
}

func (c *Client) readOne() (string, error) {
	respRaw := ""
	for {
		r, _, err := c.conn.ReadRune()
		if err != nil {
			return "", err
		}

		respRaw += string(r)
		if r == rune(imap.SpecialCharacterLF) {
			break
		}
	}

	return respRaw, nil
}

func (c *Client) setState(state imap.ConnectionState) {
	c.slock.Lock()
	c.state = state
	c.slock.Unlock()
}

func (c *Client) handle(resp *response.Response) error {
	c.hlock.Lock()
	defer c.hlock.Unlock()

	lastHandlerIndex := len(c.handlers) - 1
	for i := lastHandlerIndex; i >= 0; i-- {
		unregister, err := c.handlers[i].Handle(resp)
		if unregister {
			c.handlers = append(c.handlers[:i], c.handlers[i+1:]...)

			return err
		}

		if err != imap.ErrUnhandled {
			return err
		}
	}

	return imap.ErrUnhandled
}

func (c *Client) read() {
	for {
		respRaw, err := c.readOne()
		if err != nil && err != io.EOF {
			log.Println(err)
		}

		if respRaw != "" {
			resp := response.Parse(respRaw)
			if err := c.handle(resp); err == imap.ErrUnhandled {
				c.handleUnsolicitedResp(resp)
			}
		}
	}
}

// BEGIN EXPORTED

func Dial(network, addr string) (*Client, error) {
	handlers := make([]response.Handler, 0)
	updates := make(chan string)
	conn, err := conn.New(network, addr, false)
	if err != nil {
		return nil, err
	}

	client := &Client{
		conn:     conn,
		handlers: handlers,
		updates:  updates,
	}

	err = client.waitForAndHandleGreeting()
	if err != nil {
		return nil, err
	}

	go client.read()

	return client, nil
}

func DialWithTLS(network, addr string) (*Client, error) {
	handlers := make([]response.Handler, 0)
	updates := make(chan string)
	conn, err := conn.New(network, addr, true)
	if err != nil {
		return nil, err
	}

	client := &Client{
		conn:     conn,
		handlers: handlers,
		updates:  updates,
	}

	err = client.waitForAndHandleGreeting()
	if err != nil {
		return nil, err
	}

	go client.read()

	return client, nil
}

func (c *Client) Capability(cap string) bool {
	return c.capabilities[cap]
}

func (c *Client) Login(username, password string) error {
	cmd := command.NewCmdLogin(username, password)
	handler := response.NewHandlerLogin(cmd.Tag)

	err := c.execute(cmd.Command(), handler)
	if err == nil && len(handler.Capabilities) > 0 {
		for _, cap := range handler.Capabilities {
			c.capabilities[cap] = true
		}
	}

	return err
}

func (c *Client) Close() error {
	if c.state != imap.SelectedState {
		return ErrNotSelectedState
	}

	cmd := command.NewCmdClose()
	handler := response.NewHandlerClose(cmd.Tag)

	err := c.execute(cmd.Command(), handler)
	if err == nil {
		c.setState(imap.AuthenticatedState)
	}

	return err
}

func (c *Client) Logout() error {
	if c.state == imap.SelectedState {
		err := c.Close()
		if err != nil {
			return err
		}
	}

	cmd := command.NewCmdLogout()
	handler := response.NewHandlerLogout(cmd.Tag)

	err := c.execute(cmd.Command(), handler)
	if err == nil {
		c.setState(imap.LogoutState)
	}

	c.conn.Close()
	return err
}

func (c *Client) Select(name string) error {
	cmd := command.NewCmdSelect(name)
	handler := response.NewHandlerSelect(name, cmd.Tag)

	err := c.execute(cmd.Command(), handler)
	if err == nil {
		c.mbox = handler.Mailbox
		c.setState(imap.SelectedState)
	}

	return err
}

func (c *Client) Mailbox() *imap.Mailbox {
	return c.mbox
}

func (c *Client) Fetch() error {
	cmd := command.NewCmdFetch()
	handler := response.NewHandlerFetch(cmd.Tag)
	defer close(handler.Done)

	err := c.execute(cmd.Command(), handler)
	if err != nil {
		log.Panic(err)
	}

	<-handler.Done
	log.Printf("Received %d messages.\n", len(handler.Messages))
	return nil
}
