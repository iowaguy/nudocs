package common

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"strconv"
)

func ApplyOp(op *Operation, doc string) string {
	fmt.Println("Applying operation: " + op.String() + " doc length: " + strconv.Itoa(len(doc)))
	if op.OpType == "i" {
		temp1 := doc[:op.Position]
		temp2 := doc[op.Position:]
		return temp1 + op.Character + temp2
	} else if op.OpType == "d" {
		temp1 := doc[:op.Position]
		temp2 := doc[op.Position+1:]
		return temp1 + temp2
	} else {
		log.Warn("Unrecognized operation type: " + op.OpType)
		return doc
	}
}
