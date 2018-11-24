package client

import (
	"net"
	"os"
	"strconv"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/iowaguy/nudocs/common"
	"github.com/iowaguy/nudocs/core"
)

type Client struct {
	conn      net.Conn
	startOnce sync.Once
}

var (
	client          *Client
	clientOnce      sync.Once
	ClientConnected = make(chan int)
)

func NewClient(cConn net.Conn) *Client {
	clientOnce.Do(func() {
		client = &Client{conn: cConn}
		ClientConnected <- 1
	})

	return client
}

func (c *Client) Start(ot core.OpTransformer) {
	c.startOnce.Do(func() {
		go c.ReceiveClientOperations(ot)
		go c.SendClientOperations(ot)
	})
}

func (c *Client) SendClientOperations(ot core.OpTransformer) {
	// when they become ready
	for o := range ot.Ready() {
		c.conn.Write([]byte(o.String()))
	}
}

func (c *Client) ReceiveClientOperations(ot core.OpTransformer) {
	defer c.conn.Close()

	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)

	for {
		// Read the incoming connection into the buffer.
		n, err := c.conn.Read(buf)
		if err != nil {
			log.Warn("Error reading: ", err.Error())
			os.Exit(1)
		}

		var o common.Operation
		o.OpType = string(buf[0])
		o.Character = string(buf[1])

		// TODO lower the log file of below command after debugging
		log.Warn("op=", string(buf[:n-1]))

		if o.Position, err = strconv.Atoi(string(buf[2 : n-1])); err != nil {
			log.Warn("Error: could not parse position int", err.Error())
			os.Exit(1)
		}

		// send operation to algorithm to be processed
		// this function will handle sending to the rest of the peers
		ot.ClientPropose(&o)
	}
}
