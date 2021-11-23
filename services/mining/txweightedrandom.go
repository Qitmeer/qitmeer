package mining

import (
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/common/roughtime"
	"github.com/Qitmeer/qng-core/core/types"
	"math/rand"
)

// weighted random tx
type WeightedRandTx struct {
	tx       *types.Tx
	fee      int64
	feePerKB int64

	dependsOn map[hash.Hash]struct{}
}

// The Queue for weighted rand tx
type WeightedRandQueue struct {
	totalFee int64
	items    []*WeightedRandTx
}

// The length of WeightedRandQueue
func (wq *WeightedRandQueue) Len() int {
	return len(wq.items)
}

// Push item to WeightedRandQueue
func (wq *WeightedRandQueue) Push(tx *WeightedRandTx) {
	wq.items = append(wq.items, tx)
	wq.totalFee += tx.fee + 1
}

// Pop item from WeightedRandQueue
func (wq *WeightedRandQueue) Pop() *WeightedRandTx {
	if wq.Len() <= 0 {
		return nil
	}
	factor := rand.Int63n(wq.totalFee)

	total := int64(0)
	index := int(0)
	var item *WeightedRandTx
	for index, item = range wq.items {
		total += item.fee
		if total >= factor {
			break
		}
	}
	wq.items = append(wq.items[:index], wq.items[index+1:]...)
	//total = total - item.fee - 1

	return item
}

// Build WeightedRandQueue
func newWeightedRandQueue(reserve int) *WeightedRandQueue {
	rand.Seed(roughtime.Now().Unix())
	wq := &WeightedRandQueue{
		items: make([]*WeightedRandTx, 0, reserve),
	}
	return wq
}
