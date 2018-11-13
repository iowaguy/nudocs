package core

import (
	"container/list"
	"sync"

	"github.com/iowaguy/opt/common"
)

type reduce struct {
	historyBuffer *list.List
	proposed      chan *common.PeerOperation
	ready         chan *common.PeerOperation
}

// a singleton
var instantiatedReduce *reduce
var onceRed sync.Once

func NewReducer(peers, pid int) *reduce {
	onceRed.Do(func() {
		instantiatedReduce = &reduce{}
		instantiatedReduce.historyBuffer = list.New()
		instantiatedReduce.proposed = make(chan *common.PeerOperation, 100)
		instantiatedReduce.ready = make(chan *common.PeerOperation, 10)

		common.NewLocalVectorClock(peers, pid)
	})
	return instantiatedReduce
}

// these come from other peers
func (r *reduce) PeerPropose(o common.PeerOperation) {
	// increment vector clock and update according the the peer's vector clock
	common.GetLocalVectorClock().IncrementClock().UpdateClock(&o.VClock)

	r.proposed <- &o
}

// these come from the ui
func (r *reduce) Propose(o common.Operation) {
	// increment vector clock
	common.GetLocalVectorClock().IncrementClock()

	// send to other peers
	for _, peer := range GetPeers() {
		Send2Peer(peer, common.NewPeerOperation(o))
	}
}

// returns a channel of ready operations that a client can access
func (r *reduce) Ready() <-chan *common.PeerOperation {
	return r.ready
}

func (r *reduce) Start() {
	// pop op off proposed queue
	o := <-r.proposed

	var e *list.Element

	// section 1 of REDUCE algorithm
	// search for first operation that is independent of o in historyBuffer
	for e = r.historyBuffer.Front(); e != nil; e = e.Next() {
		po := e.Value.(common.PeerOperation)
		if o.VClock.Independent(&po.VClock) {
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
	for e = k.Next(); e != nil; e = e.Next() {
		po := e.Value.(common.PeerOperation)
		if o.VClock.HappenedAfter(&po.VClock) {
			break
		}
	}

	if e == nil {
		eo := o

		// perform an inclusion trasformation on o against everything in the history buffer, in the language of the paper: EO := LIT(O, L[k,m])
		for e = &k; e != nil; e = e.Next() {
			po := e.Value.(common.PeerOperation)
			eo = InclusionTransformation(eo, &po)
		}

	}

	// section 3 of REDUCE algorithm
	// search for first operation that is independent of o in historyBuffer

}

func (r *reduce) log(o common.Operation) {
	r.historyBuffer.PushBack(o)
}
