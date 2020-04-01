package blockdag

// BlockSlice is used to sort dag block
// Just for outside
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

// BlockSlice is used to sort dag block
// Just for inside
type BlockHashSlice []IBlock

func (bn BlockHashSlice) Len() int {
	return len(bn)
}

func (bn BlockHashSlice) Less(i, j int) bool {
	return bn[i].GetHash().String() < bn[j].GetHash().String()
}

func (bn BlockHashSlice) Swap(i, j int) {
	bn[i], bn[j] = bn[j], bn[i]
}
