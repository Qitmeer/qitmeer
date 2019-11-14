package blockchain

// BlockNodeList is used to sort blockNode
type BlockNodeList []*blockNode

func (bn BlockNodeList) Len() int {
	return len(bn)
}

func (bn BlockNodeList) Less(i, j int) bool {
	return bn[i].order < bn[j].order
}

func (bn BlockNodeList) Swap(i, j int) {
	bn[i], bn[j] = bn[j], bn[i]
}
