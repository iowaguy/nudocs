package main

import (
	"bufio"
	"flag"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/iowaguy/nudocs/core"
)

const (
	connHost = "0.0.0.0"
	connType = "tcp"
)

var (
	connPort  = flag.Int("p", 3333, "Server port to listen on")
	hostsfile = flag.String("h", "", "Path to the hosts file")
)

func init() {
	Formatter := new(log.TextFormatter)
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true
	log.SetFormatter(Formatter)
}

func main() {
	log.Info("Starting server")

	flag.Parse()

	log.Info("Hostsfile specified=" + *hostsfile + "; Port specified=" + strconv.Itoa(*connPort))
	// read hosts file
	hosts, err := readLines(*hostsfile)
	if err != nil {
		log.Error("Could not read hostsfile: " + err.Error())
		os.Exit(1)
	}
	log.Info("Hosts in hostsfile=" + strings.Join(hosts, " "))

	myHostname, err := os.Hostname()
	if err != nil {
		log.Error("Could determine hostname: " + err.Error())
		os.Exit(1)
	}

	log.Info("Local hostame=" + myHostname)

	// determine pid from hostsfile
	var myHostID int
	for i, h := range hosts {
		if h == myHostname {
			myHostID = i
		}
	}
	log.Info("Local hostame=" + myHostname + "; host ID=" + strconv.Itoa(myHostID))

	// start algorithm
	red := core.NewReducer(len(hosts), hostID)
	go red.Start()

	// Listen for incoming connections.
	log.Info("Listen for incoming connections")
	l, err := net.Listen(connType, connHost+":"+strconv.Itoa(*connPort))
	if err != nil {
		log.Error("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()
	log.Info("Listening on " + connHost + ":" + strconv.Itoa(*connPort))

	go acceptNewConnections(l, red)

	// TODO connect to other peers

	// block until a go routine returns, which should never happen
	var wg sync.WaitGroup
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

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
