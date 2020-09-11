package miner

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"github.com/Qitmeer/qitmeer/services/mining"
	"time"
)

func (m *CPUMiner) solveQitmeerKeccak256Block(msgBlock *types.Block, ticker *time.Ticker, quit chan struct{}) bool {

	header := &msgBlock.Header
	// Initial state.
	lastGenerated := roughtime.Now()
	lastTxUpdate := m.txSource.LastUpdated()
	hashesCompleted := uint64(0)
	target := pow.CompactToBig(uint32(header.Difficulty))

	// Search through the entire nonce range for a solution while
	for i := uint32(0); i <= maxNonce; i++ {
		select {
		case <-quit:
			return false

		case <-ticker.C:
			m.updateHashes <- hashesCompleted
			hashesCompleted = 0

			// The current block is stale if the memory pool
			// has been updated since the block template was
			// generated and it has been at least 3 seconds,
			// or if it's been one minute.
			if (lastTxUpdate != m.txSource.LastUpdated() &&
				roughtime.Now().After(lastGenerated.Add(3*time.Second))) ||
				roughtime.Now().After(lastGenerated.Add(60*time.Second)) {

				return false
			}

			err := mining.UpdateBlockTime(msgBlock, m.blockManager.GetChain(), m.timeSource, m.params)
			if err != nil {
				log.Warn("CPU miner unable to update block template "+
					"time: %v", err)
				return false
			}

		default:
			// Non-blocking select to fall through
		}
		//Initial pow instance
		instance := pow.GetInstance(pow.QITMEERKECCAK256, 0, []byte{})
		powStruct := instance.(*pow.QitmeerKeccak256)
		// Update the nonce and hash the block header.
		powStruct.Nonce = i

		header.Pow = powStruct

		// Each hash is actually a double hash (tow hashes), so
		// increment the number of hashes by 2
		hashesCompleted += 2
		h := hash.HashQitmeerKeccak256(header.BlockData())
		hashNum := pow.HashToBig(&h)

		if hashNum.Cmp(target) <= 0 {
			// The block is solved when the new block hash is less
			// than the target difficulty.  Yay!
			m.updateHashes <- hashesCompleted
			return true
		}
	}
	return false
}
