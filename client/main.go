package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/iowaguy/nudocs/common"
	"github.com/iowaguy/nudocs/common/communication"
)

// TODO later, send document hash to make sure initial document is the same

var (
	doc             string
	localOps        chan *common.Operation
	serverOps       chan *common.Operation
	port            = flag.Int("p", 3333, "Server port to connect to")
	host            = flag.String("h", "localhost", "Server hostname to connect to")
	file            = flag.String("f", "", "Path to shared file")
	ops             = flag.Int("o", 10, "Number of random operations to perform")
	serialEditsTest = flag.Bool("t", false, "Number of random operations to perform")
)

func init() {
	Formatter := new(log.TextFormatter)
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true
	log.SetFormatter(Formatter)
	log.SetLevel(log.WarnLevel)
}

func main() {
	flag.Parse()

	log.Info("Host specified=" + *host + "; Port specified=" + strconv.Itoa(*port))
	localOps = make(chan *common.Operation, 100)

	// channel is unbuffered, only supports one at a time
	serverOps = make(chan *common.Operation)

	var conn net.Conn
	for {
		var err error
		conn, err = net.Dial("tcp", *host+":"+strconv.Itoa(*port))
		if err != nil {
			log.Info("Could not connect. Trying again. Error: " + err.Error() + ". This is normal to see a few times at the beginning as the services are starting")
			time.Sleep(1 * time.Second)
		} else {
			log.Info("Client connected to server")
			break
		}
	}
	communication.SendToServer(conn, "client")

	// read document
	doc = readTestDoc()

	// generate random ops and send to server
	go randomOps(conn)

	go readOpsFromServer(conn)

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

func applyOp(op *common.Operation) {

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

func insertChar(op *common.Operation) {
	r := []rune(op.Character)
	var buffer bytes.Buffer
	for i, char := range doc {
		buffer.WriteRune(char)
		if i == op.Position-1 {
			buffer.WriteRune(r[0])
		}
	}

	doc = buffer.String()
}

func deleteChar(op *common.Operation) {
	var buffer bytes.Buffer
	for i, char := range doc {
		if i != op.Position-1 {
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
			log.Error("Error reading:", err.Error())
			os.Exit(1)
		}

		for _, sBuf := range strings.Split(string(buf[:n]), "\n") {
			if len(sBuf) == 0 {
				continue
			}

			var o common.Operation
			o.OpType = string(sBuf[0])
			o.Character = string(sBuf[1])
			if o.Position, err = strconv.Atoi(string(sBuf[2:len(sBuf)])); err != nil {
				log.Warn("Error: could not parse position int: ", err.Error())
				break
			}

			serverOps <- &o
		}
	}
}

func randomOps(conn net.Conn) {

	// this is just for testing edits in series, no overlap
	if *serialEditsTest {
		mult, err := strconv.Atoi(string((*host)[4]))
		if err != nil {
			log.Panic("something bad happened")
		}
		time.Sleep(2 * time.Duration(mult) * time.Second)
	}

	for ; *ops > 0; *ops-- {
		op := genRandomOp()

		// write op to a channel
		localOps <- op

		communication.SendToServer(conn, op.String()+"\n")
		time.Sleep(1 * time.Second)
	}
}

func genRandomOp() *common.Operation {
	rand.Seed(time.Now().UTC().UnixNano())
	var o common.Operation

	o.Position = rand.Intn(len(doc))
	if rand.Intn(2) == 1 {
		o.OpType = "i"
		o.Character = string(byte(rand.Intn(26) + 65))
	} else {
		o.OpType = "d"
		o.Character = string(doc[o.Position])
	}

	log.Info("op=", o.String())

	return &o
}

func readTestDoc() string {
	b, err := ioutil.ReadFile(*file) // just pass the file name
	if err != nil {
		log.Error(err)
	}

	return string(b)
}
