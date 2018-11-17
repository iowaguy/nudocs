package core

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	connHost = "localhost"
	connPort = "3333"
	connType = "tcp"
)

func main() {
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

		go receiveOperations(conn)
		go sendOperations(conn)
	}

}

func generateRandomOperation() Operation {
	rand.Seed(time.Now().UTC().UnixNano())
	var o Operation

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

func receiveOperations(conn net.Conn) {
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

		var o Operation
		o.OpType = string(buf[0])
		o.Character = string(buf[1])
		if o.Position, err = strconv.Atoi(string(buf[2:n])); err != nil {
			fmt.Println("Error: could not parse position int", err.Error())
		}
		fmt.Println(o.String())
		// Send a response back to person contacting us.
		conn.Write([]byte("ok\n"))
	}
}
