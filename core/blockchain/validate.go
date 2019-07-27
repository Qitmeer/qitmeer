// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"
	"github.com/HalalChain/qitmeer-lib/core/dag"
	"time"
	"math"
	"math/big"
	"github.com/HalalChain/qitmeer-lib/core/types"
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer-lib/params"
	"github.com/HalalChain/qitmeer/core/merkle"
	"github.com/HalalChain/qitmeer-lib/engine/txscript"
)

// This function only differs from IsExpired in that it works with a raw wire
// transaction as opposed to a higher level util transaction.
func IsExpiredTx(tx *types.Transaction, blockHeight uint64) bool {
	expiry := tx.Expire
	return expiry != types.NoExpiryValue && blockHeight >= uint64(expiry)  //TODO, remove type conversion
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
func checkBlockSanity(block *types.SerializedBlock, timeSource MedianTimeSource, flags BehaviorFlags, chainParams *params.Params) error {
	msgBlock := block.Block()
	header := &msgBlock.Header
	err := checkBlockHeaderSanity(header, timeSource, flags, chainParams)
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
	// TODO header-size check, why
	/*
	if header.Size != uint32(serializedSize) {
		str := fmt.Sprintf("serialized block is not size indicated in "+
			"header - got %d, expected %d", header.Size,
			serializedSize)
		return ruleError(ErrWrongBlockSize, str)
	}
	*/

	// The first transaction in a block's regular tree must be a coinbase.
	transactions := block.Transactions()
	if !transactions[0].Transaction().IsCoinBaseTx() {
		return ruleError(ErrFirstTxNotCoinbase, "first transaction in "+
			"block is not a coinbase")
	}

	// A block must not have more than one coinbase.
	for i, tx := range transactions[1:] {
		if tx.Transaction().IsCoinBaseTx() {
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
	merkles := merkle.BuildMerkleTreeStore(block.Transactions())
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

		msgTx := tx.Transaction()
		isCoinBase := msgTx.IsCoinBaseTx()
		totalSigOps += CountSigOps(tx, isCoinBase)
		if totalSigOps < lastSigOps || totalSigOps > MaxSigOpsPerBlock {
			str := fmt.Sprintf("block contains too many signature "+
				"operations - got %v, max %v", totalSigOps,
				MaxSigOpsPerBlock)
			return ruleError(ErrTooManySigOps, str)
		}
	}

	return nil
}

// CheckBlockSanity performs some preliminary checks on a block to ensure it is
// sane before continuing with block processing.  These checks are context
// free.
func CheckBlockSanity(block *types.SerializedBlock, timeSource MedianTimeSource, chainParams *params.Params) error {
	return checkBlockSanity(block, timeSource, BFNone, chainParams)
}

// checkBlockHeaderSanity performs some preliminary checks on a block header to
// ensure it is sane before continuing with processing.  These checks are
// context free.
//
// The flags do not modify the behavior of this function directly, however they
// are needed to pass along to checkProofOfWork.
func checkBlockHeaderSanity(header *types.BlockHeader, timeSource MedianTimeSource, flags BehaviorFlags, chainParams *params.Params) error {

	// Ensure the proof of work bits in the block header is in min/max
	// range and the block hash is less than the target value described by
	// the bits.
	err := checkProofOfWork(header, chainParams.PowLimit, flags)
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
func checkProofOfWork(header *types.BlockHeader, powLimit *big.Int, flags BehaviorFlags) error {
	// The target difficulty must be larger than zero.
	target := CompactToBig(header.Difficulty)
	if target.Sign() <= 0 {
		str := fmt.Sprintf("block target difficulty of %064x is too "+
			"low", target)
		return ruleError(ErrUnexpectedDifficulty, str)
	}

	// The target difficulty must be less than the maximum allowed.
	if target.Cmp(powLimit) > 0 {
		str := fmt.Sprintf("block target difficulty of %064x is "+
			"higher than max of %064x", target, powLimit)
		return ruleError(ErrUnexpectedDifficulty, str)
	}

	// The block hash must be less than the claimed target unless the flag
	// to avoid proof of work checks is set.
	if flags&BFNoPoWCheck != BFNoPoWCheck {
		// The block hash must be less than the claimed target.
		h := header.BlockHash()
		hashNum := HashToBig(&h)
		if hashNum.Cmp(target) > 0 {
			str := fmt.Sprintf("block hash of %064x is higher than"+
				" expected max of %064x", hashNum, target)
			return ruleError(ErrHighHash, str)
		}
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
			return ruleError(ErrBadTxOutValue, str)
		}
		if atom > types.MaxAmount {
			str := fmt.Sprintf("transaction output value of %v is "+
				"higher than max allowed value of %v", atom,
				types.MaxAmount)
			return ruleError(ErrBadTxOutValue, str)
		}

		// Two's complement int64 overflow guarantees that any overflow
		// is detected and reported.
		// TODO revisit the overflow check
		totalAtom += int64(atom)
		if totalAtom < 0 {
			str := fmt.Sprintf("total value of all transaction "+
				"outputs exceeds max allowed value of %v",
				types.MaxAmount)
			return ruleError(ErrBadTxOutValue, str)
		}
		if totalAtom > types.MaxAmount {
			str := fmt.Sprintf("total value of all transaction "+
				"outputs is %v which is higher than max "+
				"allowed value of %v", totalAtom,
				types.MaxAmount)
			return ruleError(ErrBadTxOutValue, str)
		}
	}

	// Coinbase script length must be between min and max length.
	if tx.IsCoinBaseTx() {
		// The referenced outpoint should be null.
		if !isNullOutpoint(&tx.TxIn[0].PreviousOut) {
			str := fmt.Sprintf("coinbase transaction did not use " +
				"a null outpoint")
			return ruleError(ErrBadCoinbaseOutpoint, str)
		}

		// The fraud proof should also be null.
		if !isNullFraudProof(tx.TxIn[0]) {
			str := fmt.Sprintf("coinbase transaction fraud proof " +
				"was non-null")
			return ruleError(ErrBadCoinbaseFraudProof, str)
		}

		slen := len(tx.TxIn[0].SignScript)
		if slen < MinCoinbaseScriptLen || slen > MaxCoinbaseScriptLen {
			str := fmt.Sprintf("coinbase transaction script "+
				"length of %d is out of range (min: %d, max: "+
				"%d)", slen, MinCoinbaseScriptLen,
				MaxCoinbaseScriptLen)
			return ruleError(ErrBadCoinbaseScriptLen, str)
		}
	} else {
		// Previous transaction outputs referenced by the inputs to
		// this transaction must not be null except in the case of
		// stake bases for SSGen tx.
		for _, txIn := range tx.TxIn {
			prevOut := &txIn.PreviousOut
			if isNullOutpoint(prevOut) {
				return ruleError(ErrBadTxInput, "transaction "+
					"input refers to previous output that "+
					"is null")
			}
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

// isNullFraudProof determines whether or not a previous transaction fraud
// proof is set.
func isNullFraudProof(txIn *types.TxInput) bool {
	switch {
	case txIn.BlockHeight != types.NullBlockHeight:
		return false
	case txIn.TxIndex != types.NullTxIndex:
		return false
	}

	return true
}

// CountSigOps returns the number of signature operations for all transaction
// input and output scripts in the provided transaction.  This uses the
// quicker, but imprecise, signature operation counting mechanism from
// txscript.
func CountSigOps(tx *types.Tx, isCoinBaseTx bool) int {
	msgTx := tx.Transaction()

	// Accumulate the number of signature operations in all transaction
	// inputs.
	totalSigOps := 0
	for _, txIn := range msgTx.TxIn {
		// Skip coinbase inputs.
		if isCoinBaseTx {
			continue
		}

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
func (b *BlockChain) checkBlockContext(block *types.SerializedBlock, prevNode *blockNode, flags BehaviorFlags) error {
	// The genesis block is valid by definition.
	if prevNode == nil {
		return nil
	}
	prevBlock:=b.bd.GetBlock(prevNode.GetHash())

	// Perform all block header related validation checks.
	header := &block.Block().Header
	err := b.checkBlockHeaderContext(header, prevNode, flags)
	if err != nil {
		return err
	}

	fastAdd := flags&BFFastAdd == BFFastAdd
	if !fastAdd {
		// A block must not exceed the maximum allowed size as defined
		// by the network parameters and the current status of any hard
		// fork votes to change it when serialized.
		maxBlockSize, err := b.maxBlockSize(prevNode)
		if err != nil {
			return err
		}
		//TODO, revisit block size in header
		/*
		serializedSize := int64(block.Block().Header.Size)
		*/
		blockBytes,_ :=block.Bytes()
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
func (b *BlockChain) checkBlockHeaderContext(header *types.BlockHeader, prevNode *blockNode, flags BehaviorFlags) error {
	// The genesis block is valid by definition.
	if prevNode == nil {
		return nil
	}

	fastAdd := flags&BFFastAdd == BFFastAdd
	if !fastAdd {
		// Ensure the difficulty specified in the block header matches
		// the calculated difficulty based on the previous block and
		// difficulty retarget rules.
		expDiff, err := b.calcNextRequiredDifficulty(prevNode,
			header.Timestamp)
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
		medianTime := prevNode.CalcPastMedianTime()
		if !header.Timestamp.After(medianTime) {
			str := "block timestamp of %v is not after expected %v"
			str = fmt.Sprintf(str, header.Timestamp, medianTime)
			return ruleError(ErrTimeTooOld, str)
		}
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
func (b *BlockChain) checkConnectBlock(node *blockNode, block *types.SerializedBlock, utxoView *UtxoViewpoint, stxos *[]spentTxOut) error {
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

	// Check that the coinbase pays the tax, if applicable.
	// TODO tax pay


	// Don't run scripts if this node is before the latest known good
	// checkpoint since the validity is verified via the checkpoints (all
	// transactions are included in the merkle root hash and any changes
	// will therefore be detected by the next checkpoint).  This is a huge
	// optimization because running the scripts is the most time consuming
	// portion of block handling.
	checkpoint := b.latestCheckpoint()
	runScripts := !b.noVerify
	if checkpoint != nil && node.order <= checkpoint.Height {
		runScripts = false
	}
	var scriptFlags txscript.ScriptFlags
	if runScripts {
		var err error
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

	err := utxoView.fetchInputUtxos(b.db, block,b)
	if err != nil {
		return err
	}

	// Enforce all relative lock times via sequence numbers for the regular
	// transaction tree once the stake vote for the agenda is active.

	// Use the past median time of the *previous* block in order
		// to determine if the transactions in the current block are
		// final.
	var prevMedianTime time.Time = node.GetBackParent().CalcPastMedianTime()

		// Skip the coinbase since it does not have any inputs and thus
		// lock times do not apply.
	for _, tx := range block.Transactions()[1:] {
		sequenceLock, err := b.calcSequenceLock(node, tx,
			utxoView, true)
		if err != nil {
			return err
		}
		if !SequenceLockActive(sequenceLock, int64(node.order),  //TODO, remove type conversion
			prevMedianTime) {

			str := fmt.Sprintf("block contains " +
				"transaction whose input sequence " +
				"locks are not met")
			return ruleError(ErrUnfinalizedTx, str)
		}
	}


	if runScripts {
		err = checkBlockScripts(block, utxoView, false, scriptFlags,
			b.sigCache,b)
		if err != nil {
			log.Trace("checkBlockScripts failed; error returned "+
				"on txtreestake of cur block: %v", err)
			return err
		}
	}

	// TxTreeRegular of current block. At this point, the stake
	// transactions have already added, so set this to the correct stake
	// viewpoint and disable automatic connection.
	err = b.checkDupTxs(block.Transactions(), utxoView)
	if err != nil {
		log.Trace("checkDupTxs failed for cur TxTreeRegular: %v", err)
		return err
	}

	err = utxoView.fetchInputUtxos(b.db, block,b)
	if err != nil {
		return err
	}

	//TODO, refactor/remove staketreefee
	stakeTreeFees:=int64(0)

	err = b.checkTransactionsAndConnect(b.subsidyCache, stakeTreeFees, node,
		block.Transactions(), utxoView, stxos, true)
	if err != nil {
		log.Trace("checkTransactionsAndConnect failed","err", err)
		return err
	}

	if runScripts {
		err = checkBlockScripts(block, utxoView, true,
			scriptFlags, b.sigCache,b)
		if err != nil {
			log.Trace("checkBlockScripts failed; error returned "+
				"on txtreeregular of cur block: %v", err)
			return err
		}
	}

	// Rollback the final tx tree regular so that we don't write it to
	// database.
	/*if node.height > 1 && stxos != nil {
		_, err := utxoView.disconnectTransactionSlice(block.Transactions(),
			int64(node.height), stxos) //TODO,remove type conversion
		if err != nil {
			return err
		}
		stxosDeref := *stxos
		*stxos = stxosDeref[0:idx]
	}*/

	// First block has special rules concerning the ledger.
	// TODO, block one ICO
	/*
	if node.height == 1 {
		err := BlockOneCoinbasePaysTokens(block.Transactions()[0],
			b.params)
		if err != nil {
			return err
		}
	}
	*/

	// Update the best hash for view to include this block since all of its
	// transactions have been connected.
	utxoView.SetBestHash(&node.hash)

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
func (b *BlockChain) checkTransactionsAndConnect(subsidyCache *SubsidyCache, inputFees int64, node *blockNode,
	txs []*types.Tx, utxoView *UtxoViewpoint, stxos *[]spentTxOut, txTree bool) error {
	// Perform several checks on the inputs for each transaction.  Also
	// accumulate the total fees.  This could technically be combined with
	// the loop above instead of running another loop over the
	// transactions, but by separating it we can avoid running the more
	// expensive (though still relatively cheap as compared to running the
	// scripts) checks against all the inputs when the signature operations
	// are out of bounds.
	totalFees := int64(inputFees) // Stake tx tree carry forward
	var cumulativeSigOps int
	for idx, tx := range txs {
		if b.IsBadTx(tx.Hash()) {
			continue
		}
		// Ensure that the number of signature operations is not beyond
		// the consensus limit.
		var err error
		cumulativeSigOps, err = checkNumSigOps(tx, utxoView, idx,
			txTree, cumulativeSigOps)
		if err != nil {
			log.Trace("checkNumSigOps failed","err", err)
			//return err

			b.AddBadTx(tx.Hash(),&node.hash)
			continue
		}

		// This step modifies the txStore and marks the tx outs used
		// spent, so be aware of this.
		txFee, err := CheckTransactionInputs(b.subsidyCache, tx,
			int64(node.order), utxoView, false, /* check fraud proofs */
			b.params) //TODO, remove type conversion
		if err != nil {
			log.Trace("CheckTransactionInputs failed","err", err)
			//return err
			b.AddBadTx(tx.Hash(),&node.hash)
			continue
		}

		// Sum the total fees and ensure we don't overflow the
		// accumulator.
		lastTotalFees := totalFees
		totalFees += txFee
		if totalFees < lastTotalFees {
			//return ruleError(ErrBadFees, "total fees for block "+
			//	"overflows accumulator")
			log.Trace("total fees for block overflows accumulator")

			b.AddBadTx(tx.Hash(),&node.hash)
			continue
		}

		// Connect the transaction to the UTXO viewpoint, so that in
		// flight transactions may correctly validate.
		err = utxoView.connectTransaction(tx, node.order, uint32(idx),
			stxos)
		if err != nil {
			log.Trace("connectTransaction failed","err", err)
			//return err

			b.AddBadTx(tx.Hash(),&node.hash)
			continue
		}
	}

	// The total output values of the coinbase transaction must not exceed
	// the expected subsidy value plus total transaction fees gained from
	// mining the block.  It is safe to ignore overflow and out of range
	// errors here because those error conditions would have already been
	// caught by checkTransactionSanity.
	/*if txTree { //TxTreeRegular

		var totalAtomOutRegular uint64

		for _, txOut := range txs[0].Transaction().TxOut {
			totalAtomOutRegular += txOut.Amount
		}

		var expAtomOut int64
		if node.height == 0 {
			expAtomOut = subsidyCache.CalcBlockSubsidy(int64(node.height)) //TODO, remove type conversion
		} else {
			subsidyWork := CalcBlockWorkSubsidy(subsidyCache,
				int64(node.height), 0, b.params)                    //TODO, remove type conversion
			subsidyTax := CalcBlockTaxSubsidy(subsidyCache,
				int64(node.height), 0, b.params)                    //TODO, remove type conversion
			expAtomOut = int64(subsidyWork) + subsidyTax + totalFees       //TODO, remove type conversion
		}

		// AmountIn for the input should be equal to the subsidy.
		coinbaseIn := txs[0].Transaction().TxIn[0]
		subsidyWithoutFees := expAtomOut - totalFees
		if (int64(coinbaseIn.AmountIn) != subsidyWithoutFees) &&  //TODO, remove type conversion
			(node.height > 0) {
			errStr := fmt.Sprintf("bad coinbase subsidy in input;"+
				" got %v, expected %v", coinbaseIn.AmountIn,
				subsidyWithoutFees)
			return ruleError(ErrBadCoinbaseAmountIn, errStr)
		}

		if totalAtomOutRegular > uint64(expAtomOut) { //TODO, remove type conversion
			str := fmt.Sprintf("coinbase transaction for block %v"+
				" pays %v which is more than expected value "+
				"of %v", node.hash, totalAtomOutRegular,
				expAtomOut)
			return ruleError(ErrBadCoinbaseValue, str)
		}
	} else {
		// TxTreeStake
	}
	*/
	return nil
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
	if blockHeight <= lock.MinHeight || medianTime.Unix() <= lock.MinTime {
		return false
	}

	return true
}

// checkDupTxs ensures blocks do not contain duplicate transactions which
// 'overwrite' older transactions that are not fully spent.  This prevents an
// attack where a coinbase and all of its dependent transactions could be
// duplicated to effectively revert the overwritten transactions to a single
// confirmation thereby making them vulnerable to a double spend.
//
// For more details, see https://en.bitcoin.it/wiki/BIP_0030 and
// http://r6.ca/blog/20120206T005236Z.html.
//
func (b *BlockChain) checkDupTxs(txSet []*types.Tx, view *UtxoViewpoint) error {
	if !params.CheckForDuplicateHashes {
		return nil
	}

	// Fetch utxo details for all of the transactions in this block.
	// Typically, there will not be any utxos for any of the transactions.
	fetchSet := make(map[hash.Hash]struct{})
	for _, tx := range txSet {
		fetchSet[*tx.Hash()] = struct{}{}
	}
	err := view.fetchUtxos(b.db, fetchSet)
	if err != nil {
		return err
	}

	// Duplicate transactions are only allowed if the previous transaction
	// is fully spent.
	for _, tx := range txSet {
		txEntry := view.LookupEntry(tx.Hash())
		if txEntry != nil && !txEntry.IsFullySpent() {
			str := fmt.Sprintf("tried to overwrite transaction %v "+
				"at block height %d that is not fully spent",
				tx.Hash(), txEntry.BlockHeight())
			return ruleError(ErrOverwriteTx, str)
		}
	}

	return nil
}

// checkNumSigOps Checks the number of P2SH signature operations to make
// sure they don't overflow the limits.  It takes a cumulative number of sig
// ops as an argument and increments will each call.
// TxTree true == Regular, false == Stake
func checkNumSigOps(tx *types.Tx, utxoView *UtxoViewpoint, index int, txTree bool, cumulativeSigOps int) (int, error) {

	numsigOps := CountSigOps(tx, (index == 0) && txTree )

	// Since the first (and only the first) transaction has already been
	// verified to be a coinbase transaction, use (i == 0) && TxTree as an
	// optimization for the flag to countP2SHSigOps for whether or not the
	// transaction is a coinbase transaction rather than having to do a
	// full coinbase check again.
	numP2SHSigOps, err := CountP2SHSigOps(tx, (index == 0) && txTree, utxoView)
	if err != nil {
		log.Trace("CountP2SHSigOps failed","error", err)
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
func CountP2SHSigOps(tx *types.Tx, isCoinBaseTx bool,utxoView *UtxoViewpoint) (int, error) {
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
		originTxHash := &txIn.PreviousOut.Hash
		originTxIndex := txIn.PreviousOut.OutIndex
		utxoEntry := utxoView.LookupEntry(originTxHash)
		if utxoEntry == nil || utxoEntry.IsOutputSpent(originTxIndex) {
			str := fmt.Sprintf("output %v referenced from "+
				"transaction %s:%d either does not exist or "+
				"has already been spent", txIn.PreviousOut,
				tx.Hash(), txInIndex)
			return 0, ruleError(ErrMissingTxOut, str)
		}

		// We're only interested in pay-to-script-hash types, so skip
		// this input if it's not one.
		pkScript := utxoEntry.PkScriptByIndex(originTxIndex)
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
func CheckTransactionInputs(subsidyCache *SubsidyCache, tx *types.Tx, txHeight int64, utxoView *UtxoViewpoint, checkFraudProof bool, chainParams *params.Params) (int64, error) {
	msgTx := tx.Transaction()

	txHash := tx.Hash()
	var totalAtomIn int64

	// Coinbase transactions have no inputs.
	if msgTx.IsCoinBaseTx() {
		return 0, nil
	}
	// -------------------------------------------------------------------
	// General transaction testing.
	// -------------------------------------------------------------------
	for idx, txIn := range msgTx.TxIn {

		txInHash := &txIn.PreviousOut.Hash
		originTxIndex := txIn.PreviousOut.OutIndex
		utxoEntry := utxoView.LookupEntry(txInHash)
		if utxoEntry == nil || utxoEntry.IsOutputSpent(originTxIndex) {
			str := fmt.Sprintf("output %v referenced from "+
				"transaction %s:%d either does not exist or "+
				"has already been spent", txIn.PreviousOut,
				txHash, idx)
			return 0, ruleError(ErrMissingTxOut, str)
		}

		// Check fraud proof witness data.

		// Using zero value outputs as inputs is banned.
		if utxoEntry.AmountByIndex(originTxIndex) == 0 {
			str := fmt.Sprintf("tried to spend zero value output "+
				"from input %v, idx %v", txInHash,
				originTxIndex)
			return 0, ruleError(ErrZeroValueOutputSpend, str)
		}

		if checkFraudProof {
			if txIn.AmountIn !=
				utxoEntry.AmountByIndex(originTxIndex) {
				str := fmt.Sprintf("bad fraud check value in "+
					"(expected %v, given %v) for txIn %v",
					utxoEntry.AmountByIndex(originTxIndex),
					txIn.AmountIn, idx)
				return 0, ruleError(ErrFraudAmountIn, str)
			}

			/*if txIn.BlockHeight != uint32(utxoEntry.BlockHeight()) {  //TODO, remove type conversion
				str := fmt.Sprintf("bad fraud check block "+
					"height (expected %v, given %v) for "+
					"txIn %v", utxoEntry.BlockHeight(),
					txIn.BlockHeight, idx)
				return 0, ruleError(ErrFraudBlockHeight, str)
			}*/

			if txIn.TxIndex != utxoEntry.TxIndex() {
				str := fmt.Sprintf("bad fraud check block "+
					"index (expected %v, given %v) for "+
					"txIn %v", utxoEntry.TxIndex(),
					txIn.TxIndex, idx)
				return 0, ruleError(ErrFraudBlockIndex, str)
			}
		}

		// Ensure the transaction is not spending coins which have not
		// yet reached the required coinbase maturity.
		coinbaseMaturity := int64(chainParams.CoinbaseMaturity)
		originHeight := utxoEntry.BlockHeight()
		if utxoEntry.IsCoinBase() {
			blocksSincePrev := txHeight - int64(originHeight) //TODO,remove type conversion
			if blocksSincePrev < coinbaseMaturity {
				str := fmt.Sprintf("tx %v tried to spend "+
					"coinbase transaction %v from height "+
					"%v at height %v before required "+
					"maturity of %v blocks", txHash,
					txInHash, originHeight, txHeight,
					coinbaseMaturity)
				return 0, ruleError(ErrImmatureSpend, str)
			}
		}

		// Ensure that the transaction is not spending coins from a
		// transaction that included an expiry but which has not yet
		// reached coinbase maturity many blocks.
		if utxoEntry.HasExpiry() {
			originHeight := utxoEntry.BlockHeight()
			blocksSincePrev := txHeight - int64(originHeight) //TODO, remove type conversion
			if blocksSincePrev < coinbaseMaturity {
				str := fmt.Sprintf("tx %v tried to spend "+
					"transaction %v including an expiry "+
					"from height %v at height %v before "+
					"required maturity of %v blocks",
					txHash, txInHash, originHeight,
					txHeight, coinbaseMaturity)
				return 0, ruleError(ErrExpiryTxSpentEarly, str)
			}
		}

		// Ensure the transaction amounts are in range.  Each of the
		// output values of the input transactions must not be negative
		// or more than the max allowed per transaction.  All amounts
		// in a transaction are in a unit value known as an atom.  One
		// Coin is a quantity of atoms as defined by the AtomPerCoin
		// constant.
		originTxAtom := utxoEntry.AmountByIndex(originTxIndex)
		if originTxAtom < 0 {
			str := fmt.Sprintf("transaction output has negative "+
				"value of %v", originTxAtom)
			return 0, ruleError(ErrBadTxOutValue, str)
		}
		if originTxAtom > types.MaxAmount {
			str := fmt.Sprintf("transaction output value of %v is "+
				"higher than max allowed value of %v",
				originTxAtom, types.MaxAmount)
			return 0, ruleError(ErrBadTxOutValue, str)
		}

		// The total of all outputs must not be more than the max
		// allowed per transaction.  Also, we could potentially
		// overflow the accumulator so check for overflow.
		lastAtomIn := totalAtomIn
		totalAtomIn += int64(originTxAtom)  //TODO, remove type conversion
		if totalAtomIn < lastAtomIn ||
			totalAtomIn > types.MaxAmount {
			str := fmt.Sprintf("total value of all transaction "+
				"inputs is %v which is higher than max "+
				"allowed value of %v", totalAtomIn,
				types.MaxAmount)
			return 0, ruleError(ErrBadTxOutValue, str)
		}
	}

	// Calculate the total output amount for this transaction.  It is safe
	// to ignore overflow and out of range errors here because those error
	// conditions would have already been caught by checkTransactionSanity.
	var totalAtomOut int64
	for _, txOut := range tx.Transaction().TxOut {
		totalAtomOut += int64(txOut.Amount)  //TODO, remove type conversion
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
	b.chainLock.Lock()
	defer b.chainLock.Unlock()

	// Skip the proof of work check as this is just a block template.
	flags := BFNoPoWCheck

	// The block template must build off the current tip of the main chain
	// or its parent.

	// Perform context-free sanity checks on the block and its transactions.
	err := checkBlockSanity(block, b.timeSource, flags, b.params)
	if err != nil {
		return err
	}
	view := NewUtxoViewpoint()
	parentsSet:=dag.NewHashSet()
	parentsSet.AddList(block.Block().Parents)
	view.SetBestHash(b.bd.GetMainParent(parentsSet).GetHash())
	tipsNode:=[]*blockNode{}

	for _,v:=range block.Block().Parents{
		bn:=b.index.LookupNode(v)
		if bn!=nil {
			tipsNode=append(tipsNode,bn)
		}
	}
	if len(tipsNode)==0 {
		return ruleError(ErrPrevBlockNotBest, "tipsNode")
	}
	header := &block.Block().Header
	newNode := newBlockNode(header, tipsNode)
	newNode.order=block.Order()
	newNode.SetHeight(block.Height())
	err=b.checkConnectBlock(newNode, block, view, nil)
	if err!=nil{
		return err
	}
	badTxArr:=b.GetBadTxFromBlock(block.Hash())
	if len(badTxArr)>0 {
		str :=fmt.Sprintf("some bad transactions:")
		for _,v:=range badTxArr{
			str+="\n"
			str+=v.String()
		}
		return ruleError(ErrMissingTxOut, str)
	}
	return nil

	/*
	// The block must pass all of the validation rules which depend on the
	// position of the block within the block chain.

	err = b.checkBlockContext(block, prevNode, flags)
	if err != nil {
		return err
	}

	newNode := newBlockNode(&block.Block().Header, prevNode)

	// Use the ticket database as is when extending the main (best) chain.
	if prevNode.hash == tip.hash {
		// Grab the parent block since it is required throughout the block
		// connection process.
		parent, err := b.fetchMainChainBlockByHash(&prevNode.hash)
		if err != nil {
			return ruleError(ErrMissingParent, err.Error())
		}

		view := NewUtxoViewpoint()
		view.SetBestHash(&tip.hash)
		return b.checkConnectBlock(newNode, block, parent, view, nil)
	}

	// The requested node is either on a side chain or is a node on the
	// main chain before the end of it.  In either case, we need to undo
	// the transactions and spend information for the blocks which would be
	// disconnected during a reorganize to the point of view of the node
	// just before the requested node.
	detachNodes, attachNodes := b.getReorganizeNodes(prevNode)

	view := NewUtxoViewpoint()
	view.SetBestHash(&tip.hash)
	var nextBlockToDetach *types.SerializedBlock
	for e := detachNodes.Front(); e != nil; e = e.Next() {
		// Grab the block to detach based on the node.  Use the fact that the
		// parent of the block is already required, and the next block to detach
		// will also be the parent to optimize.
		n := e.Value.(*blockNode)
		block := nextBlockToDetach
		if block == nil {
			var err error
			block, err = b.fetchMainChainBlockByHash(&n.hash)
			if err != nil {
				return err
			}
		}
		if n.hash != *block.Hash() {
			panicf("detach block node hash %v (height %v) does not match "+
				"previous parent block hash %v", &n.hash, n.height,
				block.Hash())
		}

		parent, err := b.fetchMainChainBlockByHash(&n.parent.hash)
		if err != nil {
			return err
		}
		nextBlockToDetach = parent

		// Load all of the spent txos for the block from the spend journal.
		var stxos []spentTxOut
		//TODO, refactor the direct database access
		err = b.db.View(func(dbTx database.Tx) error {
			stxos, err = dbFetchSpendJournalEntry(dbTx, block, parent)
			return err
		})
		if err != nil {
			return err
		}

		err = b.disconnectTransactions(view, block, parent, stxos)
		if err != nil {
			return err
		}
	}

	// The UTXO viewpoint is now accurate to either the node where the
	// requested node forks off the main chain (in the case where the
	// requested node is on a side chain), or the requested node itself if
	// the requested node is an old node on the main chain.  Entries in the
	// attachNodes list indicate the requested node is on a side chain, so
	// if there are no nodes to attach, we're done.
	if attachNodes.Len() == 0 {
		// Grab the parent block since it is required throughout the block
		// connection process.
		parent, err := b.fetchMainChainBlockByHash(&prevNode.hash)
		if err != nil {
			return ruleError(ErrMissingParent, err.Error())
		}

		return b.checkConnectBlock(newNode, block, parent, view, nil)
	}

	// The requested node is on a side chain, so we need to apply the
	// transactions and spend information from each of the nodes to attach.
	var prevAttachBlock *types.SerializedBlock
	for e := attachNodes.Front(); e != nil; e = e.Next() {
		// Grab the block to attach based on the node.  Use the fact that the
		// parent of the block is either the fork point for the first node being
		// attached or the previous one that was attached for subsequent blocks
		// to optimize.
		n := e.Value.(*blockNode)
		block, err := b.fetchBlockByHash(&n.hash)
		if err != nil {
			return err
		}
		parent := prevAttachBlock
		if parent == nil {
			var err error
			parent, err = b.fetchMainChainBlockByHash(&n.parent.hash)
			if err != nil {
				return err
			}
		}
		if n.parent.hash != *parent.Hash() {
			panicf("attach block node hash %v (height %v) parent hash %v does "+
				"not match previous parent block hash %v", &n.hash, n.height,
				&n.parent.hash, parent.Hash())
		}

		// Store the loaded block for the next iteration.
		prevAttachBlock = block

		err = b.connectTransactions(view, block, parent, nil)
		if err != nil {
			return err
		}
	}

	// Grab the parent block since it is required throughout the block
	// connection process.
	parent, err := b.fetchBlockByHash(&prevNode.hash)
	if err != nil {
		return ruleError(ErrMissingParent, err.Error())
	}

	// Notice the spent txout details are not requested here and thus will not
	// be generated.  This is done because the state will not be written to the
	// database, so it is not needed.
	return b.checkConnectBlock(newNode, block, parent, view, nil)*/
}
