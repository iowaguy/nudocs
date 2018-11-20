package core

import (
	"net"
	"sync"
)

type Membership struct {
	peers []Peer
}

type Peer struct {
	Hostname string
	Port     int
	Conn     net.Conn
}

// a singleton
var (
	instantiated *Membership
	onceMemb     sync.Once
)

func NewMembership(peers []Peer) *Membership {
	onceMemb.Do(func() {
		instantiated = &Membership{}
		instantiated.peers = peers
	})

	return instantiated
}

func NewPeer(hostname string, port int, conn net.Conn) *Peer {
	p := &Peer{}
	p.Hostname = hostname
	p.Port = port
	p.Conn = conn

	return p
}

func (m *Membership) GetPeers() []Peer {
	return m.peers
}
