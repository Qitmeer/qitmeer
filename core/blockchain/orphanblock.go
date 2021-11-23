package blockchain

import (
	"github.com/Qitmeer/qng-core/core/types"
	"time"
)

// orphanBlock represents a block that we don't yet have the parent for.  It
// is a normal block plus an expiration time to prevent caching the orphan
// forever.
type orphanBlock struct {
	block      *types.SerializedBlock
	expiration time.Time
	height     uint64
}

type orphanBlockSlice []*orphanBlock

func (ob orphanBlockSlice) Len() int {
	return len(ob)
}

func (ob orphanBlockSlice) Less(i, j int) bool {
	return ob[i].height < ob[j].height
}

func (bn orphanBlockSlice) Swap(i, j int) {
	bn[i], bn[j] = bn[j], bn[i]
}
