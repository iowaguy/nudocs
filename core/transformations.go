package core

import (
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/iowaguy/nudocs/common"
)

func IT(o1, o2 *common.PeerOperation) *common.PeerOperation {
	o1Operation := o1.OpType
	o2Operation := o1.OpType
	if o1Operation == "i" {
		if o2Operation == "i" {
			return IT_II(o1, o2)
		} else if o2Operation == "d" {
			return IT_ID(o1, o2)
		}
		log.Panic("Unknown opertaion received for o2 in IT")
	} else if o1Operation == "d" {
		if o2Operation == "i" {
			return IT_DI(o1, o2)
		} else if o2Operation == "d" {
			return IT_DD(o1, o2)
		}
		log.Panic("Unknown opertaion received for o2 in IT")
	}
	log.Panic("Unknown opertaion received for o1 in IT")
	return nil
}

func ET(o1, o2 *common.PeerOperation) *common.PeerOperation {
	o1Operation := o1.OpType
	o2Operation := o1.OpType
	if o1Operation == "i" {
		if o2Operation == "i" {
			return ET_II(o1, o2)
		} else if o2Operation == "d" {
			return ET_ID(o1, o2)
		}
		log.Panic("Unknown opertaion received for o2 in ET")
	} else if o1Operation == "d" {
		if o2Operation == "i" {
			return ET_DI(o1, o2)
		} else if o2Operation == "d" {
			return ET_DD(o1, o2)
		}
		log.Panic("Unknown opertaion received for o2 in ET")
	}
	log.Panic("Unknown opertaion received for o1 in ET")
	return nil
}

func LET(o *common.PeerOperation, ol []*common.PeerOperation) *common.PeerOperation {
	if len(ol) == 0 {
		return o
	}
	return LET(ET(o, ol[0]), Tail(ol))
}

func LIT(o *common.PeerOperation, ol []*common.PeerOperation) *common.PeerOperation {
	if len(ol) == 0 {
		return o
	}
	return LIT(IT(o, ol[0]), Tail(ol))
}

//returns all elements from index 1 to end
func Tail(ol []*common.PeerOperation) []*common.PeerOperation {
	if len(ol) < 2 {
		return make([]*common.PeerOperation, 0)
	}
	return ol[1:]
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
	return o.Position
}

func L(o *common.PeerOperation) int {
	return len(o.Character)
}

func S(o *common.PeerOperation) string {
	return o.Character
}

func Insert(content string, position int) *common.PeerOperation {
	return common.NewPeerOperation("i", content, position)
}

func Delete(length, position int) *common.PeerOperation {
	return common.NewPeerOperation("d", strconv.Itoa(length), position)
}
