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

// func NewOperation(opType, character string, position int) *Operation {
// 	o := &Operation{}
// 	o.OpType = opType
// 	o.Character = character
// 	o.Position = position
// 	return o
// }

func NewPeerOperation(opType, character string, position int) *PeerOperation {
	po := &PeerOperation{}
	po.OpType = opType
	po.Character = character
	po.Position = position
	po.VClock = *GetLocalVectorClock()

	return po
}

// func NewPeerOperation(o Operation) *PeerOperation {
// 	po := &PeerOperation{}
// 	po.OpType = o.OpType
// 	po.Character = o.Character
// 	po.Position = o.Position
// 	po.VClock = *GetLocalVectorClock()

// 	return po
// }
