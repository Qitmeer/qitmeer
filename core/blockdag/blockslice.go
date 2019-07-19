package blockdag

// BlockSlice is used to sort dag block
type BlockSlice []IBlock

func (bn BlockSlice) Len() int {
	return len(bn)
}

func (bn BlockSlice) Less(i, j int) bool {
	return bn[i].GetID() < bn[j].GetID()
}

func (bn BlockSlice) Swap(i, j int) {
	bn[i], bn[j] = bn[j], bn[i]
}