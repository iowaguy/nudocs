package core

import (
	"fmt"
	"strconv"
)

type Operation struct {
	OpType    string
	Character string
	Position  int
}

type PeerOperation struct {
	Operation
	VClock VectorClock
}

func (o *Operation) String() string {
	return fmt.Sprintf(o.OpType + o.Character + strconv.Itoa(o.Position))
}

func (o *PeerOperation) String() string {
	return fmt.Sprintf(o.OpType + o.Character + strconv.Itoa(o.Position) + " " + o.VClock.String())
}

type OpTransformer interface {
	PeerPropose(o PeerOperation) // from peers
	Propose(o Operation)         // from ui
	Start()

	// returns operations that are ready to be displayed, blocks if none are available
	Ready()
}

func NewPeerOperation(o Operation) *PeerOperation {
	po := &PeerOperation{}
	po.Operation = o
	po.VClock = *GetLocalVectorClock()

	return po
}
