package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/iowaguy/nudocs/common"
	"github.com/iowaguy/nudocs/common/communication"
)

// TODO later, send document hash to make sure initial document is the same

var (
	doc             string
	localOps        chan *common.Operation
	serverOps       chan string
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
	log.SetReportCaller(false)
	log.SetLevel(log.WarnLevel)
}

func main() {
	flag.Parse()
	log.Info("Host specified=" + *host + "; Port specified=" + strconv.Itoa(*port))
	localOps = make(chan *common.Operation, 100)

	// channel is unbuffered, only supports one at a time
	serverOps = make(chan string)

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
	fmt.Println("Initial Doc: " + doc + "(" + strconv.Itoa(len(doc)) + ")")

	// generate random ops and send to server
	go randomOps(conn)

	go readOpsFromServer(conn)

	// apply ops when received
	for {
		select {
		case op := <-serverOps:
			doc = op
			fmt.Println("Doc From server: " + doc + "(" + strconv.Itoa(len(doc)) + ")")
		case op := <-localOps:
			doc = common.ApplyOp(op, doc)
		}
	}
}

func readOpsFromServer(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	for {
		serverOps <- readString(r)
	}
}

func readString(r *bufio.Reader) string {
	//s, err := r.ReadString([]byte(0x2318))
	s, err := r.ReadString(byte('\n'))
	if err != nil {
		if s == "" {
			log.Panic("Error reading: ", err.Error()+" received: "+s)
		}
		return (s + readString(r))
	}
	return s
}

func randomOps(conn net.Conn) {
	time.Sleep(2 * time.Second)
	// this is just for testing edits in series, no overlap
	if *serialEditsTest {
		mult, err := strconv.Atoi(string((*host)[4]))
		if err != nil {
			log.Panic("something bad happened")
		}
		time.Sleep(3 * time.Duration(mult) * time.Second)
	}

	for ; *ops > 0; *ops-- {
		op := genRandomOp()

		// write op to a channel
		localOps <- op
		fmt.Println("Sending op to server: " + op.String() + " doc length: " + strconv.Itoa(len(doc)))
		communication.SendToServer(conn, op.String()+"\n")
		time.Sleep(1 * time.Second)
	}
}

func getRandomOp(opType string) common.Operation {
	var o common.Operation
	o.OpType = opType
	if opType == "d" {
		o.Character = string(doc[o.Position])
		return o
	}
	o.Character = string(byte(rand.Intn(26) + 65))
	return o
}

func genRandomOp() *common.Operation {
	rand.Seed(time.Now().UTC().UnixNano())
	var o common.Operation
	if rand.Intn(2) == 1 {
		o = getRandomOp("i")
	} else {
		o = getRandomOp("d")
	}
	o.Position = rand.Intn(len(doc))
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
