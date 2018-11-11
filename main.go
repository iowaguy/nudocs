package main

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

func generateRandomOperation() operation {
	rand.Seed(time.Now().UTC().UnixNano())
	var o operation

	if rand.Intn(2) == 1 {
		o.opType = "i"
	} else {
		o.opType = "d"
	}

	o.character = string(byte(rand.Intn(26) + 65))
	o.position = rand.Intn(128)

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

		var o operation
		o.opType = string(buf[0])
		o.character = string(buf[1])
		if o.position, err = strconv.Atoi(string(buf[2 : n-1])); err != nil {
			fmt.Println("Error: could not parse position int", err.Error())
		}

		// Send a response back to person contacting us.
		conn.Write([]byte("ok"))
	}
}

type operation struct {
	opType    string
	character string
	position  int
}

func (o *operation) String() string {
	return fmt.Sprintf(o.opType + o.character + strconv.Itoa(o.position) + "\n")
}
