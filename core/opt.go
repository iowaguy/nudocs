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
	Ready() <-chan string
	Start()
	HandlePeerEvent(o *common.PeerOperation)
	HandleClientEvent(o *common.Operation)
}

type Reduce struct {
	//the document string
	doc string
	// these are the operations that have been executed
	historyBuffer []*common.PeerOperation
	// these are the causally ready operations
	causallyReady chan *common.PeerOperation
	//when document string changes
	done         chan string
	peerProposed chan *common.PeerOperation
}

// a singleton
var instantiatedReduce *Reduce
var onceRed sync.Once

func GetReducer() *Reduce {
	onceRed.Do(func() {
		instantiatedReduce = &Reduce{}
		instantiatedReduce.historyBuffer = make([]*common.PeerOperation, 0, 1024)
		instantiatedReduce.causallyReady = make(chan *common.PeerOperation)
		instantiatedReduce.peerProposed = make(chan *common.PeerOperation, 1024)
		instantiatedReduce.done = make(chan string)
		//str := "My name is Jaison!"
		//str := "0000000000000000000000000000000000000000000000000000000000000000000000"
		str := ""
		instantiatedReduce.doc = str
	})
	return instantiatedReduce
}

// these come from other peers
func (r *Reduce) HandlePeerEvent(o *common.PeerOperation) {
	log.Debug("Peer proposed an operation: ", o)
	r.peerProposed <- o
}

func (r *Reduce) GetDoc() string {
	return r.doc
}

// these come from the ui
func (r *Reduce) HandleClientEvent(o *common.Operation) {
	log.Debug("Client Proposed an operation: " + o.String())
	// increment vector clock
	clock.GetLocalVectorClock().IncrementClock()
	po := common.NewPeerOperation(o.OpType, o.Character, o.Position)
	r.log(po)
	r.applyOpToDoc(po)
	r.notifyDocValue()
	// send to other peers
	for i, peer := range membership.GetMembership().GetPeers() {
		log.Info("peer=", i, peer.String())
		communication.SendToPeer(&peer, po)
	}
}

func (r *Reduce) notifyDocValue() {
	r.done <- r.doc
}

func (r *Reduce) applyOpToDoc(op *common.PeerOperation) {
	r.doc = common.ApplyOp(&op.Operation, r.doc)
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

// returns a channel of ready operations that a client can access
func (r *Reduce) Ready() <-chan string {
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

			//TODO: This should be totally preceeding (pg: 70)
			if eo.VClock.HappenedBefore(&oNew.VClock) {
				break
			}

			//add these to list to be undone
			o := common.UndoOperation(eo)
			log.Debug("Undo op: " + o.String())
			r.applyOpToDoc(o)
			i = i + 1
		}
		undone := make([]*common.PeerOperation, 0, 1024)
		lastPrecedingOpIndex := len(r.historyBuffer) - i - 1

		copy(undone, r.historyBuffer[lastPrecedingOpIndex+1:])

		// remove everything after (and including) newI + 1 from
		// history. need to do newI + 1, because we don't want to
		// include the operation that happened before oNew
		r.historyBuffer = r.historyBuffer[:lastPrecedingOpIndex+1]

		// (2) Transform Do
		eoNew := r.got(oNew)
		r.applyOpToDoc(eoNew)

		// (3) Transform Redo
		transformedRedos := make([]*common.PeerOperation, 0, 1024)
		if len(undone) > 0 && undone[0] != nil {
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
			r.applyOpToDoc(op)
		}
		r.notifyDocValue()
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

func printOperationList(ops []*common.PeerOperation) {
	log.Debug("Printing ops, length: ")
	log.Debug(len(ops))
	for _, op := range ops {
		if op == nil {
			log.Error("Op is nil")
		}
		log.Debug(op.String())
	}
}

func (r *Reduce) printDocAfterApplyingHistoryBuffer() {
	//doc := "My name is Jaison!"
	doc := "0000000000000000000000000000000000000000000000000000000000000000000000"
	doc = getStringAfterApplyingOps(doc, r.historyBuffer)
	log.Debug(doc)
}

func getStringAfterApplyingOps(doc string, ops []*common.PeerOperation) string {
	for _, op := range ops {
		if op == nil {
			log.Debug("GetStringAfter: ops are null")
			break
		}
		log.Debug("Applying op: " + op.String() + " to doc: " + doc)
		doc = common.ApplyOp(&op.Operation, doc)
	}
	return doc
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
