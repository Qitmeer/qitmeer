/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package blockdag

// BlockPrioritySlice is used to sort dag block
// Just for inside
type BlockPrioritySlice []IBlock

func (bn BlockPrioritySlice) Len() int {
	return len(bn)
}

func (bn BlockPrioritySlice) Less(i, j int) bool {
	if bn[i].(*PhantomBlock).blueNum < bn[j].(*PhantomBlock).blueNum {
		return true
	} else if bn[i].(*PhantomBlock).blueNum == bn[j].(*PhantomBlock).blueNum {
		if bn[i].GetData().GetPriority() < bn[j].GetData().GetPriority() {
			return true
		} else if bn[i].GetData().GetPriority() == bn[j].GetData().GetPriority() {
			if bn[i].GetHash().String() > bn[j].GetHash().String() {
				return true
			}
		}
	}
	return false
}

func (bn BlockPrioritySlice) Swap(i, j int) {
	bn[i], bn[j] = bn[j], bn[i]
}
