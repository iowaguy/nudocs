package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/iowaguy/nudocs/core"
)

// TODO later, send document hash to make sure initial document is the same

var (
	doc       string
	localOps  chan *core.Operation
	serverOps chan *core.Operation
	port      = flag.Int("p", 3333, "Server port to connect to")
	host      = flag.String("h", "localhost", "Server hostname to connect to")
	file      = flag.String("f", "", "Path to shared file")
)

func init() {
	Formatter := new(log.TextFormatter)
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true
	log.SetFormatter(Formatter)
}

func main() {
	flag.Parse()

	log.Info("Host specified=" + *host + "; Port specified=" + strconv.Itoa(*port))
	localOps = make(chan *core.Operation, 100)

	// channel is unbuffered, only supports one at a time
	serverOps = make(chan *core.Operation)

	var conn net.Conn
	for {
		var err error
		conn, err = net.Dial("tcp", *host+":"+strconv.Itoa(*port))
		if err != nil {
			log.Warn("Could not connect. Trying again. Error: " + err.Error())
			time.Sleep(500 * time.Millisecond)
		} else {
			log.Info("Client connected to server")
			break
		}
	}
	sendToServer(conn, "client")

	// read document
	doc = readTestDoc()

	// generate random ops and send to server
	go randomOps(conn)

	// apply ops when received
	for {
		select {
		case op := <-serverOps:
			applyOp(op)
		case op := <-localOps:
			applyOp(op)
		}
	}
}

func applyOp(op *core.Operation) {
	if op.OpType == "i" {
		insertChar(op)
	} else if op.OpType == "d" {
		deleteChar(op)
	} else {
		log.Warn("Unrecognized operation type: " + op.OpType)
		return
	}
	fmt.Print(doc)
}

func insertChar(op *core.Operation) {
	r := []rune(op.Character)
	var buffer bytes.Buffer
	for i, char := range doc {
		buffer.WriteRune(char)
		if i == op.Position {
			buffer.WriteRune(r[0])
		}
	}

	doc = buffer.String()
}

func deleteChar(op *core.Operation) {
	var buffer bytes.Buffer
	for i, char := range doc {
		if i != op.Position {
			buffer.WriteRune(char)
		}
	}

	doc = buffer.String()
}

func readOpsFromServer(conn net.Conn) {
	defer conn.Close()

	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)

	for {
		// Read the incoming connection into the buffer.
		n, err := conn.Read(buf)
		if err != nil {
			log.Warn("Error reading:", err.Error())
			break
		}

		var o core.Operation
		o.OpType = string(buf[0])
		o.Character = string(buf[1])
		if o.Position, err = strconv.Atoi(string(buf[2:n])); err != nil {
			log.Warn("Error: could not parse position int", err.Error())
			break
		}

		serverOps <- &o
	}
}

func randomOps(conn net.Conn) {
	for {
		op := genRandomOp()

		// write op to a channel
		localOps <- op

		sendToServer(conn, op.String())
		time.Sleep(1 * time.Second)
	}
}

func genRandomOp() *core.Operation {
	rand.Seed(time.Now().UTC().UnixNano())
	var o core.Operation

	if rand.Intn(2) == 1 {
		o.OpType = "i"
	} else {
		o.OpType = "d"
	}

	o.Character = string(byte(rand.Intn(26) + 65))
	o.Position = rand.Intn(128)

	return &o
}

func sendToServer(conn net.Conn, op string) {
	conn.Write([]byte(op))
}

func readTestDoc() string {
	b, err := ioutil.ReadFile(*file) // just pass the file name
	if err != nil {
		log.Error(err)
	}

	return string(b)
}
