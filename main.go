package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	client "github.com/iowaguy/nudocs/client/core"
	"github.com/iowaguy/nudocs/common"
	"github.com/iowaguy/nudocs/common/clock"
	"github.com/iowaguy/nudocs/common/communication"
	"github.com/iowaguy/nudocs/core"
	"github.com/iowaguy/nudocs/membership"
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
	log.Info("Starting peer server")

	flag.Parse()
	log.Info("Hostsfile specified=" + *hostsfile + "; Port specified=" + strconv.Itoa(*connPort))
	// read hosts file
	hosts := readHostsFile(*hostsfile)

	myHostID := determineHostID(hosts)

	log.Info("Initialize vector clock")
	clock.NewLocalVectorClock(len(hosts), myHostID)

	// Listen for incoming connections.
	log.Info("Listen for incoming connections")
	l, err := net.Listen(connType, connHost+":"+strconv.Itoa(*connPort))
	if err != nil {
		log.Error("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()
	log.Info("Listening on " + connHost + ":" + strconv.Itoa(*connPort))

	go acceptNewConnections(l)

	// connect to other peers
	connectToPeers(hosts)

	// block until client is connected
	<-client.ClientConnected

	// can pass in nil as client arg, because a client will have already been created
	client.NewClient(nil).Start(core.GetReducer())

	// start algorithm
	go core.GetReducer().Start()

	// block until a go routine returns, which should never happen
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}

// This function will not return until connections have been established with all peers
func connectToPeers(peers []string) {
	for _, h := range peers {
		var conn net.Conn
		// retry connection until it succeeds
		for {
			var err error
			conn, err = net.Dial("tcp", h+":"+strconv.Itoa(*connPort))
			if err != nil {
				log.Warn("Could not connect. Trying again. Error: " + err.Error())
				time.Sleep(500 * time.Millisecond)
			} else {
				log.Info("Client connected to server")
				break
			}
		}

		// send wakeup message to server
		communication.SendToServer(conn, "peer")

		peer := membership.NewPeer(h, *connPort, conn)
		membership.GetMembership().AddPeer(peer)
	}
}

func acceptNewConnections(l net.Listener) {
	for {
		// Listen for an incoming connection.
		log.Info("Waiting for client or peer to connect")
		conn, err := l.Accept()
		if err != nil {
			log.Error("Error accepting:", err.Error())
			os.Exit(1)
		}
		log.Info("Connection received, determining who it is...")

		if isPeer(conn) {
			log.Info("Connected to peer")
			go receivePeerOperations(conn, core.GetReducer())
		} else {
			log.Info("Connected to client")
			// there will only be one client, in fact, the client
			// is a singleton to guarantee this
			c := client.NewClient(conn)
			c.Start(core.GetReducer())
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

func readHostsFile(filename string) []string {
	hosts, err := readLines(*hostsfile)
	if err != nil {
		log.Error("Could not read hostsfile: " + err.Error())
		os.Exit(1)
	}
	log.Info("Hosts in hostsfile=" + strings.Join(hosts, " "))

	return hosts
}

func determineHostID(hosts []string) int {
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
	return myHostID
}

func receivePeerOperations(conn net.Conn, ot core.OpTransformer) {
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
		var o common.PeerOperation
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

		if vClock, err := clock.ParseVectorClock(string(buf[vcStart:n])); err != nil {
			fmt.Println("Error: could not parse vector clock", err.Error())
		} else {
			o.VClock = *vClock
		}

		// send operation to algorithm to be processed
		ot.PeerPropose(o)
	}
}

func isPeer(conn net.Conn) bool {
	buf := make([]byte, 256)

	// Read the incoming connection into the buffer.
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		os.Exit(1)
	}

	return string(buf[:n]) == "peer"
}
