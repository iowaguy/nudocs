package main

import (
	"flag"
	"strconv"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/iowaguy/nudocs/common"
	"github.com/iowaguy/nudocs/common/clock"
	"github.com/iowaguy/nudocs/connectionHandler"
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
	log.SetReportCaller(false)
	log.SetLevel(log.InfoLevel)
}

func main() {
	log.Info("Starting peer server")

	flag.Parse()
	log.Info("Hostsfile specified=" + *hostsfile + "; Port specified=" + strconv.Itoa(*connPort))
	// read hosts file
	hosts := common.ReadHostsFile(*hostsfile)

	myHostID := common.DetermineHostID(hosts)
	membership.GetMembership().SetPid(myHostID)

	log.Info("Initialize vector clock")
	clock.NewLocalVectorClock(len(hosts), myHostID)
	// start algorithm
	go core.GetReducer().Start()

	//ensure connection to client and peer
	connectionHandler.Start(connHost, *connPort, connType, hosts)
	go handleDocumentChange()
	go handleClientEvents()
	go handlePeerEvents()

	// block until a go routine returns, which should never happen
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}

func handleClientEvents() {
	for o := range connectionHandler.ClientEvent {
		core.GetReducer().HandleClientEvent(o)
	}
}

func handlePeerEvents() {
	for o := range connectionHandler.PeerEvent {
		core.GetReducer().HandlePeerEvent(o)
	}
}

func handleDocumentChange() {
	for doc := range core.GetReducer().Ready() {
		log.Debug("Doc Changed: " + doc)
		connectionHandler.SendDocToClient(doc + "\n")
	}
}
