package core

import "github.com/iowaguy/nudocs/common"

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
