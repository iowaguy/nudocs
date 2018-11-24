package core

import (
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/iowaguy/nudocs/common"
	"github.com/iowaguy/nudocs/common/clock"
	"github.com/iowaguy/nudocs/common/communication"
	"github.com/iowaguy/nudocs/membership"
)

type OpTransformer interface {
	Log(o *common.PeerOperation)
	Ready() <-chan *common.Operation
	Start()
	PeerPropose(o *common.PeerOperation)
	ClientPropose(o *common.Operation)
}

type Reduce struct {
	historyBuffer []*common.PeerOperation
	proposed      chan *common.PeerOperation
	ready         chan *common.Operation
}

// a singleton
var instantiatedReduce *Reduce
var onceRed sync.Once

func GetReducer() *Reduce {
	onceRed.Do(func() {
		instantiatedReduce = &Reduce{}
		instantiatedReduce.historyBuffer = make([]*common.PeerOperation, 0, 1024)
		instantiatedReduce.proposed = make(chan *common.PeerOperation, 1024)

		// the server can only have one ready operation at a time, will
		// need to force this by making the channel unbuffered, so
		// communication is synchronous. this is necessary because
		// otherwise, an operation could happen locally and be
		// unaccounted for in operations waiting in the ready queue.
		// a new operation can only be processed after the
		// ready queue is emptied.
		instantiatedReduce.ready = make(chan *common.Operation)
	})
	return instantiatedReduce
}

// these come from other peers
func (r *Reduce) PeerPropose(o *common.PeerOperation) {
	// increment vector clock and update according the the peer's vector clock
	clock.GetLocalVectorClock().IncrementClock().UpdateClock(&o.VClock)

	r.proposed <- o
}

// these come from the ui
func (r *Reduce) ClientPropose(o *common.Operation) {
	// increment vector clock
	clock.GetLocalVectorClock().IncrementClock()

	// send to other peers
	for i, peer := range membership.GetMembership().GetPeers() {
		log.Info("peer=", i, peer.String())
		communication.SendToPeer(&peer, common.NewPeerOperation(o.OpType, o.Character, o.Position))
	}
}

// returns a channel of ready operations that a client can access
func (r *Reduce) Ready() <-chan *common.Operation {
	return r.ready
}

func (r *Reduce) Start() {
	for {
		// pop op off proposed queue
		o := <-r.proposed

		// section 1 of REDUCE algorithm
		// search for first operation that is independent of o in historyBuffer
		noOpsIndependent := true
		var k int
		for i, po := range r.historyBuffer {
			k = i
			if o.VClock.Independent(&po.VClock) {
				noOpsIndependent = false
				break
			}
		}

		if noOpsIndependent {
			// put o in outgoing queue, o can be exectuted
			r.ready <- &o.Operation
			continue
		}

		// section 2 of REDUCE algorithm
		// if im here, o is independent of e
		// look for operations causally preceding o
		noOpsCausallyPreceding := true
		for _, po := range r.historyBuffer[k+1:] {
			if o.VClock.HappenedAfter(&po.VClock) {
				noOpsCausallyPreceding = false
				break
			}
		}

		if noOpsCausallyPreceding {
			// perform an inclusion trasformation on o against everything in
			// the history buffer, in the language of the paper:
			// EO := LIT(O, L[k,m])
			eo := LIT(o, r.historyBuffer[k:])
			r.ready <- &eo.Operation
			continue
		}

		// if im here, then there is at least one operation which causally
		// precedes o, but comes after an operation which is independent of o

		// section 3 of REDUCE algorithm
		// generate a list L1 which contains the operations in L[k,m] which are
		// causally preceding o

		// cs is a slice of the indexes of operations which are causally
		// preceding o
		cs := make([]int, len(r.historyBuffer))
		l1 := make([]*common.PeerOperation, len(r.historyBuffer))
		for i, po := range r.historyBuffer[k:] {
			if o.VClock.HappenedAfter(&po.VClock) {
				cs = append(cs, i)
				l1 = append(l1, po)
			}
		}

		// c1 is the first causally preceding operation following at least
		// one independent operation of o
		c1 := cs[0]
		eoc1Prime := LET(l1[0], reverse(r.historyBuffer[k:c1-1]))
		l1Prime := make([]*common.PeerOperation, len(r.historyBuffer))
		l1Prime = append(l1Prime, eoc1Prime)
		for i, eoci := range l1[1:] {
			ci := cs[i-1]
			ot := LET(eoci, reverse(r.historyBuffer[k:ci-1]))
			eociPrime := LIT(ot, l1Prime)
			l1Prime = append(l1Prime, eociPrime)
		}

		oPrime := LET(o, reverse(l1Prime))

		eo := LIT(oPrime, r.historyBuffer[k:])
		r.ready <- &eo.Operation
	}
}

func (r *Reduce) Log(o *common.PeerOperation) {
	r.historyBuffer = append(r.historyBuffer, o)
}

func reverse(sl []*common.PeerOperation) []*common.PeerOperation {
	rev := make([]*common.PeerOperation, len(sl))
	for i := len(sl)/2 - 1; i >= 0; i-- {
		opp := len(sl) - 1 - i
		rev[i] = sl[opp]
	}

	return rev
}
