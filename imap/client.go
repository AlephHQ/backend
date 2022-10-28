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

	if resp.Fields[2][0] == respCodeStart {
		fields := strings.Split(strings.Trim(resp.Fields[2], "[]"), " ")

		code := StatusResponseCode(fields[0])

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
	if resp.Fields[0] == string(plus) {
		return
	}

	status := StatusResponse(resp.Fields[1])
	switch status {
	case StatusResponseBAD, StatusResponseBYE, StatusResponseNO:
		log.Println(resp.Raw)
		return
	case StatusResponseOK:
		if resp.Fields[2][0] == respCodeStart {
			fields := strings.Trim(resp.Fields[2], "[]")

			code := ""
			reader := strings.NewReader(fields)
			for {
				r, _, err := reader.ReadRune()
				if r == space || err == io.EOF {
					break
				}

				if err != nil {
					log.Panic(err)
				}

				code += string(r)
			}

			switch StatusResponseCode(code) {
			case StatusResponseCodePermanentFlags:
				permflags := make([]string, 0)
				curflag := ""
				for {
					r, _, err := reader.ReadRune()
					if err != nil {
						log.Panic(err)
					}

					if r == listEnd || r == space {
						permflags = append(permflags, curflag)
						curflag = ""

						if r != space {
							break
						}
					} else {
						curflag += string(r)
					}

					if r == listStart {
						continue
					}
				}

				if c.mbox != nil {
					c.mbox.SetPermanentFlags(permflags)
				}

				return
			case StatusResponseCodeUnseen, StatusResponseCodeUIDNext, StatusResponseCodeUIDValidity:
				numstr := ""
				for {
					r, _, err := reader.ReadRune()
					if err == io.EOF {
						break
					}

					if err != nil {
						log.Panic(err)
					}

					numstr += string(r)
				}

				num, err := strconv.ParseUint(numstr, 10, 64)
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
		flags := strings.Split(strings.Trim(resp.Fields[2], "()"), " ")
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

func (c *Client) setState(state ConnectionState) {
	c.slock.Lock()
	c.state = state
	c.slock.Unlock()
}

func (c *Client) read() {
	for {
		respRaw, err := c.readOne()
		if err != nil {
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
		}

		if resp.Fields[2][0] == respCodeStart {
			fields := strings.Split(strings.Trim(resp.Fields[2], "[]"), " ")

			code := StatusResponseCode(fields[0])

			switch code {
			case StatusResponseCodeCapability:
				c.capabilities = make([]string, 0)
				c.capabilities = append(c.capabilities, fields[1:]...)
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
			permissions := StatusResponseCode(strings.Trim(resp.Fields[2], "[]"))
			if c.mbox != nil {
				c.mbox.SetReadOnly(permissions == StatusResponseCodeReadOnly)
			}

		case StatusResponseNO:
			log.Println(resp.Fields[2])
		}

		return nil
	}

	c.mbox = NewMailboxStatus().SetName(name)
	err := c.execute(fmt.Sprintf("select %s", name), handler)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Mailbox() *MailboxStatus {
	return c.mbox
}
