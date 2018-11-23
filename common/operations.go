package common

import (
	"fmt"
	"strconv"

	"github.com/iowaguy/nudocs/common/clock"
)

type Operation struct {
	OpType    string
	Character string
	Position  int
}

type PeerOperation struct {
	Operation
	VClock clock.VectorClock
}

func (o *Operation) String() string {
	return fmt.Sprintf(o.OpType + o.Character + strconv.Itoa(o.Position))
}

func (o *PeerOperation) String() string {
	return fmt.Sprintf(o.OpType + o.Character + strconv.Itoa(o.Position) + " " + o.VClock.String())
}

func NewPeerOperation(opType, character string, position int) *PeerOperation {
	po := &PeerOperation{}
	po.OpType = opType
	po.Character = character
	po.Position = position
	po.VClock = *clock.GetLocalVectorClock()

	return po
}
