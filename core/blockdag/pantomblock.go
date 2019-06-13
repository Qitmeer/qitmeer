package blockdag

type PhantomBlock struct {
	*Block
	blueNum uint

	blueDiffAnticone *HashSet
	redDiffAnticone *HashSet
}

func (pb *PhantomBlock) IsBluer(other *PhantomBlock) bool {
	if pb.blueNum > other.blueNum ||
		(pb.blueNum == other.blueNum && pb.GetHash().String() < other.GetHash().String()) {
		return true
	}
	return false
}