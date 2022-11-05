package client

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"

	"ncp/backend/imap"
	"ncp/backend/imap/conn"
	"ncp/backend/imap/response"
)

type Client struct {
	conn *conn.Conn

	hlock    sync.Mutex
	handlers []response.Handler

	capabilities []string

	state imap.ConnectionState
	slock sync.Mutex

	mbox *imap.MailboxStatus

	updates chan string
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
		fields := strings.Split(resp.Fields[4], " ")

		switch code {
		case imap.StatusResponseCodeCapability:
			c.capabilities = make([]string, 0)
			c.capabilities = append(c.capabilities, fields[1:]...)
		}
	}

	return nil
}

func (c *Client) execute(cmd string, handler response.Handler) error {
	tag := getTag()
	done := make(chan bool)
	handlerFunc := response.NewHandlerFunc(func(resp *response.Response) (bool, error) {
		unregister, err := handler.Handle(resp)
		done <- unregister && err == nil
		return unregister, err
	})

	c.registerHandler(tag, handlerFunc)

	err := c.conn.Writer.WriteCommand(tag + " " + cmd)
	if err != nil {
		log.Panic(err)
	}

	for {
		select {
		case d := <-done:
			if d {
				return nil
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
			switch imap.StatusResponseCode(code) {
			case imap.StatusResponseCodePermanentFlags:
				if c.mbox != nil {
					c.mbox.SetPermanentFlags(strings.Split(strings.Trim(resp.Fields[4], "()"), " "))
				}

				return
			case imap.StatusResponseCodeUnseen, imap.StatusResponseCodeUIDNext, imap.StatusResponseCodeUIDValidity:
				num, err := strconv.ParseUint(resp.Fields[4], 10, 64)
				if err != nil {
					log.Panic(err)
				}

				if c.mbox != nil {
					switch imap.StatusResponseCode(code) {
					case imap.StatusResponseCodeUnseen:
						c.mbox.SetUnseen(num)
					case imap.StatusResponseCodeUIDNext:
						c.mbox.SetUIDNext(num)
					case imap.StatusResponseCodeUIDValidity:
						c.mbox.SetUIDValidity(num)
					}
				}

				return
			}
		}
	}

	// at this point, we have a data response
	code := imap.DataResponseCode(resp.Fields[1])
	switch code {
	case imap.DataResponseCodeFlags:
		flags := strings.Split(resp.Fields[3], " ")
		if c.mbox != nil {
			c.mbox.SetFlags(flags)
		}

		return
	}

	code = imap.DataResponseCode(resp.Fields[2])
	switch code {
	case imap.DataResponseCodeExists, imap.DataResponseCodeRecent:
		num, err := strconv.ParseUint(resp.Fields[1], 10, 64)
		if err != nil {
			log.Panic(err)
		}

		if c.mbox != nil {
			switch code {
			case imap.DataResponseCodeExists:
				c.mbox.SetExists(num)
				return
			case imap.DataResponseCodeRecent:
				c.mbox.SetRecent(num)
				return
			}
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

func (c *Client) Capabilities() []string {
	return c.capabilities
}

func (c *Client) Login(username, password string) error {
	handler := func(resp *response.Response) (bool, error) {
		status := imap.StatusResponse(resp.Fields[1])
		switch status {
		case imap.StatusResponseNO:
			return true, fmt.Errorf("error logging in: %s", resp.Fields[2])
		case imap.StatusResponseOK:
			if resp.Fields[2] == string(imap.SpecialCharacterRespCodeStart) {
				code := imap.StatusResponseCode(resp.Fields[3])
				fields := strings.Split(resp.Fields[4], " ")

				switch code {
				case imap.StatusResponseCodeCapability:
					c.capabilities = make([]string, 0)
					c.capabilities = append(c.capabilities, fields[1:]...)
				}

				return true, nil
			}
		}

		return false, imap.ErrUnhandled
	}

	return c.execute(fmt.Sprintf("login %s %s", username, password), response.NewHandlerFunc(handler))
}

func (c *Client) Close() error {
	if c.state != imap.SelectedState {
		return ErrNotSelectedState
	}

	handler := func(resp *response.Response) (bool, error) {
		status := imap.StatusResponse(resp.Fields[1])
		switch status {
		case imap.StatusResponseOK:
			c.setState(imap.AuthenticatedState)
			c.mbox = nil

			return true, nil
		}

		return false, imap.ErrUnhandled
	}

	return c.execute("close", response.NewHandlerFunc(handler))
}

func (c *Client) Logout() error {
	if c.state == imap.SelectedState {
		err := c.Close()
		if err != nil {
			return err
		}
	}

	handler := func(resp *response.Response) (bool, error) {
		c.setState(imap.LogoutState)
		c.mbox = nil

		return true, nil
	}

	err := c.execute("logout", response.NewHandlerFunc(handler))
	if err != nil {
		return err
	}

	c.conn.Close()

	return nil
}

func (c *Client) Select(name string) error {
	handler := func(resp *response.Response) (bool, error) {
		status := imap.StatusResponse(resp.Fields[1])
		switch status {
		case imap.StatusResponseOK:
			c.setState(imap.SelectedState)

			// set read and write permissions
			statusRespCode := imap.StatusResponseCode(resp.Fields[3])
			switch statusRespCode {
			case imap.StatusResponseCodeReadOnly, imap.StatusResponseCodeReadWrite:
				if c.mbox != nil {
					c.mbox.SetReadOnly(statusRespCode == imap.StatusResponseCodeReadOnly)
				}

				return true, nil
			}

			return false, imap.ErrUnhandled
		case imap.StatusResponseNO:
			log.Println(resp.Fields[3])
			return true, fmt.Errorf("error selecting: %s", resp.Fields[3])
		}

		return false, imap.ErrUnhandled
	}

	c.mbox = imap.NewMailboxStatus().SetName(name)
	return c.execute(fmt.Sprintf("select %s", name), response.NewHandlerFunc(handler))
}

func (c *Client) Mailbox() *imap.MailboxStatus {
	return c.mbox
}

func (c *Client) Fetch() error {
	handler := &response.Fetch{
		Messages: make(chan string),
		Done:     make(chan bool),
	}
	defer close(handler.Done)

	messages := make([]string, 0)
	go func() {
		for {
			msg, more := <-handler.Messages
			if !more {
				log.Println("* NO MORE MESSAGES")
				break
			}

			messages = append(messages, msg)
		}
	}()

	err := c.execute("fetch 1:15 all", handler)
	if err != nil {
		log.Panic(err)
	}

	<-handler.Done
	log.Printf("Received %d messages.\n", len(messages))
	return nil
}
