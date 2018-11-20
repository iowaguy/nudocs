package main

import (
	"fmt"
	"net"
	"os"

	"github.com/iowaguy/nudocs/core"
)

const (
	connHost = "localhost"
	connPort = "3333"
	connType = "tcp"
)

func main() {
	// TODO read hosts file
	// TODO use num hosts as arg for NewReducer
	// TODO determine pid id from hostsfile

	// start algorithm
	red := core.NewReducer(5, 0)
	red.Start()

	// Listen for incoming connections.
	l, err := net.Listen(connType, connHost+":"+connPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	fmt.Println("Listening on " + connHost + ":" + connPort)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		if core.IsPeer(conn) {
			go core.ReceivePeerOperations(conn, red)
		} else {
			// there will only be one client, in fact, the client
			// is a singleton to guarantee this
			c := core.NewClient(conn)
			c.Start(red)
		}
	}
}
