package core

import (
	"sync"
)

type Reduce struct {
	historyBuffer []*PeerOperation
	proposed      chan *PeerOperation
	ready         chan *Operation
}

// a singleton
var instantiatedReduce *Reduce
var onceRed sync.Once

func NewReducer(peers, pid int) *Reduce {
	onceRed.Do(func() {
		instantiatedReduce = &Reduce{}
		instantiatedReduce.historyBuffer = make([]*PeerOperation, 1024)
		instantiatedReduce.proposed = make(chan *PeerOperation, 100)
		instantiatedReduce.ready = make(chan *Operation, 10)

		NewLocalVectorClock(peers, pid)
	})
	return instantiatedReduce
}

// these come from other peers
func (r *Reduce) PeerPropose(o PeerOperation) {
	// increment vector clock and update according the the peer's vector clock
	GetLocalVectorClock().IncrementClock().UpdateClock(&o.VClock)

	r.proposed <- &o
}

// these come from the ui
func (r *Reduce) ClientPropose(o Operation) {
	// increment vector clock
	GetLocalVectorClock().IncrementClock()

	// send to other peers
	for _, peer := range NewMembership(nil).GetPeers() {
		SendToPeer(&peer, NewPeerOperation(o.OpType, o.Character, o.Position))
	}
}

// returns a channel of ready operations that a client can access
func (r *Reduce) Ready() <-chan *Operation {
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
		l1 := make([]*PeerOperation, len(r.historyBuffer))
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
		l1Prime := make([]*PeerOperation, len(r.historyBuffer))
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

func (r *Reduce) Log(o *PeerOperation) {
	r.historyBuffer = append(r.historyBuffer, o)
}

func reverse(sl []*PeerOperation) []*PeerOperation {
	rev := make([]*PeerOperation, len(sl))
	for i := len(sl)/2 - 1; i >= 0; i-- {
		opp := len(sl) - 1 - i
		rev[i] = sl[opp]
	}

	return rev
}
