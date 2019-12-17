// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/merkle"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/params"
	"math"
	"time"
)

// This function only differs from IsExpired in that it works with a raw wire
// transaction as opposed to a higher level util transaction.
func IsExpiredTx(tx *types.Transaction, blockHeight uint64) bool {
	expiry := tx.Expire
	return expiry != types.NoExpiryValue && blockHeight >= uint64(expiry) //TODO, remove type conversion
}

// IsExpired returns where or not the passed transaction is expired according to
// the given block height.
//
// This function only differs from IsExpiredTx in that it works with a higher
// level util transaction as opposed to a raw wire transaction.
func IsExpired(tx *types.Tx, blockHeight uint64) bool {
	return IsExpiredTx(tx.Transaction(), blockHeight)
}

// checkBlockSanity performs some preliminary checks on a block to ensure it is
// sane before continuing with block processing.  These checks are context
// free.
//
// The flags do not modify the behavior of this function directly, however they
// are needed to pass along to checkBlockHeaderSanity.
func (b *BlockChain) checkBlockSanity(block *types.SerializedBlock, timeSource MedianTimeSource, flags BehaviorFlags, chainParams *params.Params) error {
	msgBlock := block.Block()
	header := &msgBlock.Header

	// TODO It can be considered to delete in the future when it is officially launched
	if header.Version != b.BlockVersion {
		return ruleError(ErrBlockVersionTooOld, "block version too old")
	}

	err := checkBlockHeaderSanity(header, timeSource, flags, chainParams,block.Height())
	if err != nil {
		return err
	}
	// A block must have at least one parent.
	numPb := len(msgBlock.Parents)
	if numPb == 0 {
		return ruleError(ErrNoParents, "block does not contain "+
			"any parent")
	}

	// A block must not have more parents than the max block payload or
	// else it is certainly over the weight limit.
	if numPb > types.MaxParentsPerBlock {
		str := fmt.Sprintf("block contains too many parents - "+
			"got %d, max %d", numPb, types.MaxParentsPerBlock)
		return ruleError(ErrBlockTooBig, str)
	}
	// Build the block parents merkle tree and ensure the calculated merkle
	// parents root matches the entry in the block header.
	// This also has the effect of caching all
	// of the parents hashes in the block to speed up future hash
	// checks.  The tree here and checks the merkle root
	// after the following checks, but there is no reason not to check the
	// merkle root matches here.
	paMerkles := merkle.BuildParentsMerkleTreeStore(msgBlock.Parents)
	paMerkleRoot := paMerkles[len(paMerkles)-1]
	if !header.ParentRoot.IsEqual(paMerkleRoot) {
		str := fmt.Sprintf("block parents merkle root is invalid - block "+
			"header indicates %v, but calculated value is %v",
			&header.ParentRoot, paMerkleRoot)
		return ruleError(ErrBadParentsMerkleRoot, str)
	}

	// Repeated parents
	parentsSet := blockdag.NewHashSet()
	parentsSet.AddList(msgBlock.Parents)
	if len(msgBlock.Parents) != parentsSet.Size() {
		str := fmt.Sprintf("parents:%v", msgBlock.Parents)
		return ruleError(ErrDuplicateParent, str)
	}

	// A block must have at least one regular transaction.
	numTx := len(msgBlock.Transactions)
	if numTx == 0 {
		return ruleError(ErrNoTransactions, "block does not contain "+
			"any transactions")
	}

	// A block must not exceed the maximum allowed block payload when
	// serialized.
	//
	// This is a quick and context-free sanity check of the maximum block
	// size according to the wire protocol.  Even though the wire protocol
	// already prevents blocks bigger than this limit, there are other
	// methods of receiving a block that might not have been checked
	// already.  A separate block size is enforced later that takes into
	// account the network-specific block size and the results of block
	// size votes.  Typically that block size is more restrictive than this
	// one.
	serializedSize := msgBlock.SerializeSize()
	if serializedSize > types.MaxBlockPayload {
		str := fmt.Sprintf("serialized block is too big - got %d, "+
			"max %d", serializedSize, types.MaxBlockPayload)
		return ruleError(ErrBlockTooBig, str)
	}

	// The first transaction in a block's regular tree must be a coinbase.
	transactions := block.Transactions()
	if !transactions[0].Transaction().IsCoinBase() {
		return ruleError(ErrFirstTxNotCoinbase, "first transaction in "+
			"block is not a coinbase")
	}

	// A block must not have more than one coinbase.
	for i, tx := range transactions[1:] {
		if tx.Transaction().IsCoinBase() {
			str := fmt.Sprintf("block contains second coinbase at "+
				"index %d", i+1)
			return ruleError(ErrMultipleCoinbases, str)
		}
	}

	// Do some preliminary checks on each regular transaction to ensure they
	// are sane before continuing.
	for i, tx := range transactions {
		// A block must not have stake transactions in the regular
		// transaction tree.
		msgTx := tx.Transaction()
		txType := types.DetermineTxType(msgTx)
		if txType != types.TxTypeRegular {
			errStr := fmt.Sprintf("block contains a irregular "+
				"transaction in the regular transaction tree at "+
				"index %d", i)
			return ruleError(ErrIrregTxInRegularTree, errStr)
		}

		err := CheckTransactionSanity(msgTx, chainParams)
		if err != nil {
			return err
		}
	}

	// Build merkle tree and ensure the calculated merkle root matches the
	// entry in the block header.  This also has the effect of caching all
	// of the transaction hashes in the block to speed up future hash
	// checks.  Bitcoind builds the tree here and checks the merkle root
	// after the following checks, but there is no reason not to check the
	// merkle root matches here.
	merkles := merkle.BuildMerkleTreeStore(block.Transactions(), false)
	calculatedMerkleRoot := merkles[len(merkles)-1]
	if !header.TxRoot.IsEqual(calculatedMerkleRoot) {
		str := fmt.Sprintf("block merkle root is invalid - block "+
			"header indicates %v, but calculated value is %v",
			header.TxRoot, calculatedMerkleRoot)
		return ruleError(ErrBadMerkleRoot, str)
	}

	// Check for duplicate transactions.  This check will be fairly quick
	// since the transaction hashes are already cached due to building the
	// merkle trees above.
	existingTxHashes := make(map[hash.Hash]struct{})

	allTransactions := append(transactions)

	for _, tx := range allTransactions {
		h := tx.Hash()
		if _, exists := existingTxHashes[*h]; exists {
			str := fmt.Sprintf("block contains duplicate "+
				"transaction %v", h)
			return ruleError(ErrDuplicateTx, str)
		}
		existingTxHashes[*h] = struct{}{}
	}

	// The number of signature operations must be less than the maximum
	// allowed per block.
	totalSigOps := 0
	for _, tx := range allTransactions {
		// We could potentially overflow the accumulator so check for
		// overflow.
		lastSigOps := totalSigOps
		totalSigOps += CountSigOps(tx)
		if totalSigOps < lastSigOps || totalSigOps > MaxSigOpsPerBlock {
			str := fmt.Sprintf("block contains too many signature "+
				"operations - got %v, max %v", totalSigOps,
				MaxSigOpsPerBlock)
			return ruleError(ErrTooManySigOps, str)
		}
	}

	return nil
}

// checkBlockHeaderSanity performs some preliminary checks on a block header to
// ensure it is sane before continuing with processing.  These checks are
// context free.
//
// The flags do not modify the behavior of this function directly, however they
// are needed to pass along to checkProofOfWork.
func checkBlockHeaderSanity(header *types.BlockHeader, timeSource MedianTimeSource, flags BehaviorFlags, chainParams *params.Params,mHeight uint) error {

	// Ensure the proof of work bits in the block header is in min/max
	// range and the block hash is less than the target value described by
	// the bits.
	err := checkProofOfWork(header, chainParams.PowConfig, flags,mHeight)
	if err != nil {
		return err
	}

	// A block timestamp must not have a greater precision than one second.
	// This check is necessary because Go time.Time values support
	// nanosecond precision whereas the consensus rules only apply to
	// seconds and it's much nicer to deal with standard Go time values
	// instead of converting to seconds everywhere.
	if !header.Timestamp.Equal(time.Unix(header.Timestamp.Unix(), 0)) {
		str := fmt.Sprintf("block timestamp of %v has a higher "+
			"precision than one second", header.Timestamp)
		return ruleError(ErrInvalidTime, str)
	}

	// Ensure the block time is not too far in the future.
	maxTimestamp := timeSource.AdjustedTime().Add(time.Second *
		MaxTimeOffsetSeconds)
	if header.Timestamp.After(maxTimestamp) {
		str := fmt.Sprintf("block timestamp of %v is too far in the "+
			"future", header.Timestamp)
		return ruleError(ErrTimeTooNew, str)
	}

	return nil
}

// checkProofOfWork ensures the block header bits which indicate the target
// difficulty is in min/max range and that the block hash is less than the
// target difficulty as claimed.
//
// The flags modify the behavior of this function as follows:
//  - BFNoPoWCheck: The check to ensure the block hash is less than the target
//    difficulty is not performed.
func checkProofOfWork(header *types.BlockHeader, powConfig *pow.PowConfig, flags BehaviorFlags,mHeight uint) error {

	// The block hash must be less than the claimed target unless the flag
	// to avoid proof of work checks is set.
	if flags&BFNoPoWCheck != BFNoPoWCheck {
		header.Pow.SetParams(powConfig)
		header.Pow.SetMainHeight(int64(mHeight))
		// The block hash must be less than the claimed target.
		return header.Pow.Verify(header.BlockData(), header.BlockHash(), header.Difficulty)
	}

	return nil
}

// CheckTransactionSanity performs some preliminary checks on a transaction to
// ensure it is sane.  These checks are context free.
func CheckTransactionSanity(tx *types.Transaction, params *params.Params) error {
	// A transaction must have at least one input.
	if len(tx.TxIn) == 0 {
		return ruleError(ErrNoTxInputs, "transaction has no inputs")
	}

	// A transaction must have at least one output.
	if len(tx.TxOut) == 0 {
		return ruleError(ErrNoTxOutputs, "transaction has no outputs")
	}

	// A transaction must not exceed the maximum allowed size when
	// serialized.
	serializedTxSize := tx.SerializeSize()
	if serializedTxSize > params.MaxTxSize {
		str := fmt.Sprintf("serialized transaction is too big - got "+
			"%d, max %d", serializedTxSize, params.MaxTxSize)
		return ruleError(ErrTxTooBig, str)
	}

	// Ensure the transaction amounts are in range.  Each transaction
	// output must not be negative or more than the max allowed per
	// transaction.  Also, the total of all outputs must abide by the same
	// restrictions.  All amounts in a transaction are in a unit value
	// known as an atom.  One Coin is a quantity of atoms as defined by
	// the AtomsPerCoin constant.
	var totalAtom int64
	for _, txOut := range tx.TxOut {
		atom := txOut.Amount
		if atom < 0 {
			str := fmt.Sprintf("transaction output has negative "+
				"value of %v", atom)
			return ruleError(ErrInvalidTxOutValue, str)
		}
		if atom > types.MaxAmount {
			str := fmt.Sprintf("transaction output value of %v is "+
				"higher than max allowed value of %v", atom,
				types.MaxAmount)
			return ruleError(ErrInvalidTxOutValue, str)
		}

		// Two's complement int64 overflow guarantees that any overflow
		// is detected and reported.
		// TODO revisit the overflow check
		totalAtom += int64(atom)
		if totalAtom < 0 {
			str := fmt.Sprintf("total value of all transaction "+
				"outputs exceeds max allowed value of %v",
				types.MaxAmount)
			return ruleError(ErrInvalidTxOutValue, str)
		}
		if totalAtom > types.MaxAmount {
			str := fmt.Sprintf("total value of all transaction "+
				"outputs is %v which is higher than max "+
				"allowed value of %v", totalAtom,
				types.MaxAmount)
			return ruleError(ErrInvalidTxOutValue, str)
		}
	}

	// Check for duplicate transaction inputs.
	existingTxOut := make(map[types.TxOutPoint]struct{})
	for _, txIn := range tx.TxIn {
		if _, exists := existingTxOut[txIn.PreviousOut]; exists {
			return ruleError(ErrDuplicateTxInputs, "transaction "+
				"contains duplicate inputs")
		}
		existingTxOut[txIn.PreviousOut] = struct{}{}
	}

	// Coinbase script length must be between min and max length.
	if tx.IsCoinBase() {
		slen := len(tx.TxIn[0].SignScript)
		if slen < MinCoinbaseScriptLen || slen > MaxCoinbaseScriptLen {
			str := fmt.Sprintf("coinbase transaction script "+
				"length of %d is out of range (min: %d, max: "+
				"%d)", slen, MinCoinbaseScriptLen,
				MaxCoinbaseScriptLen)
			return ruleError(ErrBadCoinbaseScriptLen, str)
		}
		if len(tx.TxOut) >= 2 {
			slen = len(tx.TxOut[1].PkScript)
			if slen < MinCoinbaseScriptLen || slen > MaxCoinbaseScriptLen {
				str := fmt.Sprintf("coinbase transaction script "+
					"length of %d is out of range (min: %d, max: "+
					"%d)", slen, MinCoinbaseScriptLen,
					MaxCoinbaseScriptLen)
				return ruleError(ErrBadCoinbaseScriptLen, str)
			}
			orgPkScriptStr := hex.EncodeToString(params.OrganizationPkScript)
			curPkScriptStr := hex.EncodeToString(tx.TxOut[1].PkScript)
			if orgPkScriptStr != curPkScriptStr {
				str := fmt.Sprintf("coinbase transaction for block pays to %s, but it is %s",
					orgPkScriptStr, curPkScriptStr)
				return ruleError(ErrBadCoinbaseValue, str)
			}

		} else if len(tx.TxOut) >= 3 {
			// Coinbase TxOut[2] is op return
			nullDataOut := tx.TxOut[2]
			// The first 4 bytes of the null data output must be the encoded height
			// of the block, so that every coinbase created has a unique transaction
			// hash.
			nullData, err := txscript.ExtractCoinbaseNullData(nullDataOut.PkScript)
			if err != nil {
				str := fmt.Sprintf("coinbase output 2:bad nulldata %v", nullData)
				return ruleError(ErrBadCoinbaseOutpoint, str)
			}
		}

	} else {
		// Previous transaction outputs referenced by the inputs to
		// this transaction must not be null except in the case of
		// stake bases for SSGen tx.
		for _, txIn := range tx.TxIn {
			prevOut := &txIn.PreviousOut
			if isNullOutpoint(prevOut) {
				return ruleError(ErrInvalidTxInput, "transaction "+
					"input refers to previous output that "+
					"is null")
			}
		}
	}
	return nil
}

// isNullOutpoint determines whether or not a previous transaction output point
// is set.
func isNullOutpoint(outpoint *types.TxOutPoint) bool {
	if outpoint.OutIndex == math.MaxUint32 &&
		outpoint.Hash.IsEqual(zeroHash) {
		return true
	}
	return false
}

// CountSigOps returns the number of signature operations for all transaction
// input and output scripts in the provided transaction.  This uses the
// quicker, but imprecise, signature operation counting mechanism from
// txscript.
func CountSigOps(tx *types.Tx) int {
	msgTx := tx.Transaction()

	// Accumulate the number of signature operations in all transaction
	// inputs.
	totalSigOps := 0
	for _, txIn := range msgTx.TxIn {
		numSigOps := txscript.GetSigOpCount(txIn.SignScript)
		totalSigOps += numSigOps
	}

	// Accumulate the number of signature operations in all transaction
	// outputs.
	for _, txOut := range msgTx.TxOut {
		numSigOps := txscript.GetSigOpCount(txOut.PkScript)
		totalSigOps += numSigOps
	}

	return totalSigOps
}

// checkBlockContext peforms several validation checks on the block which depend
// on its position within the block chain.
//
// The flags modify the behavior of this function as follows:
//  - BFFastAdd: The transactions are not checked to see if they are finalized
//    and the somewhat expensive duplication transaction check is not performed.
//
// The flags are also passed to checkBlockHeaderContext.  See its documentation
// for how the flags modify its behavior.
func (b *BlockChain) checkBlockContext(block *types.SerializedBlock, mainParent *blockNode, flags BehaviorFlags) error {
	// The genesis block is valid by definition.
	if mainParent == nil {
		return nil
	}
	prevBlock := b.bd.GetBlock(mainParent.GetHash())

	// Perform all block header related validation checks.
	err := b.checkBlockHeaderContext(block, mainParent, flags)
	if err != nil {
		return err
	}
	header := &block.Block().Header
	fastAdd := flags&BFFastAdd == BFFastAdd
	if !fastAdd {
		// A block must not exceed the maximum allowed size as defined
		// by the network parameters and the current status of any hard
		// fork votes to change it when serialized.
		maxBlockSize, err := b.maxBlockSize(mainParent)
		if err != nil {
			return err
		}
		//TODO, revisit block size in header
		/*
			serializedSize := int64(block.Block().Header.Size)
		*/
		blockBytes, _ := block.Bytes()
		serializedSize := int64(len(blockBytes))
		if serializedSize > maxBlockSize {
			str := fmt.Sprintf("serialized block is too big - "+
				"got %d, max %d", serializedSize,
				maxBlockSize)
			return ruleError(ErrBlockTooBig, str)
		}

		// Switch to using the past median time of the block prior to
		// the block being checked for all checks related to lock times
		// once the stake vote for the agenda is active.
		blockTime := header.Timestamp

		// The height of this block is one more than the referenced
		// previous block.
		blockHeight := uint64(prevBlock.GetHeight() + 1)

		// Ensure all transactions in the block are finalized and are
		// not expired.
		for _, tx := range block.Transactions() {
			if !IsFinalizedTransaction(tx, blockHeight, blockTime) {
				str := fmt.Sprintf("block contains unfinalized regular "+
					"transaction %v", tx.Hash())
				return ruleError(ErrUnfinalizedTx, str)
			}

			// The transaction must not be expired.
			if IsExpired(tx, blockHeight) {
				errStr := fmt.Sprintf("block contains expired regular "+
					"transaction %v (expiration height %d)", tx.Hash(),
					tx.Transaction().Expire)
				return ruleError(ErrExpiredTx, errStr)
			}
		}

		// Check that the coinbase contains at minimum the block
		// height in output 1.
		err = checkCoinbaseUniqueHeight(blockHeight, block)
		if err != nil {
			return err
		}

		err = b.checkBlockSubsidy(block, -1)
		if err != nil {
			return err
		}

		err = merkle.ValidateWitnessCommitment(block)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *BlockChain) checkBlockSubsidy(block *types.SerializedBlock, totalFee int64) error {
	parents := blockdag.NewHashSet()
	parents.AddList(block.Block().Parents)
	blocks := b.bd.GetBlues(parents)
	// check subsidy
	transactions := block.Transactions()
	subsidy := b.subsidyCache.CalcBlockSubsidy(int64(blocks))
	workAmountOut := int64(transactions[0].Tx.TxOut[0].Amount)

	hasTax := false
	if b.params.BlockTaxProportion > 0 &&
		len(b.params.OrganizationPkScript) > 0 {
		hasTax = true
	}

	var work int64
	var tax int64
	var taxAmountOut int64 = 0
	var totalAmountOut int64 = 0

	if hasTax {
		if len(transactions[0].Tx.TxOut) < 2 {
			str := fmt.Sprintf("coinbase transaction is illegal")
			return ruleError(ErrBadCoinbaseValue, str)
		}
		work = int64(CalcBlockWorkSubsidy(b.subsidyCache,
			int64(blocks), b.params))
		tax = int64(CalcBlockTaxSubsidy(b.subsidyCache,
			int64(blocks), b.params))

		taxAmountOut = int64(transactions[0].Tx.TxOut[1].Amount)
	} else {
		work = subsidy
		tax = 0
		taxAmountOut = 0
	}

	totalAmountOut = workAmountOut + taxAmountOut

	if totalAmountOut < subsidy {
		str := fmt.Sprintf("coinbase transaction for block pays %v which is not the subsidy %v",
			totalAmountOut, subsidy)
		return ruleError(ErrBadCoinbaseValue, str)
	}

	if workAmountOut < work ||
		tax != taxAmountOut {
		str := fmt.Sprintf("coinbase transaction for block pays %d  %d  which is not the %d  %d",
			workAmountOut, taxAmountOut, work, tax)
		return ruleError(ErrBadCoinbaseValue, str)
	}

	if hasTax {
		orgPkScriptStr := hex.EncodeToString(b.params.OrganizationPkScript)
		curPkScriptStr := hex.EncodeToString(transactions[0].Tx.TxOut[1].PkScript)
		if orgPkScriptStr != curPkScriptStr {
			str := fmt.Sprintf("coinbase transaction for block pays to %s, but it is %s",
				orgPkScriptStr, curPkScriptStr)
			return ruleError(ErrBadCoinbaseValue, str)
		}
	}
	if totalFee != -1 {
		if workAmountOut != (totalFee + work) {
			str := fmt.Sprintf("coinbase transaction for block work pays for %d, but it is %d",
				work, workAmountOut-totalFee)
			return ruleError(ErrBadCoinbaseValue, str)
		}
	}
	return nil
}

// checkBlockHeaderContext peforms several validation checks on the block
// header which depend on its position within the block chain.
//
// The flags modify the behavior of this function as follows:
//  - BFFastAdd: All checks except those involving comparing the header against
//    the checkpoints are not performed.
//
// This function MUST be called with the chain state lock held (for writes).
func (b *BlockChain) checkBlockHeaderContext(block *types.SerializedBlock, prevNode *blockNode, flags BehaviorFlags) error {
	// The genesis block is valid by definition.
	if prevNode == nil {
		return nil
	}

	header := &block.Block().Header
	fastAdd := flags&BFFastAdd == BFFastAdd
	if !fastAdd {
		instance := pow.GetInstance(header.Pow.GetPowType(), 0, []byte{})
		instance.SetMainHeight(int64(prevNode.height + 1))
		instance.SetParams(b.params.PowConfig)
		// Ensure the difficulty specified in the block header matches
		// the calculated difficulty based on the previous block and
		// difficulty retarget rules.
		expDiff, err := b.calcNextRequiredDifficulty(prevNode,
			header.Timestamp, instance)
		if err != nil {
			return err
		}
		blockDifficulty := header.Difficulty
		if blockDifficulty != expDiff {
			str := fmt.Sprintf("block difficulty of %d is not the"+
				" expected value of %d", blockDifficulty,
				expDiff)
			return ruleError(ErrUnexpectedDifficulty, str)
		}

		// Ensure the timestamp for the block header is after the
		// median time of the last several blocks (medianTimeBlocks).
		medianTime := prevNode.CalcPastMedianTime(b)
		if !header.Timestamp.After(medianTime) {
			str := "block timestamp of %v is not after expected %v"
			str = fmt.Sprintf(str, header.Timestamp.Unix(), medianTime.Unix())
			return ruleError(ErrTimeTooOld, str)
		}
	}

	// checkpoint
	if !b.HasCheckpoints() {
		return nil
	}
	parents := blockdag.NewHashSet()
	parents.AddList(block.Block().Parents)
	blockLayer, ok := b.BlockDAG().GetParentsMaxLayer(parents)
	if !ok {
		str := fmt.Sprintf("bad parents:%v", block.Block().Parents)
		return ruleError(ErrMissingParent, str)
	}
	blockLayer += 1
	blockHash := header.BlockHash()
	if !b.verifyCheckpoint(uint64(blockLayer), &blockHash) {
		str := fmt.Sprintf("block at layer %d does not match "+
			"checkpoint hash", blockLayer)
		return ruleError(ErrBadCheckpoint, str)
	}

	checkpointNode, err := b.findPreviousCheckpoint()
	if err != nil {
		return err
	}
	if checkpointNode != nil && blockLayer < checkpointNode.layer {
		str := fmt.Sprintf("block at layer %d forks the main chain "+
			"before the previous checkpoint at layer %d",
			blockLayer, checkpointNode.layer)
		return ruleError(ErrForkTooOld, str)
	}

	return nil
}

// checkConnectBlock performs several checks to confirm connecting the passed
// block to the chain represented by the passed view does not violate any
// rules.  In addition, the passed view is updated to spend all of the
// referenced outputs and add all of the new utxos created by block.  Thus, the
// view will represent the state of the chain as if the block were actually
// connected and consequently the best hash for the view is also updated to
// passed block.
//
// An example of some of the checks performed are ensuring connecting the block
// would not cause any duplicate transaction hashes for old transactions that
// aren't already fully spent, double spends, exceeding the maximum allowed
// signature operations per block, invalid values in relation to the expected
// block subsidy, or fail transaction script validation.
//
// The CheckConnectBlockTemplate function makes use of this function to perform
// the bulk of its work.
//
// This function MUST be called with the chain state lock held (for writes).
func (b *BlockChain) checkConnectBlock(node *blockNode, block *types.SerializedBlock, utxoView *UtxoViewpoint, stxos *[]SpentTxOut) error {
	// If the side chain blocks end up in the database, a call to
	// CheckBlockSanity should be done here in case a previous version
	// allowed a block that is no longer valid.  However, since the
	// implementation only currently uses memory for the side chain blocks,
	// it isn't currently necessary.

	// The coinbase for the Genesis block is not spendable, so just return
	// an error now.
	if node.hash.IsEqual(b.params.GenesisHash) {
		str := "the coinbase for the genesis block is not spendable"
		return ruleError(ErrMissingTxOut, str)
	}

	err := b.checkUtxoDuplicate(block, utxoView)
	if err != nil {
		return err
	}

	// Check that the coinbase pays the tax, if applicable.
	// TODO tax pay

	// Don't run scripts if this node is before the latest known good
	// checkpoint since the validity is verified via the checkpoints (all
	// transactions are included in the merkle root hash and any changes
	// will therefore be detected by the next checkpoint).  This is a huge
	// optimization because running the scripts is the most time consuming
	// portion of block handling.
	checkpoint := b.LatestCheckpoint()
	runScripts := !b.noVerify
	if checkpoint != nil && uint64(node.GetLayer()) <= checkpoint.Layer {
		runScripts = false
	}
	var scriptFlags txscript.ScriptFlags
	if runScripts {
		scriptFlags, err = b.consensusScriptVerifyFlags(node)
		if err != nil {
			return err
		}
	}

	// The number of signature operations must be less than the maximum
	// allowed per block.  Note that the preliminary sanity checks on a
	// block also include a check similar to this one, but this check
	// expands the count to include a precise count of pay-to-script-hash
	// signature operations in each of the input transaction public key
	// scripts.
	// Do this for all TxTrees.

	err = utxoView.fetchInputUtxos(b.db, block, b)
	if err != nil {
		return err
	}

	err = b.checkTransactionsAndConnect(node, block, b.subsidyCache, utxoView, stxos)
	if err != nil {
		log.Trace("checkTransactionsAndConnect failed", "err", err)
		return err
	}

	// Enforce all relative lock times via sequence numbers for the regular
	// transaction tree once the stake vote for the agenda is active.

	// Use the past median time of the *previous* block in order
	// to determine if the transactions in the current block are
	// final.
	prevMedianTime := node.GetMainParent(b).CalcPastMedianTime(b)

	// Skip the coinbase since it does not have any inputs and thus
	// lock times do not apply.
	for _, tx := range block.Transactions() {
		sequenceLock, err := b.calcSequenceLock(node, tx,
			utxoView, false)
		if err != nil {
			return err
		}

		if !SequenceLockActive(sequenceLock, int64(node.GetHeight()), //TODO, remove type conversion
			prevMedianTime) {

			str := fmt.Sprintf("block contains " +
				"transaction whose input sequence " +
				"locks are not met")
			return ruleError(ErrUnfinalizedTx, str)
		}
	}

	if runScripts {
		err = checkBlockScripts(block, utxoView,
			scriptFlags, b.sigCache)
		if err != nil {
			log.Trace("checkBlockScripts failed; error returned "+
				"on txtreeregular of cur block: %v", err)
			return err
		}
	}

	return nil
}

// consensusScriptVerifyFlags returns the script flags that must be used when
// executing transaction scripts to enforce the consensus rules. This includes
// any flags required as the result of any agendas that have passed and become
// active.
func (b *BlockChain) consensusScriptVerifyFlags(node *blockNode) (txscript.ScriptFlags, error) {
	//TODO, refactor the txvm flag, the flag should decided by node.parent
	scriptFlags := txscript.ScriptBip16 |
		txscript.ScriptVerifyDERSignatures |
		txscript.ScriptVerifyStrictEncoding |
		txscript.ScriptVerifyMinimalData |
		txscript.ScriptVerifyCleanStack |
		txscript.ScriptVerifyCheckLockTimeVerify

	scriptFlags |= txscript.ScriptVerifyCheckSequenceVerify
	scriptFlags |= txscript.ScriptVerifySHA256
	return scriptFlags, nil
}

// checkTransactionsAndConnect is the local function used to check the
// transaction inputs for a transaction list given a predetermined TxStore.
// After ensuring the transaction is valid, the transaction is connected to the
// UTXO viewpoint.  TxTree true == Regular, false == Stake
func (b *BlockChain) checkTransactionsAndConnect(node *blockNode, block *types.SerializedBlock, subsidyCache *SubsidyCache, utxoView *UtxoViewpoint, stxos *[]SpentTxOut) error {
	transactions := block.Transactions()
	totalSigOpCost := 0
	for _, tx := range transactions {
		sigOpCost := CountSigOps(tx)

		// Check for overflow or going over the limits.  We have to do
		// this on every loop iteration to avoid overflow.
		lastSigOpCost := totalSigOpCost
		totalSigOpCost += sigOpCost
		if totalSigOpCost < lastSigOpCost || totalSigOpCost > MaxSigOpsPerBlock {
			str := fmt.Sprintf("block contains too many "+
				"signature operations - got %v, max %v",
				totalSigOpCost, MaxSigOpsPerBlock)
			return ruleError(ErrTooManySigOps, str)
		}
	}

	var totalFees int64
	for idx, tx := range transactions {
		txFee, err := CheckTransactionInputs(tx, utxoView, b.params, b.bd)
		if err != nil {
			return err
		}

		// Sum the total fees and ensure we don't overflow the
		// accumulator.
		lastTotalFees := totalFees
		totalFees += txFee
		if totalFees < lastTotalFees {
			return ruleError(ErrBadFees, "total fees for block "+
				"overflows accumulator")
		}

		err = utxoView.connectTransaction(tx, node, uint32(idx), stxos)
		if err != nil {
			return err
		}
	}
	return b.checkBlockSubsidy(block, totalFees)
}

// SequenceLockActive determines if all of the inputs to a given transaction
// have achieved a relative age that surpasses the requirements specified by
// their respective sequence locks as calculated by CalcSequenceLock.  A single
// sequence lock is sufficient because the calculated lock selects the minimum
// required time and block height from all of the non-disabled inputs after
// which the transaction can be included.
func SequenceLockActive(lock *SequenceLock, blockHeight int64, medianTime time.Time) bool {
	// The transaction is not yet mature if it has not yet reached the
	// required minimum time and block height according to its sequence
	// locks.
	if blockHeight <= lock.BlockHeight || medianTime.Unix() <= lock.Time {
		return false
	}

	return true
}

// checkNumSigOps Checks the number of P2SH signature operations to make
// sure they don't overflow the limits.  It takes a cumulative number of sig
// ops as an argument and increments will each call.
// TxTree true == Regular, false == Stake
func checkNumSigOps(tx *types.Tx, utxoView *UtxoViewpoint, index int, txTree bool, cumulativeSigOps int) (int, error) {

	numsigOps := CountSigOps(tx)

	// Since the first (and only the first) transaction has already been
	// verified to be a coinbase transaction, use (i == 0) && TxTree as an
	// optimization for the flag to countP2SHSigOps for whether or not the
	// transaction is a coinbase transaction rather than having to do a
	// full coinbase check again.
	numP2SHSigOps, err := CountP2SHSigOps(tx, (index == 0) && txTree, utxoView)
	if err != nil {
		log.Trace("CountP2SHSigOps failed", "error", err)
		return 0, err
	}

	startCumSigOps := cumulativeSigOps
	cumulativeSigOps += numsigOps
	cumulativeSigOps += numP2SHSigOps

	// Check for overflow or going over the limits.  We have to do
	// this on every loop iteration to avoid overflow.
	if cumulativeSigOps < startCumSigOps ||
		cumulativeSigOps > MaxSigOpsPerBlock {
		str := fmt.Sprintf("block contains too many signature "+
			"operations - got %v, max %v", cumulativeSigOps,
			MaxSigOpsPerBlock)
		return 0, ruleError(ErrTooManySigOps, str)
	}

	return cumulativeSigOps, nil
}

// CountP2SHSigOps returns the number of signature operations for all input
// transactions which are of the pay-to-script-hash type.  This uses the
// precise, signature operation counting mechanism from the script engine which
// requires access to the input transaction scripts.
func CountP2SHSigOps(tx *types.Tx, isCoinBaseTx bool, utxoView *UtxoViewpoint) (int, error) {
	// Coinbase transactions have no interesting inputs.
	if isCoinBaseTx {
		return 0, nil
	}

	// Accumulate the number of signature operations in all transaction
	// inputs.
	msgTx := tx.Transaction()
	totalSigOps := 0
	for txInIndex, txIn := range msgTx.TxIn {
		// Ensure the referenced input transaction is available.
		utxoEntry := utxoView.LookupEntry(txIn.PreviousOut)
		if utxoEntry == nil || utxoEntry.IsSpent() {
			str := fmt.Sprintf("output %v referenced from "+
				"transaction %s:%d either does not exist or "+
				"has already been spent", txIn.PreviousOut,
				tx.Hash(), txInIndex)
			return 0, ruleError(ErrMissingTxOut, str)
		}

		// We're only interested in pay-to-script-hash types, so skip
		// this input if it's not one.
		pkScript := utxoEntry.PkScript()
		if !txscript.IsPayToScriptHash(pkScript) {
			continue
		}

		// Count the precise number of signature operations in the
		// referenced public key script.
		sigScript := txIn.SignScript
		numSigOps := txscript.GetPreciseSigOpCount(sigScript, pkScript,
			true)

		// We could potentially overflow the accumulator so check for
		// overflow.
		lastSigOps := totalSigOps
		totalSigOps += numSigOps
		if totalSigOps < lastSigOps {
			str := fmt.Sprintf("the public key script from output "+
				"%v contains too many signature operations - "+
				"overflow", txIn.PreviousOut)
			return 0, ruleError(ErrTooManySigOps, str)
		}
	}

	return totalSigOps, nil
}

// CheckTransactionInputs performs a series of checks on the inputs to a
// transaction to ensure they are valid.  An example of some of the checks
// include verifying all inputs exist, ensuring the coinbase seasoning
// requirements are met, detecting double spends, validating all values and
// fees are in the legal range and the total output amount doesn't exceed the
// input amount, and verifying the signatures to prove the spender was the
// owner and therefore allowed to spend them.  As it checks the inputs, it
// also calculates the total fees for the transaction and returns that value.
//
// NOTE: The transaction MUST have already been sanity checked with the
// CheckTransactionSanity function prior to calling this function.
func CheckTransactionInputs(tx *types.Tx, utxoView *UtxoViewpoint, chainParams *params.Params, bd *blockdag.BlockDAG) (int64, error) {
	msgTx := tx.Transaction()

	txHash := tx.Hash()
	var totalAtomIn int64

	// Coinbase transactions have no inputs.
	if msgTx.IsCoinBase() {
		return 0, nil
	}

	// -------------------------------------------------------------------
	// General transaction testing.
	// -------------------------------------------------------------------
	for idx, txIn := range msgTx.TxIn {
		txInHash := &txIn.PreviousOut.Hash
		utxoEntry := utxoView.LookupEntry(txIn.PreviousOut)
		if utxoEntry == nil || utxoEntry.IsSpent() {
			str := fmt.Sprintf("output %v referenced from "+
				"transaction %s:%d either does not exist or "+
				"has already been spent", txIn.PreviousOut,
				txHash, idx)
			return 0, ruleError(ErrMissingTxOut, str)
		}

		// Ensure the transaction is not spending coins which have not
		// yet reached the required coinbase maturity.
		coinbaseMaturity := int64(chainParams.CoinbaseMaturity)

		if utxoEntry.IsCoinBase() {
			if len(utxoView.viewpoints) == 0 {
				str := fmt.Sprintf("transaction %s has no viewpoints", txHash)
				return 0, ruleError(ErrNoViewpoint, str)
			}
			maturity := int64(bd.GetMaturity(utxoEntry.BlockHash(), utxoView.viewpoints))

			if maturity < coinbaseMaturity {
				str := fmt.Sprintf("tx %v tried to spend "+
					"coinbase transaction %v from "+
					"at %v before required "+
					"maturity of %v blocks", txHash,
					txInHash, maturity, coinbaseMaturity)
				return 0, ruleError(ErrImmatureSpend, str)
			}

			if !bd.IsBlue(utxoEntry.BlockHash()) {
				str := fmt.Sprintf("tx %v tried to spend "+
					"coinbase transaction %v from "+
					"at %v before required "+
					"maturity of %v blocks, but it is't in blue set", txHash,
					txInHash, maturity, coinbaseMaturity)
				return 0, ruleError(ErrNoBlueCoinbase, str)
			}
		}

		// Ensure the transaction amounts are in range.  Each of the
		// output values of the input transactions must not be negative
		// or more than the max allowed per transaction.  All amounts
		// in a transaction are in a unit value known as an atom.  One
		// Coin is a quantity of atoms as defined by the AtomPerCoin
		// constant.
		originTxAtom := utxoEntry.Amount()
		if originTxAtom < 0 {
			str := fmt.Sprintf("transaction output has negative "+
				"value of %v", originTxAtom)
			return 0, ruleError(ErrInvalidTxOutValue, str)
		}
		if originTxAtom > types.MaxAmount {
			str := fmt.Sprintf("transaction output value of %v is "+
				"higher than max allowed value of %v",
				originTxAtom, types.MaxAmount)
			return 0, ruleError(ErrInvalidTxOutValue, str)
		}

		// The total of all outputs must not be more than the max
		// allowed per transaction.  Also, we could potentially
		// overflow the accumulator so check for overflow.
		lastAtomIn := totalAtomIn
		totalAtomIn += int64(originTxAtom) //TODO, remove type conversion
		if totalAtomIn < lastAtomIn ||
			totalAtomIn > types.MaxAmount {
			str := fmt.Sprintf("total value of all transaction "+
				"inputs is %v which is higher than max "+
				"allowed value of %v", totalAtomIn,
				types.MaxAmount)
			return 0, ruleError(ErrInvalidTxOutValue, str)
		}
	}

	// Calculate the total output amount for this transaction.  It is safe
	// to ignore overflow and out of range errors here because those error
	// conditions would have already been caught by checkTransactionSanity.
	var totalAtomOut int64
	for _, txOut := range tx.Transaction().TxOut {
		totalAtomOut += int64(txOut.Amount) //TODO, remove type conversion
	}

	// Ensure the transaction does not spend more than its inputs.
	if totalAtomIn < totalAtomOut {
		str := fmt.Sprintf("total value of all transaction inputs for "+
			"transaction %v is %v which is less than the amount "+
			"spent of %v", txHash, totalAtomIn, totalAtomOut)
		return 0, ruleError(ErrSpendTooHigh, str)
	}

	// NOTE: bitcoind checks if the transaction fees are < 0 here, but that
	// is an impossible condition because of the check above that ensures
	// the inputs are >= the outputs.
	txFeeInAtom := totalAtomIn - totalAtomOut

	return txFeeInAtom, nil
}

// CheckConnectBlockTemplate fully validates that connecting the passed block to
// either the tip of the main chain or its parent does not violate any consensus
// rules, aside from the proof of work requirement.  The block must connect to
// the current tip of the main chain or its parent.
//
// This function is safe for concurrent access.
func (b *BlockChain) CheckConnectBlockTemplate(block *types.SerializedBlock) error {
	b.ChainRLock()
	defer b.ChainRUnlock()

	// Skip the proof of work check as this is just a block template.
	flags := BFNoPoWCheck

	// The block template must build off the current tip of the main chain
	// or its parent.

	// Perform context-free sanity checks on the block and its transactions.
	err := b.checkBlockSanity(block, b.timeSource, flags, b.params)
	if err != nil {
		return err
	}

	tipsNode := []*blockNode{}
	for _, v := range block.Block().Parents {
		bn := b.index.LookupNode(v)
		if bn == nil {
			return ruleError(ErrPrevBlockNotBest, "tipsNode")
		}
		tipsNode = append(tipsNode, bn)
	}
	if len(tipsNode) == 0 {
		return ruleError(ErrPrevBlockNotBest, "tipsNode")
	}
	header := &block.Block().Header
	newNode := newBlockNode(header, tipsNode)
	newNode.SetOrder(block.Order())
	newNode.SetHeight(block.Height())
	newNode.SetLayer(GetMaxLayerFromList(tipsNode) + 1)

	view := NewUtxoViewpoint()
	view.SetViewpoints(block.Block().Parents)

	mainParent := newNode.GetMainParent(b)
	mainParentNode := b.index.LookupNode(mainParent.GetHash())
	newNode.CalcWorkSum(mainParentNode)

	err = b.checkBlockContext(block, mainParentNode, flags)
	if err != nil {
		return err
	}
	err = b.checkConnectBlock(newNode, block, view, nil)
	if err != nil {
		return err
	}
	return nil
}

func ExtractCoinbaseHeight(coinbaseTx *types.Transaction) (uint64, error) {
	sigScript := coinbaseTx.TxIn[0].SignScript
	if len(sigScript) < 1 {
		str := "It has not the coinbase signature script for blocks"
		return 0, ruleError(ErrMissingCoinbaseHeight, str)
	}

	// Detect the case when the block height is a small integer encoded with
	// as single byte.
	opcode := int(sigScript[0])
	if opcode == txscript.OP_0 {
		return 0, nil
	}
	if opcode >= txscript.OP_1 && opcode <= txscript.OP_16 {
		return uint64(opcode - (txscript.OP_1 - 1)), nil
	}

	// Otherwise, the opcode is the length of the following bytes which
	// encode in the block height.
	serializedLen := int(sigScript[0])
	if len(sigScript[1:]) < serializedLen {
		str := "It has not the coinbase signature script for blocks"
		return 0, ruleError(ErrMissingCoinbaseHeight, str)
	}

	serializedHeightBytes := make([]byte, 8)
	copy(serializedHeightBytes, sigScript[1:serializedLen+1])
	serializedHeight := binary.LittleEndian.Uint64(serializedHeightBytes)

	return serializedHeight, nil
}
