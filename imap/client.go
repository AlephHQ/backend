package imap

import (
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
	wg       sync.WaitGroup
	handlers map[string]HandlerFunc

	capabilities []string

	state ConnectionState
	slock sync.Mutex

	mbox *MailboxStatus
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
	status := StatusResponse(resp.Fields[1])
	switch status {
	case StatusResponseOK:
		c.setState(NotAuthenticatedState)
	case StatusResponsePREAUTH:
		c.setState(AuthenticatedState)
	case StatusResponseBAD, StatusResponseBYE, StatusResponseNO:
		return ErrStatusNotOK
	}

	log.Println("Connected")

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
				if err != nil {
					log.Panic(err)
				}

				if r == space {
					break
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
			if handler := c.handlers[resp.Fields[0]]; handler != nil {
				handler(resp)
			} else if resp.Fields[0] == string(star) {
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

		if StatusResponse(resp.Fields[1]) == StatusResponseOK {
			c.setState(AuthenticatedState)
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

		c.wg.Done()
	}

	return c.execute(fmt.Sprintf("login %s %s", username, password), handler)
}

func (c *Client) Logout() error {
	handler := func(resp *Response) {
		log.Println(resp.Raw)

		c.setState(LogoutState)
		c.mbox = nil

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

func (c *Client) Select(name string) error {
	handler := func(resp *Response) {

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

		c.wg.Done()
	}

	c.mbox = NewMailboxStatus().SetName(name)
	return c.execute(fmt.Sprintf("select %s", name), handler)
}

func (c *Client) Mailbox() *MailboxStatus {
	return c.mbox
}
