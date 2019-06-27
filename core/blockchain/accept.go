// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"encoding/binary"
	"fmt"
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer/core/types"
	"github.com/HalalChain/qitmeer/database"
	"github.com/HalalChain/qitmeer/engine/txscript"
	"math"
	"time"
)

// checkCoinbaseUniqueHeight checks to ensure that for all blocks height > 1 the
// coinbase contains the height encoding to make coinbase hash collisions
// impossible.
func checkCoinbaseUniqueHeight(blockHeight uint64, block *types.SerializedBlock) error {
	// Coinbase TxOut[0] is always tax, TxOut[1] is always
	// height + extranonce, so at least two outputs must
	// exist.
	if len(block.Block().Transactions[0].TxOut) < 2 {
		str := fmt.Sprintf("block %v is missing necessary coinbase "+
			"outputs", block.Hash())
		return ruleError(ErrFirstTxNotCoinbase, str)
	}

	// Only version 0 scripts are currently valid.
	nullDataOut := block.Block().Transactions[0].TxOut[1]
	// TODO, revisit version & check should go to validation
	/*
		if nullDataOut.Version != 0 {
			str := fmt.Sprintf("block %v output 1 has wrong script version",
				block.Hash())
			return ruleError(ErrFirstTxNotCoinbase, str)
		}
	*/

	// The first 4 bytes of the null data output must be the encoded height
	// of the block, so that every coinbase created has a unique transaction
	// hash.
	nullData, err := txscript.ExtractCoinbaseNullData(nullDataOut.PkScript)
	if err != nil {
		str := fmt.Sprintf("block %v output 1 has wrong script type",
			block.Hash())
		return ruleError(ErrFirstTxNotCoinbase, str)
	}
	if len(nullData) < 4 {
		str := fmt.Sprintf("block %v output 1 data push too short to "+
			"contain height", block.Hash())
		return ruleError(ErrFirstTxNotCoinbase, str)
	}

	// Check the height and ensure it is correct.
	cbHeight := binary.LittleEndian.Uint32(nullData[0:4])
	if cbHeight < uint32(blockHeight) {
		prevBlock := block.Block().Header.ParentRoot
		str := fmt.Sprintf("block %v output 1 has wrong order in "+
			"coinbase; want %v, got %v; prevBlock %v, header order %v",
			block.Hash(), blockHeight, cbHeight, prevBlock,
			block.Order())
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
	blockHeader := &block.Block().Header
	newNode := newBlockNode(blockHeader, parentsNode)
	// The block must pass all of the validation rules which depend on the
	// position of the block within the block chain.
	err := b.checkBlockContext(block, newNode.GetBackParent(), flags)
	if err != nil {
		return false, err
	}

	// Prune stake nodes which are no longer needed before creating a new
	// node.
	b.pruner.pruneChainIfNeeded()

	//dag
	list := b.bd.AddBlock(newNode)
	if list == nil || list.Len() == 0 {
		log.Debug(fmt.Sprintf("Irreparable error![%s]", newNode.hash.String()))
		return false, nil
	}
	b.index.addNode(newNode)
	//
	for e := list.Front(); e != nil; e = e.Next() {
		refHash := e.Value.(*hash.Hash)
		refblock := b.bd.GetBlock(refHash)
		refnode := b.index.lookupNode(refHash)
		refnode.SetOrder(uint64(refblock.GetOrder()))

		if newNode.GetHash().IsEqual(refHash) {
			block.SetOrder(uint64(refblock.GetOrder()))
		}
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

		if err := dbPutBlockNode(dbTx, newNode); err != nil {
			return err
		}
		b.index.SetStatusFlags(newNode, statusDataStored)
		return nil
	})
	if err != nil {
		b.index.SetStatusFlags(newNode, statusValidateFailed)
		return false, err
	}

	// Connect the passed block to the chain while respecting proper chain
	// selection according to the chain with the most proof of work.  This
	// also handles validation of the transaction scripts.
	success, err := b.connectDagChain(newNode, block, list)
	if !success || err != nil {
		b.index.SetStatusFlags(newNode, statusValidateFailed)
		return false, err
	}

	b.index.SetStatusFlags(newNode, statusValid)
	// Notify the caller that the new block was accepted into the block
	// chain.  The caller would typically want to react by relaying the
	// inventory to other peers.
	b.chainLock.Unlock()

	b.sendNotification(BlockConnected, []*types.SerializedBlock{block})

	//TODO, refactor to event subscript/publish
	b.sendNotification(BlockAccepted, &BlockAcceptedNotifyData{
		BestHeight: block.Order(),
		ForkLen:    0,
		Block:      block,
	})
	b.chainLock.Lock()

	return true, nil
}
