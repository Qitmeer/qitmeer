package miner

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"github.com/Qitmeer/qitmeer/services/mining"
	"time"
)

func (m *CPUMiner) solveX8r16Block(msgBlock *types.Block, ticker *time.Ticker, quit chan struct{}) bool {

	// TODO, decided if need extra nonce for coinbase-tx
	// Choose a random extra nonce offset for this block template and
	// worker.
	/*
		enOffset, err := s.RandomUint64()
		if err != nil {
			log.Error("Unexpected error while generating random "+
				"extra nonce offset: %v", err)
			enOffset = 0
		}
	*/

	// Create a couple of convenience variables.
	header := &msgBlock.Header

	// Initial state.
	lastGenerated := time.Now()
	lastTxUpdate := m.txSource.LastUpdated()
	hashesCompleted := uint64(0)
	target := pow.CompactToBig(uint32(header.Difficulty))
	// TODO, decided if need extra nonce for coinbase-tx
	// Note that the entire extra nonce range is iterated and the offset is
	// added relying on the fact that overflow will wrap around 0 as
	// provided by the Go spec.
	// for extraNonce := uint64(0); extraNonce < maxExtraNonce; extraNonce++ {

	// Update the extra nonce in the block template with the
	// new value by regenerating the coinbase script and
	// setting the merkle root to the new value.
	// TODO, decided if need extra nonce for coinbase-tx
	// updateExtraNonce(msgBlock, extraNonce+enOffset)

	// Update the extra nonce in the block template header with the
	// new value.
	// binary.LittleEndian.PutUint64(header.ExtraData[:], extraNonce+enOffset)

	// Search through the entire nonce range for a solution while
	// periodically checking for early quit and stale block
	// conditions along with updates to the speed monitor.
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
				time.Now().After(lastGenerated.Add(3*time.Second))) ||
				time.Now().After(lastGenerated.Add(60*time.Second)) {

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
		instance := pow.GetInstance(pow.X8R16, 0, []byte{})
		powStruct := instance.(*pow.X8r16)
		// Update the nonce and hash the block header.
		powStruct.Nonce = i

		header.Pow = powStruct

		// Each hash is actually a double hash (tow hashes), so
		// increment the number of hashes by 2
		hashesCompleted += 2
		h := hash.HashX8r16(header.BlockData())
		hashNum := pow.HashToBig(&h)

		if hashNum.Cmp(target) <= 0 {
			// The block is solved when the new block hash is less
			// than the target difficulty.  Yay!
			m.updateHashes <- hashesCompleted
			return true
		}
	}
	//}
	return false
}
