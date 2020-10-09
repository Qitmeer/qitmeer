/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package p2p

import (
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
)

type BlockDataSlice []*pb.BlockData

func (bd BlockDataSlice) Len() int {
	return len(bd)
}

func (bd BlockDataSlice) Less(i, j int) bool {
	return bd[i].DagID < bd[j].DagID
}

func (bd BlockDataSlice) Swap(i, j int) {
	bd[i], bd[j] = bd[j], bd[i]
}
