package core

import (
	"fmt"
	"net"
	"strconv"
	"sync"
)

type Client struct {
	conn      net.Conn
	startOnce sync.Once
}

var (
	client     *Client
	clientOnce sync.Once
)

func NewClient(cConn net.Conn) *Client {
	clientOnce.Do(func() {
		client = &Client{conn: cConn}
	})

	return client
}

func (c *Client) Start(red *Reduce) {
	c.startOnce.Do(func() {
		go c.ReceiveClientOperations(red)
		go c.SendClientOperations(red)
	})
}

func (c *Client) SendClientOperations(reducer *Reduce) {
	// when they become ready
	for o := range reducer.Ready() {
		c.conn.Write([]byte(o.String()))
	}
}

func (c *Client) ReceiveClientOperations(reducer *Reduce) {
	defer c.conn.Close()

	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)

	for {
		// Read the incoming connection into the buffer.
		n, err := c.conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			break
		}

		var o Operation
		o.OpType = string(buf[0])
		o.Character = string(buf[1])

		if o.Position, err = strconv.Atoi(string(buf[2:n])); err != nil {
			fmt.Println("Error: could not parse position int", err.Error())
		}

		// send operation to algorithm to be processed
		// this function will handle sending to the rest of the peers
		reducer.ClientPropose(o)
	}
}
