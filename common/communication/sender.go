package communication

import (
	"net"

	"github.com/iowaguy/nudocs/common"
	"github.com/iowaguy/nudocs/membership"
)

func SendToPeer(p *membership.Peer, op *common.PeerOperation) {
	p.Conn.Write([]byte(op.String()))
}

func SendToAllPeers(peers []membership.Peer, op *common.PeerOperation) {
	for _, p := range peers {
		SendToPeer(&p, op)
	}
}

func SendToServer(conn net.Conn, s string) {
	conn.Write([]byte(s))
}
