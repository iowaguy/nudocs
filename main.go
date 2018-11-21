package main

import (
	"net"
	"os"

	log "github.com/Sirupsen/logrus"
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
	Formatter := new(log.TextFormatter)
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true
	log.SetFormatter(Formatter)

	log.Info("Starting server")

	red := core.NewReducer(5, 0)
	go red.Start()

	// Listen for incoming connections.
	log.Info("Listen for incoming connections")
	l, err := net.Listen(connType, connHost+":"+connPort)
	if err != nil {
		log.Error("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	log.Info("Listening on " + connHost + ":" + connPort)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			log.Error("Error accepting:", err.Error())
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
