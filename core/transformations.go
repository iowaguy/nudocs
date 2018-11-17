package core

func InclusionTransformation(o1, o2 *PeerOperation) *PeerOperation {
	// TODO needs to determine exactly which function in the transformation matrix to call
	return nil
}

func ExclusionTransformation(o1, o2 *PeerOperation) *PeerOperation {
	// TODO needs to determine exactly which function in the transformation matrix to call
	return nil
}

func LET(eo *PeerOperation, l []*PeerOperation) *PeerOperation {
	for _, po := range l {
		eo = ExclusionTransformation(eo, po)
	}

	return eo
}

func LIT(eo *PeerOperation, l []*PeerOperation) *PeerOperation {
	for _, po := range l {
		eo = InclusionTransformation(eo, po)
	}

	return eo
}
