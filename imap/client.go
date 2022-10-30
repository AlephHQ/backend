package imap

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
)

type Client struct {
	conn *Conn

	hlock    sync.Mutex
	handlers map[string]HandlerFunc

	capabilities []string

	state ConnectionState
	slock sync.Mutex

	mbox *MailboxStatus
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

	resp := Parse(greeting)
	status := StatusResponse(resp.Fields[1])
	switch status {
	case StatusResponseOK:
		c.setState(NotAuthenticatedState)
	case StatusResponsePREAUTH:
		c.setState(AuthenticatedState)
	case StatusResponseBAD, StatusResponseBYE, StatusResponseNO:
		return ErrStatusNotOK
	}

	if resp.Fields[2] == string(respCodeStart) {
		code := StatusResponseCode(resp.Fields[3])
		fields := strings.Split(resp.Fields[4], " ")

		switch code {
		case StatusResponseCodeCapability:
			c.capabilities = make([]string, 0)
			c.capabilities = append(c.capabilities, fields[1:]...)
		}
	}

	return nil
}

func (c *Client) execute(cmd string, handler HandlerFunc) error {
	tag := getTag()
	done := make(chan error)
	c.registerHandler(tag, func(resp *Response) error {
		err := handler(resp)
		done <- err
		return err
	})

	err := c.conn.Writer.WriteString(tag + " " + cmd)
	if err != nil {
		log.Panic(err)
	}

	return <-done
}

func (c *Client) registerHandler(tag string, f HandlerFunc) {
	c.hlock.Lock()
	c.handlers[tag] = f
	c.hlock.Unlock()
}

func (c *Client) handleUnsolicitedResp(resp *Response) {
	log.Println(resp.Raw)
	if resp.Fields[0] == string(plus) {
		return
	}

	status := StatusResponse(resp.Fields[1])
	switch status {
	case StatusResponseBAD, StatusResponseBYE, StatusResponseNO:
		log.Println(resp.Raw)
		return
	case StatusResponseOK:
		if resp.Fields[2] == string(respCodeStart) {
			code := resp.Fields[3]
			switch StatusResponseCode(code) {
			case StatusResponseCodePermanentFlags:
				if c.mbox != nil {
					c.mbox.SetPermanentFlags(strings.Split(strings.Trim(resp.Fields[4], "()"), " "))
				}

				return
			case StatusResponseCodeUnseen, StatusResponseCodeUIDNext, StatusResponseCodeUIDValidity:
				num, err := strconv.ParseUint(resp.Fields[4], 10, 64)
				if err != nil {
					log.Panic(err)
				}

				if c.mbox != nil {
					switch StatusResponseCode(code) {
					case StatusResponseCodeUnseen:
						c.mbox.SetUnseen(num)
					case StatusResponseCodeUIDNext:
						c.mbox.SetUIDNext(num)
					case StatusResponseCodeUIDValidity:
						c.mbox.SetUIDValidity(num)
					}
				}

				return
			}
		} else {
			log.Println(resp.Fields[2])
		}
	}

	// at this point, we have a data response
	code := DataResponseCode(resp.Fields[1])
	switch code {
	case DataResponseCodeFlags:
		flags := strings.Split(resp.Fields[3], " ")
		if c.mbox != nil {
			c.mbox.SetFlags(flags)
		}

		return
	}

	code = DataResponseCode(resp.Fields[2])
	switch code {
	case DataResponseCodeExists, DataResponseCodeRecent:
		num, err := strconv.ParseUint(resp.Fields[1], 10, 64)
		if err != nil {
			log.Panic(err)
		}

		if c.mbox != nil {
			switch code {
			case DataResponseCodeExists:
				c.mbox.SetExists(num)
				return
			case DataResponseCodeRecent:
				c.mbox.SetRecent(num)
				return
			}
		}
	default:
		log.Println(code)
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
		if r == lf {
			break
		}
	}

	return respRaw, nil
}

func (c *Client) setState(state ConnectionState) {
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
			resp := Parse(respRaw)
			if handler := c.handlers[resp.Fields[0]]; handler != nil {
				handler(resp)
			} else if resp.Fields[0] == string(star) {
				c.handleUnsolicitedResp(resp)
			}
		}
	}
}

// BEGIN EXPORTED

func New(conn *Conn) (*Client, error) {
	handlers := make(map[string]HandlerFunc)
	client := &Client{conn: conn, handlers: handlers}

	err := client.waitForAndHandleGreeting()
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
	handler := func(resp *Response) error {
		log.Println(resp.Raw)

		status := StatusResponse(resp.Fields[1])
		switch status {
		case StatusResponseNO:
			return fmt.Errorf("error logging in: %s", resp.Fields[2])
		case StatusResponseOK:
			if resp.Fields[2] == string(respCodeStart) {
				code := StatusResponseCode(resp.Fields[3])
				fields := strings.Split(resp.Fields[4], " ")

				switch code {
				case StatusResponseCodeCapability:
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
	if c.state != SelectedState {
		return ErrNotSelectedState
	}

	handler := func(resp *Response) error {
		status := StatusResponse(resp.Fields[1])
		switch status {
		case StatusResponseOK:
			c.setState(AuthenticatedState)
			c.mbox = nil
		}

		return nil
	}

	return c.execute("close", handler)
}

func (c *Client) Logout() error {
	if c.state == SelectedState {
		err := c.Close()
		if err != nil {
			return err
		}
	}

	handler := func(resp *Response) error {
		log.Println(resp.Raw)

		c.setState(LogoutState)
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
	handler := func(resp *Response) error {
		status := StatusResponse(resp.Fields[1])
		switch status {
		case StatusResponseOK:
			c.setState(SelectedState)

			// set read and write permissions
			permissions := StatusResponseCode(resp.Fields[3])
			if c.mbox != nil {
				c.mbox.SetReadOnly(permissions == StatusResponseCodeReadOnly)
			}

		case StatusResponseNO:
			log.Println(resp.Fields[3])
		}

		return nil
	}

	c.mbox = NewMailboxStatus().SetName(name)
	return c.execute(fmt.Sprintf("select %s", name), handler)
}

func (c *Client) Mailbox() *MailboxStatus {
	return c.mbox
}

func (c *Client) Fetch() error {
	handler := func(resp *Response) error {
		status := StatusResponse(resp.Fields[1])
		switch status {
		case StatusResponseNO:
			return fmt.Errorf("error fetching: %s", resp.Fields[2])
		case StatusResponseOK:
			log.Println(resp)
		}

		return nil
	}

	return c.execute("fetch 1:15 all", handler)
}
