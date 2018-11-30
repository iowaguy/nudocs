package core

import (
	log "github.com/Sirupsen/logrus"
	"github.com/iowaguy/nudocs/common"
	"strconv"
)

func InclusionTransformation(o1, o2 *common.PeerOperation) *common.PeerOperation {
	// TODO needs to determine exactly which function in the transformation matrix to call
	return o1
}

func ExclusionTransformation(o1, o2 *common.PeerOperation) *common.PeerOperation {
	// TODO needs to determine exactly which function in the transformation matrix to call
	return o1
}

func LET(eo *common.PeerOperation, l []*common.PeerOperation) *common.PeerOperation {
	for _, po := range l {
		eo = ExclusionTransformation(eo, po)
	}

	return eo
}

func LIT(eo *common.PeerOperation, l []*common.PeerOperation) *common.PeerOperation {
	for _, po := range l {
		eo = InclusionTransformation(eo, po)
	}

	return eo
}

func Save_LI(o1_1, o1, o2 *common.PeerOperation) {
	//TODO:
	log.Warn("Save_LI is not implemented yet")
}

func Save_RA(o1, o2 *common.PeerOperation) {
	//TODO:
	log.Warn("Save_RA is not implemented yet")
}

func Check_LI(o1, o2 *common.PeerOperation) bool {
	//TODO:
	log.Warn("Check_LI is not implemented yet. Returning false by default.")
	return false
}

func Recover_LI(o *common.PeerOperation) *common.PeerOperation {
	//TODO:
	log.Warn("Recover_LI is not implemented yet. Returning same operation.")
	return o
}

func IT_II(o1, o2 *common.PeerOperation) *common.PeerOperation {
	var o1_1 *common.PeerOperation
	if P(o1) < P(o2) {
		o1_1 = o1
	} else {
		o1_1 = Insert(S(o1), P(o1)+L(o2))
	}
	return o1_1
}

func IT_ID(o1, o2 *common.PeerOperation) *common.PeerOperation {
	var o1_1 *common.PeerOperation
	if P(o1) <= P(o2) {
		o1_1 = o1
	} else if P(o1) > P(o2)+L(o2) {
		o1_1 = Insert(S(o1), P(o1)-L(o2))
	} else {
		Insert(S(o1), P(o2))
		Save_LI(o1_1, o1, o2)
	}
	return o1_1
}

func IT_DI(o1, o2 *common.PeerOperation) *common.PeerOperation {
	var o1_1 *common.PeerOperation
	if P(o2) >= P(o1)+L(o1) {
		o1_1 = o1
	} else if P(o1) >= P(o2) {
		o1_1 = Delete(L(o1), P(o1)+L(o2))
	} else {
		log.Panic("IT_DI 3rd condition not handled.")
		/*
			o1_1 = Delete(P(o2)-P(o1), P(o1)) + Delete(L(o1)-(P(o2)-P(o1)), P(o2) + L(o2))
		*/
	}
	return o1_1
}

func IT_DD(o1, o2 *common.PeerOperation) *common.PeerOperation {
	var o1_1 *common.PeerOperation
	if P(o2) >= P(o1)+L(o1) {
		o1_1 = o1
	} else if P(o1) >= P(o2)+L(o2) {
		o1_1 = Delete(L(o1), P(o1)-L(o2))
	} else {
		if P(o2) <= P(o1) && (P(o1)+L(o1)) <= (P(o2)+L(o2)) {
			o1_1 = Delete(0, P(o1))
		} else if P(o2) <= P(o1) && (P(o1)+L(o1)) > (P(o2)+L(o2)) {
			o1_1 = Delete(P(o1)+L(o1)-(P(o2)+L(o2)), P(o2))
		} else if P(o2) > P(o1) && (P(o2)+L(o2)) >= (P(o1)+L(o1)) {
			o1_1 = Delete(P(o2)-P(o1), P(o1))
		} else {
			o1_1 = Delete(L(o1)-L(o2), P(o1))
		}
		Save_LI(o1_1, o1, o2)
	}
	return o1_1
}

func ET_II(o1, o2 *common.PeerOperation) *common.PeerOperation {
	var o1_1 *common.PeerOperation
	if P(o1) <= P(o2) {
		o1_1 = o1
	} else if P(o1) >= (P(o2) + L(o2)) {
		o1_1 = Insert(S(o1), P(o1)-L(o2))
	} else {
		o1_1 = Insert(S(o1), P(o1)-P(o2))
		Save_RA(o1_1, o2)
	}
	return o1_1
}

func ET_ID(o1, o2 *common.PeerOperation) *common.PeerOperation {
	var o1_1 *common.PeerOperation
	if Check_LI(o1, o2) {
		o1_1 = Recover_LI(o1)
	} else if P(o1) <= P(o2) {
		o1_1 = o1
	} else {
		o1_1 = Insert(S(o1), P(o1)+L(o2))
	}
	return o1_1
}

func ET_DI(o1, o2 *common.PeerOperation) *common.PeerOperation {
	var o1_1 *common.PeerOperation
	if (P(o1) + L(o1)) <= P(o2) {
		o1_1 = o1
	} else if P(o1) >= (P(o2) + L(o2)) {
		o1_1 = Delete(L(o1), P(o1)-L(o2))
	} else {
		if P(o2) <= P(o1) && (P(o1)+L(o1)) <= (P(o2)+L(o2)) {
			o1_1 = Delete(L(o1), P(o1)-P(o2))
		} else if P(o2) <= P(o1) && (P(o1)+L(o1) > P(o2)+L(o2)) {
			/**
			o1_1 = Delete(P(o2)+L(o2)-P(o1), (P(o1)-P(o2))) + Delete((P(o1) + L(o1) - (P(o2) + L(o2)), P(o2))
			**/
			log.Panic("ET_DI XOR condition 1 not handled.")
		} else if P(o1) < P(o2) && (P(o2)+L(o2)) <= (P(o1)+L(o1)) {
			/**
			o1_1 = Delete(L(o2), 0) + Delete(L(o1)-L(o2), P(o1))
			**/
			log.Panic("ET_DI XOR condition 2 not handled.")
		} else {
			/**
			o1_1 = Delete(P(o1) + L(o1) - P(o2), 0) + Delete(P(o2)-P(o1), P(o1))
			**/
			log.Panic("ET_DI XOR condition 3 not handled.")
		}
		Save_RA(o1_1, o2)
	}
	return o1_1
}

func ET_DD(o1, o2 *common.PeerOperation) *common.PeerOperation {
	var o1_1 *common.PeerOperation
	if Check_LI(o1, o2) {
		o1_1 = Recover_LI(o1)
	} else if P(o2) >= (P(o1) + L(o1)) {
		o1_1 = o1
	} else if P(o1) >= P(o2) {
		o1_1 = Delete(L(o1), P(o1)+L(o2))
	} else {
		/**
		o1_1 = Delete(P(o2) - P(o1), P(o1)) + Delete(L(o1) - (P(o2) - P(o1)), P(o2) + L(o2))
		**/
		log.Panic("ET_DD XOR condition not handled.")
	}
	return o1_1
}

func P(o *common.PeerOperation) int {
	return o.Operation.Position
}

func L(o *common.PeerOperation) int {
	return len(o.Operation.Character)
}

func S(o *common.PeerOperation) string {
	return o.Operation.Character
}

func Insert(content string, position int) *common.PeerOperation {
	return common.NewPeerOperation("i", content, position)
}

func Delete(length, position int) *common.PeerOperation {
	return common.NewPeerOperation("d", strconv.Itoa(length), position)
}
