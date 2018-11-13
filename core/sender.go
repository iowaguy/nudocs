package core

import (
	"fmt"
	"net"
	"strconv"
)

func Send2Peer(p Peer, op PeerOperation) {
	SendPeerOperation(p.Hostname, p.Port, op)
}

func SendPeerOperation(host string, port int, op PeerOperation) {
	conn, err := net.Dial("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	conn.Write([]byte(op.String()))
}
