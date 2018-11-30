package common

import (
	"fmt"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
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

func newOperation(opType, character string, position int) *Operation {
	po := &Operation{}
	po.OpType = opType
	po.Character = character
	po.Position = position

	return po
}

func UndoOperation(op *PeerOperation) *Operation {
	if op.OpType == "i" {
		return newOperation("d", op.Character, op.Position)
	} else {
		return newOperation("i", op.Character, op.Position)
	}
}

func ParseOperation(ops []byte, length int) *PeerOperation {
	log.Info("Peer message=", string(ops[:length]))

	var o PeerOperation
	o.OpType = string(ops[0])
	o.Character = string(ops[1])

	// need to add 3, because: 2 because the slice we're looking at starts
	// at 2, and another 1 because Index tells us the index of the space,
	// but we care about the vector clock which starts one index later
	vcStart := strings.Index(string(ops[2:]), " ") + 3
	log.Info("Vector clock starts at: ", vcStart)
	if vcStart <= 0 {
		log.Panic("Error parsing peer operation")
	}

	var err error
	pos := string(ops[2 : vcStart-1])
	log.Info("position string=", pos)
	if o.Position, err = strconv.Atoi(pos); err != nil {
		log.Panic("Error: could not parse position int: ", err.Error())
	}

	vc := string(ops[vcStart:length])
	log.Info("Vector clock string=", vc)
	if vClock, err := clock.ParseVectorClock(vc); err != nil {
		log.Panic("Error: could not parse vector clock: ", err.Error())
	} else {
		o.VClock = *vClock
	}

	return &o
}
