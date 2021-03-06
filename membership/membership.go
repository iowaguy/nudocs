package membership

import (
	"fmt"
	"net"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type Membership struct {
	peers []Peer
	pid   int
}

type Peer struct {
	Hostname string
	Port     int
	Conn     net.Conn
}

const (
	INITIAL_PEER_CAPACITY = 10
)

// a singleton
var (
	instantiated *Membership
	onceMemb     sync.Once
)

func GetMembership() *Membership {
	onceMemb.Do(func() {
		instantiated = &Membership{}
		instantiated.peers = make([]Peer, 0, INITIAL_PEER_CAPACITY)
	})

	return instantiated
}

func (m *Membership) SetPid(pid int) {
	m.pid = pid
}

func (m *Membership) GetPid() int {
	return m.pid
}

func (m *Membership) AddPeer(peer Peer) {
	m.peers = append(m.peers, peer)
}

func NewPeer(hostname string, port int, conn net.Conn) Peer {
	p := Peer{}
	p.Hostname = hostname
	p.Port = port
	p.Conn = conn

	log.Info("New peer=" + p.String())
	return p
}

func (m *Membership) GetPeers() []Peer {
	return m.peers
}

func (p *Peer) String() string {
	return fmt.Sprintf("%v %v %v", p.Hostname, p.Port, p.Conn)
}
