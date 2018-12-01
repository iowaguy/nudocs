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
	Ready() <-chan *common.Operation
	Start()
	PeerPropose(o *common.PeerOperation)
	ClientPropose(o *common.Operation)
}

type Reduce struct {
	// these are the operations that have been executed
	historyBuffer []*common.PeerOperation

	// these are the causally ready operations
	causallyReady chan *common.PeerOperation

	// these are operations that have been transformed and are ready to be
	// applied by the client
	done chan *common.Operation

	peerProposed chan *common.PeerOperation
}

// a singleton
var instantiatedReduce *Reduce
var onceRed sync.Once

func GetReducer() *Reduce {
	onceRed.Do(func() {
		instantiatedReduce = &Reduce{}
		instantiatedReduce.historyBuffer = make([]*common.PeerOperation, 0, 1024)

		// the server can only have one ready operation at a time, will
		// need to force this by making the channel unbuffered, so
		// communication is synchronous. this is necessary because
		// otherwise, an operation could happen locally and be
		// unaccounted for in operations waiting in the ready queue.
		// a new operation can only be processed after the
		// causallyReady channel is emptied.
		instantiatedReduce.causallyReady = make(chan *common.PeerOperation)
		instantiatedReduce.peerProposed = make(chan *common.PeerOperation, 1024)
		instantiatedReduce.done = make(chan *common.Operation, 1024)
	})
	return instantiatedReduce
}

// these come from other peers
func (r *Reduce) PeerPropose(o *common.PeerOperation) {
	log.Info("Peer proposed an operation: ", o)
	r.peerProposed <- o
}

func (r *Reduce) queueCausallyReady() {

	// find causally ready operation
	for proposed := range r.peerProposed {
		if clock.GetLocalVectorClock().CausallyPreceding(&proposed.VClock) {
			log.Info("Found causally ready operation: ", proposed)
			r.causallyReady <- proposed
			clock.GetLocalVectorClock().UpdateClock(&proposed.VClock)
		} else {
			log.Info("Op is not causally ready. my state: ", clock.GetLocalVectorClock(), "; proposed clock: ", proposed.VClock)
			// put operation back on channel until it's causally ready
			r.peerProposed <- proposed
		}
	}
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
	return r.done
}

func (r *Reduce) Start() {
	go r.queueCausallyReady()

	// pop causally ready operation off proposed queue
	for oNew := range r.causallyReady {

		// (1) Undo
		var i int
		for _, eo := range reverse(r.historyBuffer) {
			if eo == nil {
				break
			}

			if eo.VClock.HappenedBefore(&oNew.VClock) {
				break
			}

			//   write the undo of each operation in HB to ready
			r.done <- common.UndoOperation(eo)
		}

		undone := make([]*common.PeerOperation, 1024)
		lastPrecedingOpIndex := len(r.historyBuffer) - i - 1

		copy(undone, r.historyBuffer[lastPrecedingOpIndex+1:])

		// remove everything after (and including) newI + 1 from
		// history. need to do newI + 1, because we don't want to
		// include the operation that happened before oNew
		r.historyBuffer = r.historyBuffer[:lastPrecedingOpIndex+1]

		// (2) Transform Do
		eoNew := r.got(oNew)

		// (3) Transform Redo
		transformedRedos := make([]*common.PeerOperation, 0, 1024)
		if undone[0] != nil {
			eom1Prime := IT(undone[0], eoNew)
			undone = undone[1:]

			transformedRedos = append(transformedRedos, eoNew, eom1Prime)
			for i, eomi := range undone {
				// (3.1)
				to := LET(eomi, reverse(undone[:i]))

				// (3.2)
				eomiPrime := LIT(to, transformedRedos)
				transformedRedos = append(transformedRedos, eomiPrime)
			}
		} else {
			transformedRedos = append(transformedRedos, eoNew)
		}

		// write transformed ops to ready
		for _, op := range transformedRedos {
			if op == nil {
				break
			}
			r.log(op)
			r.done <- &op.Operation
		}
	}
}

// This is the Generic Operational Transformation algorithm
func (r *Reduce) got(o *common.PeerOperation) *common.PeerOperation {
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
		return o
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
		return eo
	}

	// if im here, then there is at least one operation which causally
	// precedes o, but comes after an operation which is independent of o

	// section 3 of REDUCE algorithm
	// generate a list L1 which contains the operations in L[k,m] which are
	// causally preceding o

	// cs is a slice of the indexes of operations which are causally
	// preceding o
	cs := make([]int, 0, len(r.historyBuffer))
	l1 := make([]*common.PeerOperation, 0, len(r.historyBuffer))
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
	l1Prime := make([]*common.PeerOperation, 0, len(r.historyBuffer))
	l1Prime = append(l1Prime, eoc1Prime)
	for i, eoci := range l1 {
		if i == 0 {
			// skip the first, because c1 was already caluclated
			continue
		}
		ci := cs[i-1]
		ot := LET(eoci, reverse(r.historyBuffer[k:ci-1]))
		eociPrime := LIT(ot, l1Prime)
		l1Prime = append(l1Prime, eociPrime)
	}

	oPrime := LET(o, reverse(l1Prime))

	eo := LIT(oPrime, r.historyBuffer[k:])
	return eo
}

func (r *Reduce) log(o *common.PeerOperation) {
	r.historyBuffer = append(r.historyBuffer, o)
}

func reverse(sl []*common.PeerOperation) []*common.PeerOperation {
	rev := make([]*common.PeerOperation, len(sl))
	for i, v := range sl {
		opp := len(sl) - 1 - i
		rev[opp] = v
	}

	return rev
}

func del(a []*common.PeerOperation, i int) []*common.PeerOperation {
	copy(a[i:], a[i+1:])
	a[len(a)-1] = nil // or the zero value of T
	a = a[:len(a)-1]
	return a
}
