package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

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

		if IsPeer(conn) {
			go receivePeerOperations(conn, red)
		}

		go sendOperations(conn)
	}

}

func receivePeerOperations(conn net.Conn, reducer *core.Reduce) {
	defer conn.Close()

	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)

	for {
		// Read the incoming connection into the buffer.
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			break
		}

		sBuf := string(buf)
		var o core.PeerOperation
		o.OpType = string(buf[0])
		o.Character = string(buf[1])
		vcStart := strings.Index(sBuf[2:], " ") + 1
		if vcStart <= 0 {
			fmt.Println("Error parsing peer operation")
			return
		}

		if o.Position, err = strconv.Atoi(string(buf[2 : vcStart-1])); err != nil {
			fmt.Println("Error: could not parse position int", err.Error())
		}

		if vClock, err := core.ParseVectorClock(string(buf[vcStart:n])); err != nil {
			fmt.Println("Error: could not parse vector clock", err.Error())
		} else {
			o.VClock = *vClock
		}

		// send operation to algorithm to be processed
		reducer.PeerPropose(o)

		// fmt.Println(o.String())
	}
}

func IsPeer(conn net.Conn) bool {
	buf := make([]byte, 256)

	// Read the incoming connection into the buffer.
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		os.Exit(1)
	}

	return string(buf[:n]) == "peer"
}

func generateRandomOperation() core.Operation {
	rand.Seed(time.Now().UTC().UnixNano())
	var o core.Operation

	if rand.Intn(2) == 1 {
		o.OpType = "i"
	} else {
		o.OpType = "d"
	}

	o.Character = string(byte(rand.Intn(26) + 65))
	o.Position = rand.Intn(128)

	return o
}

func sendOperations(conn net.Conn) {
	defer conn.Close()

	for {
		o := generateRandomOperation()

		conn.Write([]byte(o.String()))
		time.Sleep(2 * time.Second)
	}
}
