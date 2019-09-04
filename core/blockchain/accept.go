// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"
	"github.com/Qitmeer/qitmeer-lib/core/types"
	"github.com/Qitmeer/qitmeer-lib/engine/txscript"
	"github.com/Qitmeer/qitmeer/database"
	"math"
	"time"
)

// checkCoinbaseUniqueHeight checks to ensure that for all blocks height > 1 the
// coinbase contains the height encoding to make coinbase hash collisions
// impossible.
func checkCoinbaseUniqueHeight(blockHeight uint64, block *types.SerializedBlock) error {
	// check height
	serializedHeight, err := ExtractCoinbaseHeight(block.Block().Transactions[0])
	if err != nil {
		return err
	}
	if uint64(serializedHeight) != blockHeight {
		str := fmt.Sprintf("the coinbase signature script serialized "+
			"block height is %d when %d was expected",
			serializedHeight, blockHeight)
		return ruleError(ErrCoinbaseHeight, str)
	}
	return nil
}

// IsFinalizedTransaction determines whether or not a transaction is finalized.
func IsFinalizedTransaction(tx *types.Tx, blockHeight uint64, blockTime time.Time) bool {
	// Lock time of zero means the transaction is finalized.
	msgTx := tx.Transaction()
	lockTime := msgTx.LockTime
	if lockTime == 0 {
		return true
	}

	// The lock time field of a transaction is either a block height at
	// which the transaction is finalized or a timestamp depending on if the
	// value is before the txscript.LockTimeThreshold.  When it is under the
	// threshold it is a block height.
	var blockTimeOrHeight int64
	if lockTime < txscript.LockTimeThreshold {
		//TODO, remove the type conversion
		blockTimeOrHeight = int64(blockHeight)
	} else {
		blockTimeOrHeight = blockTime.Unix()
	}
	if int64(lockTime) < blockTimeOrHeight {
		return true
	}

	// At this point, the transaction's lock time hasn't occurred yet, but
	// the transaction might still be finalized if the sequence number
	// for all transaction inputs is maxed out.
	for _, txIn := range msgTx.TxIn {
		if txIn.Sequence != math.MaxUint32 {
			return false
		}
	}
	return true
}

// maybeAcceptBlock potentially accepts a block into the block chain and, if
// accepted, returns the length of the fork the block extended.  It performs
// several validation checks which depend on its position within the block chain
// before adding it.  The block is expected to have already gone through
// ProcessBlock before calling this function with it.  In the case the block
// extends the best chain or is now the tip of the best chain due to causing a
// reorganize, the fork length will be 0.
//
// The flags are also passed to checkBlockContext and connectBestChain.  See
// their documentation for how the flags modify their behavior.
//
// This function MUST be called with the chain state lock held (for writes).
func (b *BlockChain) maybeAcceptBlock(block *types.SerializedBlock, flags BehaviorFlags) (bool, error) {
	// This function should never be called with orphan blocks or the
	// genesis block.

	parentsNode := []*blockNode{}
	for _, pb := range block.Block().Parents {
		prevHash := pb
		prevNode := b.index.LookupNode(prevHash)
		if prevNode == nil {
			str := fmt.Sprintf("Parents block %s is unknown", prevHash)
			log.Debug(str)
			return false, nil
		}
		parentsNode = append(parentsNode, prevNode)
	}

	err:=b.bd.CheckLayerGap(block.Block().Parents)
	if err != nil {
		return false,err
	}

	blockHeader := &block.Block().Header
	newNode := newBlockNode(blockHeader, parentsNode)
	mainParent:=newNode.GetMainParent(b)
	if mainParent == nil {
		return false,fmt.Errorf("Can't find main parent")
	}

	newNode.SetHeight(mainParent.GetHeight()+1)

	block.SetHeight(newNode.GetHeight())
	// The block must pass all of the validation rules which depend on the
	// position of the block within the block chain.
	err = b.checkBlockContext(block,mainParent, flags)
	if err != nil {
		return false, err
	}

	// Prune stake nodes which are no longer needed before creating a new
	// node.
	b.pruner.pruneChainIfNeeded()

	//dag
	newOrders := b.bd.AddBlock(newNode)
	if newOrders == nil || newOrders.Len() == 0 {
		return false, fmt.Errorf("Irreparable error![%s]", newNode.hash.String())
	}
	oldOrders:=BlockNodeList{}
	b.getReorganizeNodes(newNode,block,newOrders,&oldOrders)
	b.index.AddNode(newNode)
	b.index.SetStatusFlags(newNode, statusDataStored)
	err = b.index.flushToDB(b.bd)
	if err != nil {
		panic(err.Error())
	}
	// Insert the block into the database if it's not already there.  Even
	// though it is possible the block will ultimately fail to connect, it
	// has already passed all proof-of-work and validity tests which means
	// it would be prohibitively expensive for an attacker to fill up the
	// disk with a bunch of blocks that fail to connect.  This is necessary
	// since it allows block download to be decoupled from the much more
	// expensive connection logic.  It also has some other nice properties
	// such as making blocks that never become part of the main chain or
	// blocks that fail to connect available for further analysis.
	//
	// Also, store the associated block index entry.
	err = b.db.Update(func(dbTx database.Tx) error {
		if err := dbMaybeStoreBlock(dbTx, block); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		panic(err.Error())
	}

	// Connect the passed block to the chain while respecting proper chain
	// selection according to the chain with the most proof of work.  This
	// also handles validation of the transaction scripts.
	_, err = b.connectDagChain(newNode, block, newOrders,oldOrders)
	if err != nil {
		log.Warn(fmt.Sprintf("%s",err))
	}
	b.updateBestState(newNode, block)
	// Notify the caller that the new block was accepted into the block
	// chain.  The caller would typically want to react by relaying the
	// inventory to other peers.
	b.chainLock.Unlock()
	//TODO, refactor to event subscript/publish
	b.sendNotification(BlockAccepted, &BlockAcceptedNotifyData{
		ForkLen:    0,
		Block:      block,
	})
	b.chainLock.Lock()

	err = b.index.flushToDB(b.bd)
	if err != nil {
		return true, nil
	}

	return true, nil
}
