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
var once sync.Once

func NewMembership(peers []Peer) *membership {
	once.Do(func() {
		instantiated = &membership{}
		instantiated.peers = peers
	})

	return instantiated
}

func NewPeer(hostname string, port int) *peer {
	p := &peer{}
	p.Hostname = hostname
	p.Port = port
}

func GetPeers() []Peer {
	return instantiated.peers
}
