package connectionHandler

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/googollee/go-socket.io"
	"github.com/iowaguy/nudocs/common"
	"github.com/iowaguy/nudocs/common/communication"
	"github.com/iowaguy/nudocs/core"
	"github.com/iowaguy/nudocs/membership"
)

var (
	performOnce     sync.Once
	connPort        int
	connHost        string
	connType        string
	clientConn      net.Conn        = nil
	socketIOEmitter socketio.Socket = nil
	Ready                           = make(chan int)
	ClientEvent                     = make(chan *common.Operation)
	PeerEvent                       = make(chan *common.PeerOperation)
)

func Start(host string, port int, ctype string, peers []string) {
	connPort = port
	connHost = host
	connType = ctype
	go awaitSocketIoConnection()
	go awaitTcpConnections()
	//blocks until all peers are connected
	connectToPeers(peers)
	//blocks until atleast one client is connected
	for !isClientConnected() {
	}
}

func isClientConnected() bool {
	return clientConn != nil || socketIOEmitter != nil
}

func awaitSocketIoConnection() {
	AcceptSocketIoConnectionOn(8080,
		func(event string, so socketio.Socket) {
			log.Info("Received socketIO event: " + event)
			switch event {
			case "connection":
				if isClientConnected() {
					log.Panic("More than one client is trying to connect.")
				}
				socketIOEmitter = so
				emitDocToClient()
			case "disconnection":
				socketIOEmitter = nil
			case "error":
				socketIOEmitter = nil
			}
		},
		func(msg string) {
			fmt.Println(msg)
			o := common.ParseOperationFromString(msg)
			ClientEvent <- o
		})
}

func awaitTcpConnections() {
	// Listen for incoming connections.
	l, err := net.Listen(connType, connHost+":"+strconv.Itoa(connPort))
	if err != nil {
		log.Error("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()
	log.Info("Listening on " + connHost + ":" + strconv.Itoa(connPort))
	for {
		// Listen for an incoming connection.
		log.Trace("Waiting for client or peer to connect")
		conn, err := l.Accept()
		if err != nil {
			log.Error("Error accepting: ", err.Error())
			os.Exit(1)
		}
		log.Info("Connection received, determining who it is...")
		initMsg := getBufferString(conn)
		if isPeer(initMsg) {
			log.Info("Connected to peer: " + conn.RemoteAddr().String())
			go receivePeerEvents(conn)
		} else if isClient(initMsg) {
			log.Info("Connected to client")
			if isClientConnected() {
				log.Panic("More than one client is trying to connect.")
			}
			clientConn = conn
			go receiveClientEvents(conn)
			emitDocToClient()
		}
	}
}

func emitDocToClient() {
	doc := core.GetReducer().GetDoc()
	SendDocToClient(doc)
}

func receivePeerEvents(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	for {
		o := common.ParsePeerOperation(r)
		PeerEvent <- o
	}
}

func receiveClientEvents(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	for {
		o := common.ParseOperation(r)
		ClientEvent <- o
	}
}

func SendDocToClient(doc string) {
	docLength := strconv.Itoa(len(doc))
	toSend := docLength + ":" + doc
	if socketIOEmitter != nil {
		socketIOEmitter.Emit(EMIT_TO_CLIENT, toSend)
		return
	}
	clientConn.Write([]byte(toSend))
}

func isPeer(msg string) bool {
	return msg == "peer"
}

func isClient(msg string) bool {
	return msg == "client"
}

func getBufferString(conn net.Conn) string {
	buf := make([]byte, 256)
	// Read the incoming connection into the buffer.
	n, err := conn.Read(buf)
	if err != nil {
		log.Panic("Error reading: ", err.Error())
	}
	return string(buf[:n])
}

// This function will not return until connections have been established with all peers
func connectToPeers(peers []string) {
	for _, h := range peers {
		// ignore self connection
		if h == peers[membership.GetMembership().GetPid()] {
			continue
		}

		var conn net.Conn
		// retry connection until it succeeds
		for {
			var err error
			conn, err = net.Dial("tcp", h+":"+strconv.Itoa(connPort))
			if err != nil {
				log.Trace("Could not connect. Trying again. Error: " + err.Error() + ". This is normal to see a few times at the beginning as the services are starting")
				time.Sleep(500 * time.Millisecond)
			} else {
				break
			}
		}

		// send wakeup message to server
		communication.SendToServer(conn, "peer")

		peer := membership.NewPeer(h, connPort, conn)
		membership.GetMembership().AddPeer(peer)
	}
}
