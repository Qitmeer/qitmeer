package blockchain

import (
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/common/roughtime"
	"github.com/Qitmeer/qng-core/meerdag"
	"github.com/Qitmeer/qng-core/core/types"
	"math"
	"sort"
	"time"
)

const (
	MaxOrphanStallDuration = 10 * time.Minute
)

// IsKnownOrphan returns whether the passed hash is currently a known orphan.
// Keep in mind that only a limited number of orphans are held onto for a
// limited amount of time, so this function must not be used as an absolute
// way to test if a block is an orphan block.  A full block (as opposed to just
// its hash) must be passed to ProcessBlock for that purpose.  However, calling
// ProcessBlock with an orphan that already exists results in an error, so this
// function provides a mechanism for a caller to intelligently detect *recent*
// duplicate orphans and react accordingly.
//
// This function is safe for concurrent access.
func (b *BlockChain) IsOrphan(hash *hash.Hash) bool {
	// Protect concurrent access.  Using a read lock only so multiple
	// readers can query without blocking each other.
	b.orphanLock.RLock()
	exists := b.isOrphan(hash)
	b.orphanLock.RUnlock()

	return exists
}

func (b *BlockChain) isOrphan(hash *hash.Hash) bool {
	_, exists := b.orphans[*hash]
	return exists
}

// Whether it is connected by all parents
func (b *BlockChain) IsUnconnectedOrphan(hash *hash.Hash) bool {
	op := b.GetRecentOrphanParents(hash)
	return len(op) > 0
}

// GetOrphansParents returns the parents for the provided hash from the
// map of orphan blocks.
func (b *BlockChain) GetOrphansParents() []*hash.Hash {
	b.orphanLock.RLock()
	defer b.orphanLock.RUnlock()
	//
	result := meerdag.NewHashSet()
	for _, v := range b.orphans {
		for _, h := range v.block.Block().Parents {
			if b.bd.HasBlock(h) || b.isOrphan(h) {
				continue
			}
			result.Add(h)
		}

	}
	return result.List()
}

// GetOrphansParents returns the parents for the provided hash from the
// map of orphan blocks.
func (b *BlockChain) GetRecentOrphanParents(h *hash.Hash) []*hash.Hash {
	b.orphanLock.RLock()
	defer b.orphanLock.RUnlock()
	//
	ob := b.getOrphan(h)
	if ob == nil {
		return nil
	}
	result := meerdag.NewHashSet()
	for _, h := range ob.Block().Parents {
		if b.bd.HasBlock(h) || b.isOrphan(h) {
			continue
		}
		result.Add(h)
	}

	return result.List()
}

// Get the total of all orphans
func (b *BlockChain) GetOrphansTotal() int {
	b.orphanLock.RLock()
	ol := len(b.orphans)
	b.orphanLock.RUnlock()
	return ol
}

func (b *BlockChain) GetRecentOrphansParents() []*hash.Hash {
	b.orphanLock.RLock()
	defer b.orphanLock.RUnlock()

	result := meerdag.NewHashSet()
	mh := b.BestSnapshot().GraphState.GetMainHeight()
	for _, v := range b.orphans {
		for _, h := range v.block.Block().Parents {
			if len(b.orphans) >= MaxOrphanBlocks {
				dist := math.Abs(float64(v.height) - float64(mh))
				if dist > float64(meerdag.StableConfirmations) {
					continue
				}
			}
			if b.bd.HasBlock(h) || b.isOrphan(h) {
				continue
			}
			result.Add(h)
		}

	}
	return result.List()
}

func (b *BlockChain) IsOrphanOK(serializedHeight uint64) bool {
	dist := serializedHeight + meerdag.StableConfirmations*2
	return uint(dist) >= b.BestSnapshot().GraphState.GetMainHeight()
}

// removeOrphanBlock removes the passed orphan block from the orphan pool and
// previous orphan index.
func (b *BlockChain) RemoveOrphanBlock(orphan *orphanBlock) {
	// Protect concurrent access.
	b.orphanLock.Lock()
	defer b.orphanLock.Unlock()
	b.removeOrphanBlock(orphan)
}

func (b *BlockChain) removeOrphanBlock(orphan *orphanBlock) {
	// Remove the orphan block from the orphan pool.
	orphanHash := orphan.block.Hash()
	delete(b.orphans, *orphanHash)
}

// addOrphanBlock adds the passed block (which is already determined to be
// an orphan prior calling this function) to the orphan pool.  It lazily cleans
// up any expired blocks so a separate cleanup poller doesn't need to be run.
// It also imposes a maximum limit on the number of outstanding orphan
// blocks and will remove the oldest received orphan block if the limit is
// exceeded.
func (b *BlockChain) addOrphanBlock(block *types.SerializedBlock) {
	serializedHeight, err := ExtractCoinbaseHeight(block.Block().Transactions[0])
	if err != nil {
		return
	}
	if !b.IsOrphanOK(serializedHeight) {
		return
	}
	// Protect concurrent access.  This is intentionally done here instead
	// of near the top since removeOrphanBlock does its own locking and
	// the range iterator is not invalidated by removing map entries.
	b.orphanLock.Lock()
	defer b.orphanLock.Unlock()

	b.refreshOrphans()
	// Limit orphan blocks to prevent memory exhaustion.
	if len(b.orphans)+1 > MaxOrphanBlocks*2 {
		// Remove the oldest orphan to make room for the new one.
		b.removeOrphanBlock(b.oldestOrphan)
		b.oldestOrphan = nil
	}

	// Insert the block into the orphan map with an expiration time
	// 1 hour from now.
	expiration := roughtime.Now().Add(MaxOrphanStallDuration)
	oBlock := &orphanBlock{
		block:      block,
		expiration: expiration,
		height:     serializedHeight,
	}
	b.orphans[*block.Hash()] = oBlock
}

// processOrphans determines if there are any orphans which depend on the passed
// block hash (they are no longer orphans if true) and potentially accepts them.
// It repeats the process for the newly accepted blocks (to detect further
// orphans which may no longer be orphans) until there are no more.
//
// The flags do not modify the behavior of this function directly, however they
// are needed to pass along to maybeAcceptBlock.
//
// This function MUST be called with the chain state lock held (for writes).
func (b *BlockChain) processOrphans(flags BehaviorFlags) error {
	if len(b.orphans) <= 0 {
		return nil
	}
	queue := orphanBlockSlice{}
	for _, v := range b.orphans {
		queue = append(queue, v)
	}
	if len(queue) >= 2 {
		sort.Sort(queue)
	}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		exists := b.bd.HasBlock(cur.block.Hash())
		if exists {
			b.RemoveOrphanBlock(cur)
			continue
		}

		allExists := true
		for _, h := range cur.block.Block().Parents {
			exists := b.bd.HasBlock(h)
			if !exists {
				allExists = false
			}

		}
		if !allExists {
			continue
		}
		b.RemoveOrphanBlock(cur)
		b.maybeAcceptBlock(cur.block, flags)
	}
	return nil
}

func (b *BlockChain) GetOrphan(hash *hash.Hash) *types.SerializedBlock {
	b.orphanLock.RLock()
	orphan := b.getOrphan(hash)
	b.orphanLock.RUnlock()
	return orphan
}

func (b *BlockChain) getOrphan(hash *hash.Hash) *types.SerializedBlock {
	orphan, exists := b.orphans[*hash]
	if !exists {
		return nil
	}
	return orphan.block
}

func (b *BlockChain) RefreshOrphans() error {
	b.orphanLock.Lock()
	b.refreshOrphans()
	b.orphanLock.Unlock()

	return b.processOrphans(BFP2PAdd)
}

func (b *BlockChain) refreshOrphans() {
	// Remove expired orphan blocks.
	for _, oBlock := range b.orphans {
		if roughtime.Now().After(oBlock.expiration) {
			b.removeOrphanBlock(oBlock)
			continue
		}
		if !b.IsOrphanOK(oBlock.height) {
			b.removeOrphanBlock(oBlock)
			continue
		}
		// Update the oldest orphan block pointer so it can be discarded
		// in case the orphan pool fills up.
		if b.oldestOrphan == nil ||
			oBlock.expiration.Before(b.oldestOrphan.expiration) {
			b.oldestOrphan = oBlock
		}
	}
}
