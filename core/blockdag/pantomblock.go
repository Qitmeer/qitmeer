package blockdag

import "github.com/HalalChain/qitmeer-lib/core/dag"

type PhantomBlock struct {
	*Block
	blueNum uint

	blueDiffAnticone *dag.HashSet
	redDiffAnticone *dag.HashSet
}

func (pb *PhantomBlock) IsBluer(other *PhantomBlock) bool {
	if pb.blueNum > other.blueNum ||
		(pb.blueNum == other.blueNum && pb.GetHash().String() < other.GetHash().String()) {
		return true
	}
	return false
}