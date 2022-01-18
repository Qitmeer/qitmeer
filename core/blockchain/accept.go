// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/blockchain/token"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/params"
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
func (b *BlockChain) maybeAcceptBlock(block *types.SerializedBlock, flags BehaviorFlags) error {
	// This function should never be called with orphan blocks or the
	// genesis block.
	b.ChainLock()
	defer func() {
		b.ChainUnlock()
		b.flushNotifications()
	}()

	newNode := NewBlockNode(block, block.Block().Parents)
	mainParent := b.bd.GetMainParentByHashs(block.Block().Parents)
	if mainParent == nil {
		return fmt.Errorf("Can't find main parent\n")
	}
	// The block must pass all of the validation rules which depend on the
	// position of the block within the block chain.
	err := b.checkBlockContext(block, mainParent, flags)
	if err != nil {
		return err
	}

	// Prune stake nodes which are no longer needed before creating a new
	// node.
	b.pruner.pruneChainIfNeeded()

	//dag
	newOrders, oldOrders, ib, isMainChainTipChange := b.bd.AddBlock(newNode)
	if newOrders == nil || newOrders.Len() == 0 || ib == nil {
		return fmt.Errorf("Irreparable error![%s]\n", newNode.GetHash().String())
	}
	block.SetOrder(uint64(ib.GetOrder()))
	block.SetHeight(ib.GetHeight())

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
		exists, err := dbTx.HasBlock(block.Hash())
		if err != nil {
			return err
		}
		if exists {
			return nil
		}
		err = dbMaybeStoreBlock(dbTx, block)
		if err != nil {
			if database.IsError(err, database.ErrBlockExists) {
				return nil
			}
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
	_, err = b.connectDagChain(ib, block, newOrders, oldOrders)
	if err != nil {
		log.Warn(fmt.Sprintf("%s", err))
	}

	err = b.updateBestState(ib, block, newOrders)
	if err != nil {
		panic(err.Error())
	}
	b.ChainUnlock()
	// Notify the caller that the new block was accepted into the block
	// chain.  The caller would typically want to react by relaying the
	// inventory to other peers.
	b.sendNotification(BlockAccepted, &BlockAcceptedNotifyData{
		IsMainChainTipChange: isMainChainTipChange,
		Block:                block,
		Flags:                flags,
	})
	b.ChainLock()
	return nil
}

func (b *BlockChain) FastAcceptBlock(block *types.SerializedBlock, flags BehaviorFlags) error {
	b.ChainLock()
	defer func() {
		b.ChainUnlock()
		b.flushNotifications()
	}()

	newNode := NewBlockNode(block, block.Block().Parents)

	fastAdd := flags&BFFastAdd == BFFastAdd
	if !fastAdd {
		mainParent := b.bd.GetMainParentByHashs(block.Block().Parents)
		if mainParent == nil {
			return fmt.Errorf("Can't find main parent\n")
		}
		// The block must pass all of the validation rules which depend on the
		// position of the block within the block chain.
		err := b.checkBlockContext(block, mainParent, flags)
		if err != nil {
			return err
		}
	}
	//dag
	newOrders, oldOrders, ib, _ := b.bd.AddBlock(newNode)
	if newOrders == nil || newOrders.Len() == 0 || ib == nil {
		return fmt.Errorf("Irreparable error![%s]\n", newNode.GetHash().String())
	}

	block.SetOrder(uint64(ib.GetOrder()))
	block.SetHeight(ib.GetHeight())

	err := b.db.Update(func(dbTx database.Tx) error {
		if err := dbMaybeStoreBlock(dbTx, block); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	_, err = b.connectDagChain(ib, block, newOrders, oldOrders)
	if err != nil {
		log.Warn(fmt.Sprintf("%s", err))
	}

	return b.updateBestState(ib, block, newOrders)
}

func (b *BlockChain) updateTokenState(node blockdag.IBlock, block *types.SerializedBlock, rollback bool) error {
	if rollback {
		if uint32(node.GetID()) == b.TokenTipID {
			state := b.GetTokenState(b.TokenTipID)
			if state != nil {
				err := b.db.Update(func(dbTx database.Tx) error {
					return token.DBRemoveTokenState(dbTx, uint32(node.GetID()))
				})
				if err != nil {
					return err
				}
				b.TokenTipID = state.PrevStateID
			}

		}
		return nil
	}
	updates := []token.ITokenUpdate{}
	for _, tx := range block.Transactions() {
		if tx.IsDuplicate {
			log.Trace(fmt.Sprintf("updateTokenBalance skip duplicate tx %v", tx.Hash()))
			continue
		}

		if types.IsTokenTx(tx.Tx) {
			update, err := token.NewUpdateFromTx(tx.Tx)
			if err != nil {
				return err
			}
			updates = append(updates, update)
		}
	}
	if len(updates) <= 0 {
		return nil
	}
	state := b.GetTokenState(b.TokenTipID)
	if state == nil {
		state = &token.TokenState{PrevStateID: uint32(blockdag.MaxId), Updates: updates}
	} else {
		state.PrevStateID = b.TokenTipID
		state.Updates = updates
	}

	err := state.Update()
	if err != nil {
		return err
	}

	err = b.db.Update(func(dbTx database.Tx) error {
		return token.DBPutTokenState(dbTx, uint32(node.GetID()), state)
	})
	if err != nil {
		return err
	}
	b.TokenTipID = uint32(node.GetID())
	return state.Commit()
}

func (b *BlockChain) GetTokenState(bid uint32) *token.TokenState {
	var state *token.TokenState
	err := b.db.View(func(dbTx database.Tx) error {
		ts, err := token.DBFetchTokenState(dbTx, bid)
		if err != nil {
			return err
		}
		state = ts
		return nil
	})
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	return state
}

func (b *BlockChain) GetCurTokenState() *token.TokenState {
	b.ChainRLock()
	defer b.ChainRUnlock()
	return b.GetTokenState(b.TokenTipID)
}

func (b *BlockChain) GetCurTokenOwners(coinId types.CoinID) ([]byte, error) {
	b.ChainRLock()
	defer b.ChainRUnlock()
	state := b.GetTokenState(b.TokenTipID)
	if state == nil {
		return nil, fmt.Errorf("Token state error\n")
	}
	tt, ok := state.Types[coinId]
	if !ok {
		return nil, fmt.Errorf("It doesn't exist: Coin id (%d)\n", coinId)
	}
	return tt.Owners, nil
}

func (b *BlockChain) CheckTokenState(block *types.SerializedBlock) error {
	updates := []token.ITokenUpdate{}
	for _, tx := range block.Transactions() {
		if tx.IsDuplicate {
			log.Trace(fmt.Sprintf("updateTokenBalance skip duplicate tx %v", tx.Hash()))
			continue
		}

		if types.IsTokenTx(tx.Tx) {
			update, err := token.NewUpdateFromTx(tx.Tx)
			if err != nil {
				return err
			}
			updates = append(updates, update)
		}
	}
	if len(updates) <= 0 {
		return nil
	}
	state := b.GetTokenState(b.TokenTipID)
	if state == nil {
		state = &token.TokenState{PrevStateID: uint32(blockdag.MaxId), Updates: updates}
	} else {
		state.PrevStateID = b.TokenTipID
		state.Updates = updates
	}
	return state.Update()
}

func (b *BlockChain) IsValidTxType(tt types.TxType) bool {
	txTypesCfg := types.StdTxs
	ok, err := b.isDeploymentActive(params.DeploymentToken)
	if err == nil && ok && len(types.NonStdTxs) > 0 {
		txTypesCfg = append(txTypesCfg, types.NonStdTxs...)
	}

	for _, txt := range txTypesCfg {
		if txt == tt {
			return true
		}
	}
	return false
}
