package core

func SendToPeer(p *Peer, op *PeerOperation) {
	p.Conn.Write([]byte(op.String()))
}

func SendToAllPeers(peers []Peer, op *PeerOperation) {
	for _, p := range peers {
		SendToPeer(&p, op)
	}
}
