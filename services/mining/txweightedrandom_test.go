package mining

import (
	"fmt"
	"testing"
)

func Test_TXWeightedRandom(t *testing.T) {
	const reserve = 10
	itemQueue := newWeightedRandQueue(reserve)
	for i := 0; i < reserve; i++ {
		item := &WeightedRandTx{fee: int64(i)}
		itemQueue.Push(item)
	}

	for itemQueue.Len() > 0 {
		item := itemQueue.Pop()
		fmt.Println(item.fee)
	}
}
