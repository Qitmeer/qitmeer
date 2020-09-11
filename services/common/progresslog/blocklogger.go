// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2016 The btcsuite developers
// Copyright (c) 2016-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package progresslog

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/log"
	"sync"
	"time"
)

// BlockProgressLogger provides periodic logging for other services in order
// to show users progress of certain "actions" involving some or all current
// blocks. Ex: syncing to best chain, indexing all blocks, etc.
type BlockProgressLogger struct {
	receivedLogBlocks int64
	receivedLogTx     int64
	lastBlockLogTime  time.Time

	subsystemLogger log.Logger
	progressAction  string
	sync.Mutex
}

// NewBlockProgressLogger returns a new block progress logger.
// The progress message is templated as follows:
//  {progressAction} {numProcessed} {blocks|block} in the last {timePeriod}
//  ({numTxs}, height {lastBlockHeight}, {lastBlockTimeStamp})
func NewBlockProgressLogger(progressMessage string, logger log.Logger) *BlockProgressLogger {
	return &BlockProgressLogger{
		lastBlockLogTime: roughtime.Now(),
		progressAction:   progressMessage,
		subsystemLogger:  logger,
	}
}

// LogBlockHeight logs a new block height as an information message to show
// progress to the user. In order to prevent spam, it limits logging to one
// message every 10 seconds with duration and totals included.
func (b *BlockProgressLogger) LogBlockHeightByParent(block, parent *types.SerializedBlock) {
	b.Lock()
	defer b.Unlock()
	b.receivedLogBlocks++
	b.receivedLogTx += int64(len(parent.Transactions()))

	now := roughtime.Now()
	duration := now.Sub(b.lastBlockLogTime)
	if duration < time.Second*10 {
		return
	}

	// Truncate the duration to 10s of milliseconds.
	durationMillis := int64(duration / time.Millisecond)
	tDuration := 10 * time.Millisecond * time.Duration(durationMillis/10)

	// Log information about new block height.
	blockStr := "blocks"
	if b.receivedLogBlocks == 1 {
		blockStr = "block"
	}
	txStr := "transactions"
	if b.receivedLogTx == 1 {
		txStr = "transaction"
	}

	b.subsystemLogger.Info("%s %d %s in the last %s (%d %s, height %d, %s)",
		b.progressAction, b.receivedLogBlocks, blockStr, tDuration,
		b.receivedLogTx, txStr, block.Height(),
		block.Block().Header.Timestamp)

	b.receivedLogBlocks = 0
	b.receivedLogTx = 0
	b.lastBlockLogTime = now
}

// logBlockHeight logs a new block height as an information message to show
// progress to the user. In order to prevent spam, it limits logging to one
// message every 10 seconds with duration and totals included.
func (b *BlockProgressLogger) LogBlockHeight(block *types.SerializedBlock) {
	b.Lock()
	defer b.Unlock()
	b.receivedLogBlocks++
	b.receivedLogTx += int64(len(block.Block().Transactions))
	now := roughtime.Now()
	duration := now.Sub(b.lastBlockLogTime)
	if duration < time.Second*10 {
		return
	}

	// Truncate the duration to 10s of milliseconds.
	durationMillis := int64(duration / time.Millisecond)
	tDuration := 10 * time.Millisecond * time.Duration(durationMillis/10)

	// Log information about new block height.
	blockStr := "blocks"
	if b.receivedLogBlocks == 1 {
		blockStr = "block"
	}
	txStr := "transactions"
	if b.receivedLogTx == 1 {
		txStr = "transaction"
	}

	b.subsystemLogger.Info(fmt.Sprintf("%s %d %s in the last %s (%d %s, order %s, %s)",
		b.progressAction, b.receivedLogBlocks, blockStr, tDuration,
		b.receivedLogTx, txStr, blockdag.GetOrderLogStr(uint(block.Order())), block.Block().Header.Timestamp))

	b.receivedLogBlocks = 0
	b.receivedLogTx = 0
	b.lastBlockLogTime = now
}

func (b *BlockProgressLogger) SetLastLogTime(time time.Time) {
	b.lastBlockLogTime = time
}
