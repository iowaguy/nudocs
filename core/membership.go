package core

import "sync"

type membership struct {
	peers []Peer
}

type Peer struct {
	Hostname string
	Port     int
}

// a singleton
var instantiated *membership
var onceMemb sync.Once

func NewMembership(peers []Peer) *membership {
	onceMemb.Do(func() {
		instantiated = &membership{}
		instantiated.peers = peers
	})

	return instantiated
}

func NewPeer(hostname string, port int) *Peer {
	p := &Peer{}
	p.Hostname = hostname
	p.Port = port

	return p
}

func GetPeers() []Peer {
	return instantiated.peers
}
