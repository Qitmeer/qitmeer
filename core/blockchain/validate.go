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
	"github.com/Qitmeer/qitmeer/core/blockchain/opreturn"
	"github.com/Qitmeer/qitmeer/core/blockchain/token"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/core/merkle"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/params"
	"math"
	"time"
)

const (
	// The index of coinbase output for subsidy
	CoinbaseOutput_subsidy = 0
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

	// A block must have at least one regular transaction.
	numTx := len(msgBlock.Transactions)
	if numTx == 0 {
		return ruleError(ErrNoTransactions, "block does not contain "+
			"any transactions")
	}

	height, err := ExtractCoinbaseHeight(block.Block().Transactions[0])
	if err != nil {
		return err
	}

	// TODO this hard code fork
	if !chainParams.CoinbaseConfig.CheckVersion(int64(height), block.Block().Transactions[0].TxIn[0].GetSignScript()) {
		return ruleError(ErrorCoinbaseBlockVersion, "block coinbase version error")
	}

	err = checkBlockHeaderSanity(header, timeSource, flags, chainParams, uint(height))
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
	for _, tx := range transactions {
		if !b.IsValidTxType(types.DetermineTxType(tx.Tx)) {
			errStr := fmt.Sprintf("%s is not support transaction type.", types.DetermineTxType(tx.Tx).String())
			return ruleError(ErrIrregTxInRegularTree, errStr)
		}
		// A block must not have stake transactions in the regular
		// transaction tree.
		err := CheckTransactionSanity(tx.Transaction(), chainParams)
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

	// Build merkle tree and ensure the calculated merkle root matches the
	// entry in the block header.  This also has the effect of caching all
	// of the token balance hashes in the block to speed up future hash
	// checks.
	calculatedTokenStateRoot := b.CalculateTokenStateRoot(block.Transactions(), msgBlock.Parents)
	if !header.StateRoot.IsEqual(&calculatedTokenStateRoot) {
		str := fmt.Sprintf("block merkle state root is invalid - block "+
			"header indicates %s, but calculated value is %s",
			header.StateRoot, calculatedTokenStateRoot)
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
func checkBlockHeaderSanity(header *types.BlockHeader, timeSource MedianTimeSource, flags BehaviorFlags, chainParams *params.Params, mHeight uint) error {
	instance := pow.GetInstance(header.Pow.GetPowType(), 0, []byte{})
	instance.SetMainHeight(pow.MainHeight(mHeight))
	instance.SetParams(chainParams.PowConfig)
	if !instance.CheckAvailable() {
		str := fmt.Sprintf("pow type : %d is not available!", header.Pow.GetPowType())
		return ruleError(ErrInValidPowType, str)
	}
	// Ensure the proof of work bits in the block header is in min/max
	// range and the block hash is less than the target value described by
	// the bits.
	err := checkProofOfWork(header, chainParams.PowConfig, flags, mHeight)
	if err != nil {
		return ruleError(ErrInvalidPow, err.Error())
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
func checkProofOfWork(header *types.BlockHeader, powConfig *pow.PowConfig, flags BehaviorFlags, mHeight uint) error {

	// The block hash must be less than the claimed target unless the flag
	// to avoid proof of work checks is set.
	if flags&BFNoPoWCheck != BFNoPoWCheck {
		header.Pow.SetParams(powConfig)
		header.Pow.SetMainHeight(pow.MainHeight(mHeight))
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

	if types.IsTokenTx(tx) {
		update, err := token.NewUpdateFromTx(tx)
		if err != nil {
			return err
		}
		return update.CheckSanity()
	}

	// Ensure the transaction amounts are in range.  Each transaction
	// output must not be negative or more than the max allowed per
	// transaction.  Also, the total of all outputs must abide by the same
	// restrictions.  All amounts in a transaction are in a unit value
	// known as an atom.  One Coin is a quantity of atoms as defined by
	// the AtomsPerCoin constant.
	totalAtom := make(map[types.CoinID]int64)
	for _, txOut := range tx.TxOut {
		atom := txOut.Amount
		if atom.Value > types.MaxAmount {
			str := fmt.Sprintf("transaction output value of %v is "+
				"higher than max allowed value of %v", atom,
				types.MaxAmount)
			return ruleError(ErrInvalidTxOutValue, str)
		}

		// Two's complement int64 overflow guarantees that any overflow
		// is detected and reported.
		// TODO revisit the overflow check
		totalAtom[atom.Id] += atom.Value
		if totalAtom[atom.Id] < 0 {
			str := fmt.Sprintf("total value of all transaction "+
				"outputs exceeds max allowed value of %v",
				types.MaxAmount)
			return ruleError(ErrInvalidTxOutValue, str)
		}
		if totalAtom[atom.Id] > types.MaxAmount {
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
		err := validateCoinbase(tx, params)
		if err != nil {
			return err
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

func validateCoinbase(tx *types.Transaction, pa *params.Params) error {
	slen := len(tx.TxIn[0].SignScript)
	if slen < MinCoinbaseScriptLen || slen > MaxCoinbaseScriptLen {
		str := fmt.Sprintf("coinbase transaction script "+
			"length of %d is out of range (min: %d, max: "+
			"%d)", slen, MinCoinbaseScriptLen,
			MaxCoinbaseScriptLen)
		return ruleError(ErrBadCoinbaseScriptLen, str)
	}
	if pa.HasTax() {
		err := validateCoinbaseTax(tx, pa)
		if err != nil {
			return err
		}
	} else {
		if len(tx.TxOut) <= CoinbaseOutput_subsidy+1 {
			str := fmt.Sprintf("Coinbase output number error")
			return ruleError(ErrBadCoinbaseOutpoint, str)
		}
		if tx.TxOut[CoinbaseOutput_subsidy].Amount.Id != types.MEERID {
			str := fmt.Sprintf("Subsidy output amount type is error")
			return ruleError(ErrBadCoinbaseOutpoint, str)
		}
		endIndex := len(tx.TxOut) - 1
		if !opreturn.IsOPReturn(tx.TxOut[endIndex].PkScript) {
			str := fmt.Sprintf("TxOutput(%d) must coinbase op return type", endIndex)
			return ruleError(ErrBadCoinbaseOutpoint, str)
		}
		opr, err := opreturn.NewOPReturnFrom(tx.TxOut[endIndex].PkScript)
		if err != nil {
			return err
		}
		err = opr.Verify(tx)
		if err != nil {
			return err
		}

		err = validateCoinbaseToken(tx.TxOut[1:endIndex])
		if err != nil {
			return err
		}
	}
	return nil
}

func validateCoinbaseToken(outputs []*types.TxOutput) error {
	if len(outputs) <= 0 {
		return nil
	}
	for _, v := range outputs {
		if v.Amount.Id == types.MEERID {
			continue
		}
		if !types.IsKnownCoinID(v.Amount.Id) {
			str := fmt.Sprintf("Unknown coin %s", v.Amount.Id.Name())
			return ruleError(ErrBadCoinbaseOutpoint, str)
		}
		if v.Amount.Value != 0 {
			str := fmt.Sprintf("You don't have permission to change the consensus balance: %d", v.Amount.Value)
			return ruleError(ErrBadCoinbaseOutpoint, str)
		}
	}
	return nil
}

// Validate the tax in coinbase transaction. Prevent miners from attacking.
func validateCoinbaseTax(tx *types.Transaction, pa *params.Params) error {
	if len(tx.TxOut) <= CoinbaseOutput_subsidy+2 {
		str := fmt.Sprintf("Lack of output")
		return ruleError(ErrBadCoinbaseOutpoint, str)
	}
	if tx.TxOut[CoinbaseOutput_subsidy].Amount.Id != types.MEERID {
		str := fmt.Sprintf("Subsidy output amount type is error")
		return ruleError(ErrBadCoinbaseOutpoint, str)
	}
	endIndex := len(tx.TxOut) - 1
	taxIndex := endIndex - 1
	if !opreturn.IsOPReturn(tx.TxOut[endIndex].PkScript) {
		str := fmt.Sprintf("TxOutput(%d) must coinbase op return type", endIndex)
		return ruleError(ErrBadCoinbaseOutpoint, str)
	}

	opr, err := opreturn.NewOPReturnFrom(tx.TxOut[endIndex].PkScript)
	if err != nil {
		return err
	}
	err = opr.Verify(tx)
	if err != nil {
		return err
	}

	if tx.TxOut[taxIndex].Amount.Id != types.MEERID {
		str := fmt.Sprintf("Tax output amount type is error")
		return ruleError(ErrBadCoinbaseOutpoint, str)
	}
	err = validateCoinbaseToken(tx.TxOut[1:taxIndex])
	if err != nil {
		return err
	}
	orgPkScriptStr := hex.EncodeToString(pa.OrganizationPkScript)
	curPkScriptStr := hex.EncodeToString(tx.TxOut[taxIndex].PkScript)
	if orgPkScriptStr != curPkScriptStr {
		str := fmt.Sprintf("coinbase transaction for block pays to %s, but it is %s",
			orgPkScriptStr, curPkScriptStr)
		return ruleError(ErrBadCoinbaseValue, str)
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
func (b *BlockChain) checkBlockContext(block *types.SerializedBlock, mainParent blockdag.IBlock, flags BehaviorFlags) error {
	// The genesis block is valid by definition.
	if mainParent == nil {
		return nil
	}
	if !mainParent.GetHash().IsEqual(block.Block().Parents[0]) {
		return fmt.Errorf("Main parent (%s) is inconsistent in block (%s)\n", mainParent.GetHash().String(), block.Block().Parents[0].String())
	}
	prevBlock := mainParent

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
		maxBlockSize, err := b.maxBlockSize()
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

		err = b.checkBlockSubsidy(block)
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

func (b *BlockChain) checkBlockSubsidy(block *types.SerializedBlock) error {
	bi := b.bd.GetBlueInfoByHash(block.Block().Parents[0])
	// check subsidy
	transactions := block.Transactions()
	subsidy := b.subsidyCache.CalcBlockSubsidy(bi)
	workAmountOut := int64(0)
	txoutLen := len(transactions[0].Tx.TxOut)
	hasOPR := opreturn.IsOPReturn(transactions[0].Tx.TxOut[txoutLen-1].PkScript)
	for k, v := range transactions[0].Tx.TxOut {
		// the coinbase should always use meer coin
		if v.Amount.Id != types.MEERID {
			continue
		}
		if b.params.HasTax() {
			if hasOPR {
				if k == txoutLen-2 {
					continue
				}
				if k == txoutLen-1 {
					continue
				}
			} else {
				if k == txoutLen-1 {
					continue
				}
			}

		}
		workAmountOut += v.Amount.Value
	}

	var work int64
	var tax int64
	var taxAmountOut int64 = 0
	var taxOutput *types.TxOutput
	var totalAmountOut int64 = 0

	if b.params.HasTax() {
		work = int64(CalcBlockWorkSubsidy(b.subsidyCache, bi, b.params))
		tax = int64(CalcBlockTaxSubsidy(b.subsidyCache, bi, b.params))
		taxOutput = transactions[0].Tx.TxOut[len(transactions[0].Tx.TxOut)-1]

		taxAmountOut = taxOutput.Amount.Value
	} else {
		work = subsidy
		tax = 0
		taxAmountOut = 0
	}

	totalAmountOut = workAmountOut + taxAmountOut

	if totalAmountOut != subsidy {
		str := fmt.Sprintf("coinbase transaction for block pays %v which is not the subsidy %v",
			totalAmountOut, subsidy)
		return ruleError(ErrBadCoinbaseValue, str)
	}

	if workAmountOut != work ||
		tax != taxAmountOut {
		str := fmt.Sprintf("coinbase transaction for block pays %d  %d  which is not the %d  %d",
			workAmountOut, taxAmountOut, work, tax)
		return ruleError(ErrBadCoinbaseValue, str)
	}

	if b.params.HasTax() {
		orgPkScriptStr := hex.EncodeToString(b.params.OrganizationPkScript)
		curPkScriptStr := hex.EncodeToString(taxOutput.PkScript)
		if orgPkScriptStr != curPkScriptStr {
			str := fmt.Sprintf("coinbase transaction for block pays to %s, but it is %s",
				orgPkScriptStr, curPkScriptStr)
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
func (b *BlockChain) checkBlockHeaderContext(block *types.SerializedBlock, prevNode blockdag.IBlock, flags BehaviorFlags) error {
	// The genesis block is valid by definition.
	if prevNode == nil {
		return nil
	}

	header := &block.Block().Header
	fastAdd := flags&BFFastAdd == BFFastAdd
	if !fastAdd {
		instance := pow.GetInstance(header.Pow.GetPowType(), 0, []byte{})
		instance.SetMainHeight(pow.MainHeight(prevNode.GetHeight() + 1))
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
		medianTime := b.CalcPastMedianTime(prevNode)
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
	parents := blockdag.NewIdSet()
	for _, v := range block.Block().Parents {
		parents.Add(b.bd.GetBlockId(v))
	}
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
	if checkpointNode != nil && blockLayer < checkpointNode.GetLayer() {
		str := fmt.Sprintf("block at layer %d forks the main chain "+
			"before the previous checkpoint at layer %d",
			blockLayer, checkpointNode.GetLayer())
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
func (b *BlockChain) checkConnectBlock(ib blockdag.IBlock, block *types.SerializedBlock, utxoView *UtxoViewpoint, stxos *[]SpentTxOut) error {
	// If the side chain blocks end up in the database, a call to
	// CheckBlockSanity should be done here in case a previous version
	// allowed a block that is no longer valid.  However, since the
	// implementation only currently uses memory for the side chain blocks,
	// it isn't currently necessary.

	// The coinbase for the Genesis block is not spendable, so just return
	// an error now.
	if ib.GetHash().IsEqual(b.params.GenesisHash) {
		str := "the coinbase for the genesis block is not spendable"
		return ruleError(ErrMissingTxOut, str)
	}
	// Don't run scripts if this node is before the latest known good
	// checkpoint since the validity is verified via the checkpoints (all
	// transactions are included in the merkle root hash and any changes
	// will therefore be detected by the next checkpoint).  This is a huge
	// optimization because running the scripts is the most time consuming
	// portion of block handling.
	checkpoint := b.LatestCheckpoint()
	runScripts := !b.noVerify
	if checkpoint != nil && uint64(ib.GetLayer()) <= checkpoint.Layer {
		runScripts = false
	}
	var scriptFlags txscript.ScriptFlags
	var err error
	if runScripts {
		scriptFlags, err = b.consensusScriptVerifyFlags()
		if err != nil {
			return err
		}
	}

	// At first, we must calculate the dag duplicate tx for block.
	b.CalculateDAGDuplicateTxs(block)

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

	node := b.GetBlockNode(ib)
	if node == nil {
		return fmt.Errorf("Block Node error:%s\n", ib.GetHash().String())
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
	mainParent := b.bd.GetBlockById(ib.GetMainParent())
	if mainParent == nil {
		return fmt.Errorf("Block Main Parent error:%s\n", ib.GetHash().String())
	}
	prevMedianTime := b.CalcPastMedianTime(mainParent)

	// Skip the coinbase since it does not have any inputs and thus
	// lock times do not apply.
	for _, tx := range block.Transactions() {
		sequenceLock, err := b.calcSequenceLock(tx,
			utxoView, false)
		if err != nil {
			return err
		}

		if !SequenceLockActive(sequenceLock, int64(ib.GetHeight()), //TODO, remove type conversion
			prevMedianTime) {

			str := fmt.Sprintf("block contains " +
				"transaction whose input sequence " +
				"locks are not met")
			return ruleError(ErrUnfinalizedTx, str)
		}
	}

	if runScripts {
		err = b.checkBlockScripts(block, utxoView,
			scriptFlags, b.sigCache)
		if err != nil {
			log.Trace("checkBlockScripts failed; error returned "+
				"on txtreeregular of cur block: %v", err)
			return err
		}
	}

	return b.CheckTokenState(block)
}

// consensusScriptVerifyFlags returns the script flags that must be used when
// executing transaction scripts to enforce the consensus rules. This includes
// any flags required as the result of any agendas that have passed and become
// active.
func (b *BlockChain) consensusScriptVerifyFlags() (txscript.ScriptFlags, error) {
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
func (b *BlockChain) checkTransactionsAndConnect(node *BlockNode, block *types.SerializedBlock, subsidyCache *SubsidyCache, utxoView *UtxoViewpoint, stxos *[]SpentTxOut) error {
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

	totalFees := types.AmountMap{}
	for idx, tx := range transactions {
		if tx.IsDuplicate {
			if tx.Tx.IsCoinBase() {
				return ruleError(ErrDuplicateTx, fmt.Sprintf("Coinbase Tx(%s) is duplicate in block(%s)", tx.Hash().String(), block.Hash().String()))
			}
			continue
		}
		if types.IsTokenTx(tx.Tx) {
			if types.IsTokenMintTx(tx.Tx) {
				err := b.CheckTokenTransactionInputs(tx, utxoView)
				if err != nil {
					return err
				}
				err = utxoView.connectTransaction(tx, node, uint32(idx), stxos, b)
				if err != nil {
					return err
				}
			}
			continue
		}
		txFee, err := b.CheckTransactionInputs(tx, utxoView)
		if err != nil {
			return err
		}
		// Sum the total fees and ensure we don't overflow the
		// accumulator.
		for _, coinId := range types.CoinIDList {
			lastTotalFees := totalFees[coinId]
			totalFees[coinId] += txFee[coinId]
			if totalFees[coinId] < lastTotalFees {
				return ruleError(ErrBadFees, "total fees for block "+
					"overflows accumulator")
			}

		}

		err = utxoView.connectTransaction(tx, node, uint32(idx), stxos, b)
		if err != nil {
			return err
		}
	}
	return b.checkBlockSubsidy(block)
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
func (b *BlockChain) CheckTransactionInputs(tx *types.Tx, utxoView *UtxoViewpoint) (types.AmountMap, error) {
	msgTx := tx.Transaction()

	txHash := tx.Hash()
	totalAtomIn := make(map[types.CoinID]int64)

	// Coinbase transactions have no inputs.
	if msgTx.IsCoinBase() {
		return nil, nil
	}
	bd := b.bd
	// -------------------------------------------------------------------
	// General transaction testing.
	// -------------------------------------------------------------------
	targets := []uint{}

	for idx, txIn := range msgTx.TxIn {
		utxoEntry := utxoView.LookupEntry(txIn.PreviousOut)
		if utxoEntry == nil || utxoEntry.IsSpent() {
			str := fmt.Sprintf("output %v referenced from "+
				"transaction %s:%d either does not exist or "+
				"has already been spent", txIn.PreviousOut,
				txHash, idx)
			return nil, ruleError(ErrMissingTxOut, str)
		}

		// Ensure the coinId is known
		err := types.CheckCoinID(utxoEntry.amount.Id)
		if err != nil {
			return nil, err
		}

		// Ensure the transaction is not spending coins which have not
		// yet reached the required coinbase maturity.

		// Ensure the transaction amounts are in range.  Each of the
		// output values of the input transactions must not be negative
		// or more than the max allowed per transaction.  All amounts
		// in a transaction are in a unit value known as an atom.  One
		// Coin is a quantity of atoms as defined by the AtomPerCoin
		// constant.
		originTxAtom := utxoEntry.Amount()

		if utxoEntry.IsCoinBase() {
			ubhIB := bd.GetBlock(utxoEntry.BlockHash())
			if ubhIB == nil {
				str := fmt.Sprintf("utxoEntry blockhash error:%s", utxoEntry.BlockHash())
				return nil, ruleError(ErrNoViewpoint, str)
			}
			targets = append(targets, ubhIB.GetID())
			if !utxoEntry.BlockHash().IsEqual(b.params.GenesisHash) {
				if originTxAtom.Id == types.MEERID {
					if txIn.PreviousOut.OutIndex == CoinbaseOutput_subsidy {
						originTxAtom.Value += b.GetFeeByCoinID(utxoEntry.BlockHash(), originTxAtom.Id)
					}
				} else {
					originTxAtom.Value = b.GetFeeByCoinID(utxoEntry.BlockHash(), originTxAtom.Id)
				}
			}
		}
		if originTxAtom.Value < 0 {
			str := fmt.Sprintf("transaction output has negative "+
				"value of %v", originTxAtom)
			return nil, ruleError(ErrInvalidTxOutValue, str)
		}
		if originTxAtom.Value > types.MaxAmount {
			str := fmt.Sprintf("transaction output value of %v is "+
				"higher than max allowed value of %v",
				originTxAtom, types.MaxAmount)
			return nil, ruleError(ErrInvalidTxOutValue, str)
		}

		// The total of all outputs must not be more than the max
		// allowed per transaction.  Also, we could potentially
		// overflow the accumulator so check for overflow.
		lastAtomIn := totalAtomIn
		totalAtomIn[originTxAtom.Id] += originTxAtom.Value
		if totalAtomIn[originTxAtom.Id] < lastAtomIn[originTxAtom.Id] ||
			totalAtomIn[originTxAtom.Id] > types.MaxAmount {
			str := fmt.Sprintf("total value of all transaction "+
				"inputs is %v which is higher than max "+
				"allowed value of %v", totalAtomIn,
				types.MaxAmount)
			return nil, ruleError(ErrInvalidTxOutValue, str)
		}
	}

	if len(targets) > 0 {
		viewpoints := []uint{}
		for _, blockHash := range utxoView.viewpoints {
			vIB := bd.GetBlock(blockHash)
			if vIB != nil {
				viewpoints = append(viewpoints, vIB.GetID())
			}
		}
		if len(viewpoints) == 0 {
			str := fmt.Sprintf("transaction %s has no viewpoints", txHash)
			return nil, ruleError(ErrNoViewpoint, str)
		}
		err := bd.CheckBlueAndMatureMT(targets, viewpoints, uint(b.params.CoinbaseMaturity))
		if err != nil {
			return nil, ruleError(ErrImmatureSpend, err.Error())
		}
	}

	// Calculate the total output amount for this transaction.  It is safe
	// to ignore overflow and out of range errors here because those error
	// conditions would have already been caught by checkTransactionSanity.
	totalAtomOut := make(map[types.CoinID]int64)
	for _, txOut := range tx.Transaction().TxOut {
		// Ensure the coinId is known
		err := types.CheckCoinID(txOut.Amount.Id)
		if err != nil {
			return nil, err
		}
		totalAtomOut[txOut.Amount.Id] += txOut.Amount.Value
	}

	// Ensure no unbalanced/unknowned coin type from input/output
	if len(totalAtomIn) != len(totalAtomOut) ||
		len(totalAtomIn) > len(types.CoinIDList) ||
		len(totalAtomOut) > len(types.CoinIDList) {
		return nil, fmt.Errorf("transaction contains unknown or unbalanced coin types")
	}

	allFees := make(map[types.CoinID]int64)
	for _, coinId := range types.CoinIDList {
		atomin, okin := totalAtomIn[coinId]
		atomout, okout := totalAtomOut[coinId]
		if !okin && !okout {
			continue
		} else if !(okin && okout) {
			str := fmt.Sprintf("transaction output CoinID does not match input. (%s)", coinId.Name())
			return nil, ruleError(ErrInvalidTxOutValue, str)
		}

		// Ensure the transaction does not spend more than its inputs.
		if atomin < atomout {
			str := fmt.Sprintf("total %s value of all transaction inputs for "+
				"transaction %v is %v which is less than the amount "+
				"spent of %v", coinId.Name(), txHash, atomin, atomout)
			return nil, ruleError(ErrSpendTooHigh, str)
		}
		allFees[coinId] = atomin - atomout
	}
	state := b.GetTokenState(b.TokenTipID)
	if state == nil {
		return nil, fmt.Errorf("No token sate:%d\n", b.TokenTipID)
	}
	err := state.CheckFees(allFees)
	if err != nil {
		return nil, err
	}
	return allFees, nil
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

	newNode := NewBlockNode(block, block.Block().Parents)
	virBlock := b.bd.CreateVirtualBlock(newNode)
	if virBlock == nil {
		return ruleError(ErrPrevBlockNotBest, "tipsNode")
	}
	virBlock.SetOrder(uint(block.Order()))
	if virBlock.GetHeight() != block.Height() {
		return ruleError(ErrPrevBlockNotBest, "tipsNode height")
	}
	mainParent := b.bd.GetBlockById(virBlock.GetMainParent())
	if mainParent == nil {
		return ruleError(ErrPrevBlockNotBest, "main parent")
	}

	err = b.checkBlockContext(block, mainParent, flags)
	if err != nil {
		return err
	}

	view := NewUtxoViewpoint()
	view.SetViewpoints(block.Block().Parents)

	err = b.checkConnectBlock(virBlock, block, view, nil)
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

func (b *BlockChain) CheckTokenTransactionInputs(tx *types.Tx, utxoView *UtxoViewpoint) error {
	msgTx := tx.Transaction()
	totalAtomIn := int64(0)
	targets := []uint{}

	for idx, txIn := range msgTx.TxIn {
		if idx == 0 {
			continue
		}
		utxoEntry := utxoView.LookupEntry(txIn.PreviousOut)
		if utxoEntry == nil || utxoEntry.IsSpent() {
			str := fmt.Sprintf("output %v referenced from "+
				"transaction %s:%d either does not exist or "+
				"has already been spent", txIn.PreviousOut,
				tx.Hash(), idx)
			return ruleError(ErrMissingTxOut, str)
		}
		if !utxoEntry.amount.Id.IsBase() {
			return fmt.Errorf("Token transaction(%s) input (%s %d) must be MEERID\n", tx.Hash(), txIn.PreviousOut.Hash, txIn.PreviousOut.OutIndex)
		}

		originTxAtom := utxoEntry.Amount()
		if originTxAtom.Value < 0 {
			str := fmt.Sprintf("transaction output has negative "+
				"value of %v", originTxAtom)
			return ruleError(ErrInvalidTxOutValue, str)
		}
		if originTxAtom.Value > types.MaxAmount {
			str := fmt.Sprintf("transaction output value of %v is "+
				"higher than max allowed value of %v",
				originTxAtom, types.MaxAmount)
			return ruleError(ErrInvalidTxOutValue, str)
		}

		if utxoEntry.IsCoinBase() {
			ubhIB := b.bd.GetBlock(utxoEntry.BlockHash())
			if ubhIB == nil {
				str := fmt.Sprintf("utxoEntry blockhash error:%s", utxoEntry.BlockHash())
				return ruleError(ErrNoViewpoint, str)
			}
			targets = append(targets, ubhIB.GetID())
			if !utxoEntry.BlockHash().IsEqual(b.params.GenesisHash) {
				if originTxAtom.Id == types.MEERID {
					if txIn.PreviousOut.OutIndex == CoinbaseOutput_subsidy {
						originTxAtom.Value += b.GetFeeByCoinID(utxoEntry.BlockHash(), originTxAtom.Id)
					}
				} else {
					originTxAtom.Value = b.GetFeeByCoinID(utxoEntry.BlockHash(), originTxAtom.Id)
				}
			}
		}

		totalAtomIn += originTxAtom.Value
	}

	lockMeer := int64(dbnamespace.ByteOrder.Uint64(msgTx.TxIn[0].PreviousOut.Hash[0:8]))
	if totalAtomIn != lockMeer {
		return fmt.Errorf("Utxo (%d) and input amount (%d) are inconsistent\n", totalAtomIn, lockMeer)
	}

	//
	if len(targets) > 0 {
		viewpoints := []uint{}
		for _, blockHash := range utxoView.viewpoints {
			vIB := b.bd.GetBlock(blockHash)
			if vIB != nil {
				viewpoints = append(viewpoints, vIB.GetID())
			}
		}
		if len(viewpoints) == 0 {
			str := fmt.Sprintf("transaction %s has no viewpoints", tx.Hash())
			return ruleError(ErrNoViewpoint, str)
		}
		err := b.bd.CheckBlueAndMatureMT(targets, viewpoints, uint(b.params.CoinbaseMaturity))
		if err != nil {
			return ruleError(ErrImmatureSpend, err.Error())
		}
	}
	//

	totalAtomOut := int64(0)
	state := b.GetTokenState(b.TokenTipID)
	if state == nil {
		return fmt.Errorf("Token state error\n")
	}
	coinId := msgTx.TxOut[0].Amount.Id
	tt, ok := state.Types[coinId]
	if !ok {
		return fmt.Errorf("It doesn't exist: Coin id (%d)\n", coinId)
	}
	tokenAmount := int64(0)
	tb, ok := state.Balances[coinId]
	if ok {
		tokenAmount = tb.Balance
	}

	for idx, txOut := range tx.Transaction().TxOut {
		if txOut.Amount.Id != coinId {
			return fmt.Errorf("Transaction(%s) output(%d) coin id is invalid\n", tx.Hash(), idx)
		}
		totalAtomOut += txOut.Amount.Value
	}
	if totalAtomOut+tokenAmount > int64(tt.UpLimit) {
		return fmt.Errorf("Token transaction mint (%d) exceeds the maximum (%d)\n", totalAtomOut, tt.UpLimit)
	}

	return nil
}
