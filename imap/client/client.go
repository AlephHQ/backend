package client

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"ncp/backend/imap"
	"net"
	"sync"
)

var ErrStatusNotOK = errors.New("status not ok")

type Client struct {
	conn   net.Conn
	reader *bufio.Reader

	State imap.ConnectionState
	wg    sync.WaitGroup
	// lock  sync.Mutex
}

func (c *Client) Wait() {
	c.wg.Wait()
}

func (c *Client) Add(delta int) {
	c.wg.Add(delta)
}

func (c *Client) Done() {
	c.wg.Done()
}

func (c *Client) read() {
	for {
		r, err := c.reader.ReadString('\n')
		log.Println(r)
		if err != nil {
			log.Panic(err)
		}

		resp := imap.NewResponse(r).Parse()
		if resp.StatusResp == imap.StatusResponseBYE {
			return
		}

		if resp.Tag == "a01" {
			c.Done()
		}
	}
}

func New() (*Client, error) {
	conn, err := net.Dial("tcp", "modsoussi.com:143")
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)

	// did we get a greeting?
	r, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	resp := imap.NewResponse(r).Parse()
	if resp.StatusResp != imap.StatusResponseOK {
		return nil, ErrStatusNotOK
	}
	reader.Reset(conn)

	client := &Client{conn: conn, State: imap.ConnectedState, reader: reader}
	go client.read()

	return client, nil
}

func (c *Client) Capability() {
	c.Add(1)
	fmt.Fprintf(c.conn, "a01 capability")
}

func (c *Client) Login() {
	c.Add(1)
	fmt.Fprintf(c.conn, "a01 login mo@modsoussi.com alohomora")
}

func (c *Client) Logout() {
	fmt.Fprintf(c.conn, "a logout")

	c.Wait()
	c.conn.Close()
}
