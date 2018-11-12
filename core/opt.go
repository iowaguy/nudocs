package core

import (
	"container/list"
	"sync"

	"github.com/iowaguy/opt/common"
)

type reduce struct {
	historyBuffer *list.List
	proposed      *list.List
	ready         *list.List
	myClock       *common.VectorClock
}

// a singleton
var instantiated *reduce
var once sync.Once

func NewReducer(peers int) *reduce {
	once.Do(func() {
		instantiated = &reduce{}
		instantiated.historyBuffer = list.New()
		instantiated.proposed = list.New()
		instantiated.ready = list.New()
		instantiated.myClock = common.NewLocalVectorClock(peers)
	})
	return instantiated
}

// these come from other peers
func (r *reduce) PeerPropose(o common.PeerOperation) {
	// increment vector clock

	// update my vector clock according the the peers vector clock
}

// these come from the ui
func (r *reduce) Propose(o common.Operation) {
	// increment vector clock

	r.proposed.PushBack(o)

	// TODO send to other peers

}

func (r *reduce) Ready() {
	// TODO returns operations that are ready to be displayed, blocks if none are available

}

func (r *reduce) Start() {
	// pop op off proposed queue
	eo := r.proposed.Remove(r.proposed.Front())

	// search for first operation that is independent of o in historyBuffer
	for e := r.historyBuffer.Front(); e != nil; e = e.Next() {
		po := e.Value.(common.PeerOperation)
		if r.myClock.Independent(&po.VClock) {
			break
		}
	}

	if eo == nil {
		// put eo in outgoing queue, eo can be exectuted
		r.ready.PushBack(eo)
	}

}

func (r *reduce) log(o common.Operation) {
	r.historyBuffer.PushBack(o)
}
