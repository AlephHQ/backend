package client

import (
	"io"
	"log"
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

// BEGIN UNEXPORTED

func (c *Client) waitForAndHandleGreeting() error {
	var err error
	var resp *response.Response
	for resp == nil {
		resp, err = c.conn.Read()
		if err != nil {
			log.Panic(err)
		}
	}

	status := imap.StatusResponse(resp.Fields[1].(string))
	switch status {
	case imap.StatusResponseOK:
		c.setState(imap.NotAuthenticatedState)
	case imap.StatusResponsePREAUTH:
		c.setState(imap.AuthenticatedState)
	case imap.StatusResponseBAD, imap.StatusResponseBYE, imap.StatusResponseNO:
		return imap.ErrStatusNotOK
	}

	if statusRespCode, ok := resp.Fields[2].([]interface{}); ok {
		code := statusRespCode[0].(string)

		if code == string(imap.StatusResponseCodeCapability) && len(statusRespCode) > 1 {
			c.capabilities = make(map[string]bool)
			for _, arg := range statusRespCode[1:] {
				c.capabilities[arg.(string)] = true
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

	err := c.conn.Write(cmd)
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

	status := imap.StatusResponse(resp.Fields[1].(string))
	switch status {
	case imap.StatusResponseBAD, imap.StatusResponseBYE, imap.StatusResponseNO:
		log.Println(resp.Raw)
		return
	case imap.StatusResponseOK:
		if resp.Fields[2] == string(imap.SpecialCharacterOpenBracket) {
			code := resp.Fields[3]
			log.Printf("*** Unsolicited - %s\n", code)
		}
	}
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
		resp, err := c.conn.Read()
		if err != nil && err != io.EOF {
			log.Println(err)
		}

		if resp != nil {
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
	if err == nil {
		if len(handler.Capabilities) > 0 {
			for _, cap := range handler.Capabilities {
				c.capabilities[cap] = true
			}
		}

		c.setState(imap.AuthenticatedState)
	}

	return err
}

func (c *Client) Close() error {
	if c.state != imap.SelectedState {
		return imap.ErrNotSelected
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
	if c.state != imap.AuthenticatedState {
		return imap.ErrNotAuthenticated
	}

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

func (c *Client) Fetch(seqset []imap.SeqSet, items []*imap.DataItem, m imap.FetchMacro) ([]*imap.Message, error) {
	if c.state != imap.SelectedState {
		return nil, imap.ErrNotSelected
	}

	cmd := command.NewCmdFetch(seqset)
	if len(items) == 0 && string(m) == "" {
		return nil, imap.ErrBadFetchMissingParams
	}

	if len(items) > 0 {
		for _, item := range items {
			cmd.AppendDataItem(item)
		}
	} else {
		cmd.SetMacro(m)
	}

	handler := response.NewHandlerFetch(cmd.Tag)
	defer close(handler.Done)

	err := c.execute(cmd.Command(), handler)
	if err != nil {
		return nil, err
	}

	<-handler.Done
	log.Printf("Received %d messages.\n", len(handler.Messages))
	return handler.Messages, nil
}

func (c *Client) Search(items []*imap.SearchItem) ([]uint64, error) {
	if c.state != imap.SelectedState {
		return nil, imap.ErrNotSelected
	}

	cmd := command.NewCmdSearch()
	if len(items) > 0 {
		for _, item := range items {
			cmd.AddSearchItem(item)
		}
	}

	handler := response.NewHandlerSearch(cmd.Tag)
	err := c.execute(cmd.Command(), handler)
	if err != nil {
		return nil, err
	}

	return handler.Results, nil
}

func (c *Client) Expunge() error {
	if c.state != imap.SelectedState {
		return imap.ErrNotSelected
	}

	cmd := command.NewCmdExpunge()
	handler := response.NewHandlerExpunge(cmd.Tag)

	return c.execute(cmd.Command(), handler)
}

func (c *Client) Store(seqset *imap.SeqSet, name imap.DataItemName, values []imap.Flag) error {
	if c.state != imap.SelectedState {
		return imap.ErrNotSelected
	}

	cmd := command.NewCmdStore(seqset, name, values)
	handler := response.NewHandlerStore(cmd.Tag)

	return c.execute(cmd.Command(), handler)
}
