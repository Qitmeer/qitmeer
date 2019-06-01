package blockdag

import "qitmeer/common/hash"

type PhantomBlock struct {
	*Block
	blueNum uint
	coloringParent *hash.Hash

	blueDiffPastOrder *HashSet
	redDiffPastOrder *HashSet
}

func (pb *PhantomBlock) IsBluer(other *PhantomBlock) bool {
	if pb.blueNum > other.blueNum ||
		(pb.blueNum == other.blueNum && pb.GetHash().String() < other.GetHash().String()) {
		return true
	}
	return false
}