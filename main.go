package main

import (
	"net"
	"os"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/iowaguy/nudocs/core"
)

const (
	connHost = "0.0.0.0"
	connPort = "3333"
	connType = "tcp"
)

func main() {
	var wg sync.WaitGroup

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

	go acceptNewConnections(l, red)

	// TODO connect to other peers

	// block until a go routine returns, which should never happen
	wg.Add(1)
	wg.Wait()
}

func acceptNewConnections(l net.Listener, red *core.Reduce) {
	for {
		// Listen for an incoming connection.
		log.Info("Waiting for client or peer to connect")
		conn, err := l.Accept()
		if err != nil {
			log.Error("Error accepting:", err.Error())
			os.Exit(1)
		}
		log.Info("Connection received, determining who it is...")

		if core.IsPeer(conn) {
			log.Info("Connected to peer")
			go core.ReceivePeerOperations(conn, red)
		} else {
			log.Info("Connected to client")
			// there will only be one client, in fact, the client
			// is a singleton to guarantee this
			c := core.NewClient(conn)
			c.Start(red)
		}
	}
}
