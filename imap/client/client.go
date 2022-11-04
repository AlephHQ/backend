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
	handlers map[string]response.Handler

	capabilities []string

	state imap.ConnectionState
	slock sync.Mutex

	mbox *imap.MailboxStatus

	updates  chan string
	messages chan string
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

func (c *Client) execute(cmd string, handler response.HandlerFunc) error {
	tag := getTag()
	done := make(chan error)
	handlerFunc := response.NewHandlerFunc(func(resp *response.Response) error {
		err := handler.Handle(resp)
		done <- err
		return err
	})

	c.registerHandler(tag, handlerFunc)

	err := c.conn.Writer.WriteString(tag + " " + cmd)
	if err != nil {
		log.Panic(err)
	}

	return <-done
}

func (c *Client) registerHandler(tag string, h response.Handler) {
	c.hlock.Lock()
	c.handlers[tag] = h
	c.hlock.Unlock()
}

func (c *Client) handleUnsolicitedResp(resp *response.Response) {
	// log.Println(resp.Raw)
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

	// message status response
	msgStatusRespCode := imap.MessageStatusResponseCode(resp.Fields[2])
	switch msgStatusRespCode {
	case imap.MessageStatusResponseCodeFetch:
		// 1. read message uid
		// 2. read and parse message envelope
		// log.Println(resp.Raw)
		if c.messages != nil {
			c.messages <- resp.Raw
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

func (c *Client) read() {
	for {
		respRaw, err := c.readOne()
		if err != nil && err != io.EOF {
			log.Panic(err)
		}

		if respRaw != "" {
			resp := response.Parse(respRaw)
			if handler := c.handlers[resp.Fields[0]]; handler != nil {
				handler.Handle(resp)
			} else if resp.Fields[0] == string(imap.SpecialCharacterStar) {
				c.handleUnsolicitedResp(resp)
			}
		}
	}
}

// BEGIN EXPORTED

func Dial(network, addr string) (*Client, error) {
	handlers := make(map[string]response.Handler)
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
	handlers := make(map[string]response.Handler)
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
	handler := func(resp *response.Response) error {
		log.Println(resp.Raw)

		status := imap.StatusResponse(resp.Fields[1])
		switch status {
		case imap.StatusResponseNO:
			return fmt.Errorf("error logging in: %s", resp.Fields[2])
		case imap.StatusResponseOK:
			if resp.Fields[2] == string(imap.SpecialCharacterRespCodeStart) {
				code := imap.StatusResponseCode(resp.Fields[3])
				fields := strings.Split(resp.Fields[4], " ")

				switch code {
				case imap.StatusResponseCodeCapability:
					c.capabilities = make([]string, 0)
					c.capabilities = append(c.capabilities, fields[1:]...)
				}
			}
		}

		return nil
	}

	return c.execute(fmt.Sprintf("login %s %s", username, password), handler)
}

func (c *Client) Close() error {
	if c.state != imap.SelectedState {
		return ErrNotSelectedState
	}

	handler := func(resp *response.Response) error {
		status := imap.StatusResponse(resp.Fields[1])
		switch status {
		case imap.StatusResponseOK:
			c.setState(imap.AuthenticatedState)
			c.mbox = nil
		}

		return nil
	}

	return c.execute("close", handler)
}

func (c *Client) Logout() error {
	if c.state == imap.SelectedState {
		err := c.Close()
		if err != nil {
			return err
		}
	}

	handler := func(resp *response.Response) error {
		log.Println(resp.Raw)

		c.setState(imap.LogoutState)
		c.mbox = nil

		return nil
	}

	err := c.execute("logout", handler)
	if err != nil {
		return err
	}

	c.conn.Close()

	return nil
}

func (c *Client) Select(name string) error {
	handler := func(resp *response.Response) error {
		status := imap.StatusResponse(resp.Fields[1])
		switch status {
		case imap.StatusResponseOK:
			c.setState(imap.SelectedState)

			// set read and write permissions
			permissions := imap.StatusResponseCode(resp.Fields[3])
			if c.mbox != nil {
				c.mbox.SetReadOnly(permissions == imap.StatusResponseCodeReadOnly)
			}

		case imap.StatusResponseNO:
			log.Println(resp.Fields[3])
		}

		return nil
	}

	c.mbox = imap.NewMailboxStatus().SetName(name)
	return c.execute(fmt.Sprintf("select %s", name), handler)
}

func (c *Client) Mailbox() *imap.MailboxStatus {
	return c.mbox
}

func (c *Client) Fetch() error {
	c.messages = make(chan string)
	done := make(chan bool)

	handler := func(resp *response.Response) error {
		status := imap.StatusResponse(resp.Fields[1])
		switch status {
		case imap.StatusResponseNO:
			return fmt.Errorf("error fetching: %s", resp.Fields[2])
		case imap.StatusResponseOK:
			close(c.messages)
			log.Println(resp)
		}

		return nil
	}

	go func() {
		for {
			msg, more := <-c.messages
			if !more {
				log.Println("* NO MORE MESSAGES")
				done <- true
				break
			}

			log.Println(msg)
		}
	}()

	err := c.execute("fetch 1:15 all", handler)
	if err != nil {
		log.Panic(err)
	}

	<-done
	return nil
}
