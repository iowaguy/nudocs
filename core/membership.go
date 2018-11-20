package core

import "sync"

type Membership struct {
	peers []Peer
}

type Peer struct {
	Hostname string
	Port     int
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

func NewPeer(hostname string, port int) *Peer {
	p := &Peer{}
	p.Hostname = hostname
	p.Port = port

	return p
}

func (m *Membership) GetPeers() []Peer {
	return m.peers
}
