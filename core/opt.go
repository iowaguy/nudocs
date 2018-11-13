package core

import (
	"container/list"
	"sync"

	"github.com/iowaguy/opt/common"
)

type reduce struct {
	historyBuffer *list.List
	proposed      chan *Operation
	ready         chan *Operation
}

// a singleton
var instantiated *reduce
var once sync.Once

func NewReducer(peers, pid int) *reduce {
	once.Do(func() {
		instantiated = &reduce{}
		instantiated.historyBuffer = list.New()
		instantiated.proposed = make(chan *Operation, 100)
		instantiated.ready = make(chan *Operation, 10)

		common.NewLocalVectorClock(peers, pid)
	})
	return instantiated
}

// these come from other peers
func (r *reduce) PeerPropose(o common.PeerOperation) {
	// increment vector clock and update according the the peer's vector clock
	GetLocalVectorClock().IncrementClock().UpdateClock(o.VClock)

	r.proposed <- o
}

// these come from the ui
func (r *reduce) Propose(o common.Operation) {
	// increment vector clock
	GetLocalVectorClock().IncrementClock()

	// send to other peers
	for _, peer := range GetPeers() {
		Send2Peer(peer, NewPeerOperation(o))
	}
}

func (r *reduce) Ready() {

	for {
		// gets operations that are ready to be displayed, blocks if none are available
		o := <-r.ready

		// TODO send to client
	}

}

func (r *reduce) Start() {
	// pop op off proposed queue
	o := <-r.proposed

	// section 1 of REDUCE algorithm
	// search for first operation that is independent of o in historyBuffer
	for e := r.historyBuffer.Front(); e != nil; e = e.Next() {
		po := e.Value.(common.PeerOperation)
		if r.myClock.Independent(&po.VClock) {
			break
		}
	}

	if e == nil {
		// put o in outgoing queue, o can be exectuted
		r.ready <- o
	}

	// section 2 of REDUCE algorithm
	// if im here, o is independent of e
	// look for operations causally preceding o
	k := *e
	for e = e.Next(); e != nil; e = e.Next() {
		po := e.Value.(common.PeerOperation)
		if r.myClock.HappenedAfter(&po.VClock) {
			break
		}
	}

	if e == nil {
		eo := o

		// perform an inclusion trasformation of o against everything in the history buffer, in the language of the paper: EO := LIT(O, L[k,m])
		for e = &k; e != nil; e = e.Next() {
			po := e.Value.(common.PeerOperation)
			eo = InclusionTransformation(eo, po)
		}

	}

	// section 3 of REDUCE algorithm
	// search for first operation that is independent of o in historyBuffer

}

func (r *reduce) log(o common.Operation) {
	r.historyBuffer.PushBack(o)
}
