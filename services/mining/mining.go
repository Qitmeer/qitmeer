// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2014-2016 The btcsuite developers
// Copyright (c) 2016-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mining

import (
	"github.com/Qitmeer/qitmeer-lib/common/hash"
	"github.com/Qitmeer/qitmeer-lib/core/protocol"
	s "github.com/Qitmeer/qitmeer-lib/core/serialization"
	"github.com/Qitmeer/qitmeer-lib/core/types"
	"github.com/Qitmeer/qitmeer-lib/engine/txscript"
	"github.com/Qitmeer/qitmeer-lib/params"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"time"
)

const (

	// kilobyte is the size of a kilobyte.
	// TODO, refactor the location of kilobyte const
	kilobyte = 1000

	// blockHeaderOverhead is the max number of bytes it takes to serialize
	// a block header and max possible transaction count.
	// TODO, refactor the location of blockHeaderOverhead const
	blockHeaderOverhead = types.MaxBlockHeaderPayload + s.MaxVarIntPayload

	// coinbaseFlags is some extra data appended to the coinbase script
	// sig.
	// TODO, refactor the location of coinbaseFlags const
	CoinbaseFlags = "/qitmeer/"

	// generatedBlockVersion is the version of the block being generated for
	// the main network.  It is defined as a constant here rather than using
	// the wire.BlockVersion constant since a change in the block version
	// will require changes to the generated block.  Using the wire constant
	// for generated block version could allow creation of invalid blocks
	// for the updated version.
	// TODO, refactor the location of generatedBlockVersion const
	GeneratedBlockVersion = 0

	// generatedBlockVersionTest is the version of the block being generated
	// for networks other than the main network.
	// TODO, refactor the location of generatedBlockVersionTest const
	GeneratedBlockVersionTest = 6

)

// TxSource represents a source of transactions to consider for inclusion in
// new blocks.
//
// The interface contract requires that all of these methods are safe for
// concurrent access with respect to the source.
type TxSource interface {
	// LastUpdated returns the last time a transaction was added to or
	// removed from the source pool.
	LastUpdated() time.Time

	// MiningDescs returns a slice of mining descriptors for all the
	// transactions in the source pool.
	MiningDescs() []*types.TxDesc

	// HaveTransaction returns whether or not the passed transaction hash
	// exists in the source pool.
	HaveTransaction(hash *hash.Hash) bool

	// HaveAllTransactions returns whether or not all of the passed
	// transaction hashes exist in the source pool.
	HaveAllTransactions(hashes []hash.Hash) bool

}

// Allowed timestamp for a block building on the end of the provided best chain.
func MinimumMedianTime(bc *blockchain.BlockChain) time.Time {
	mainTip:=bc.BlockIndex().LookupNode(bc.BlockDAG().GetMainChainTip().GetHash())
	mainTipTime:=time.Unix(mainTip.GetTimestamp(), 0)
	return mainTipTime.Add(time.Second)
}

// medianAdjustedTime returns the current time adjusted
func MedianAdjustedTime(bc *blockchain.BlockChain, timeSource blockchain.MedianTimeSource) time.Time {
	newTimestamp := timeSource.AdjustedTime()
	minTimestamp := MinimumMedianTime(bc)
	if newTimestamp.Before(minTimestamp) {
		newTimestamp = minTimestamp
	}

	return newTimestamp
}

func standardCoinbaseScript(nextBlockHeight uint64, extraNonce uint64) ([]byte, error) {
	return txscript.NewScriptBuilder().AddInt64(int64(nextBlockHeight)).
		AddInt64(int64(extraNonce)).AddData([]byte(CoinbaseFlags)).
		Script()
}

// standardCoinbaseOpReturn creates a standard OP_RETURN output to insert into
// coinbase to use as extranonces. The OP_RETURN pushes 32 bytes.
func standardCoinbaseOpReturn(enData []byte) ([]byte, error) {
	if len(enData) == 0 {
		return nil,nil
	}
	extraNonceScript, err := txscript.GenerateProvablyPruneableOut(enData)
	if err != nil {
		return nil, err
	}
	return extraNonceScript, nil
}


// createCoinbaseTx returns a coinbase transaction paying an appropriate subsidy
// based on the passed block height to the provided address.  When the address
// is nil, the coinbase transaction will instead be redeemable by anyone.
//
// See the comment for NewBlockTemplate for more information about why the nil
// address handling is useful.
func createCoinbaseTx(subsidyCache *blockchain.SubsidyCache, coinbaseScript []byte, opReturnPkScript []byte, nextBlockHeight int64, addr types.Address, params *params.Params) (*types.Tx, error) {
	tx := types.NewTransaction()
	tx.AddTxIn(&types.TxInput{
		// Coinbase transactions have no inputs, so previous outpoint is
		// zero hash and max index.
		PreviousOut: *types.NewOutPoint(&hash.Hash{},
			types.MaxPrevOutIndex ),
		Sequence:        types.MaxTxInSequenceNum,
		SignScript:      coinbaseScript,
	})

	hasTax:=false
	if params.BlockTaxProportion > 0 &&
		len(params.OrganizationPkScript) > 0{
		hasTax=true
	}
	// Create a coinbase with correct block subsidy and extranonce.
	subsidy := blockchain.CalcBlockWorkSubsidy(subsidyCache,
		nextBlockHeight, params)
	tax := blockchain.CalcBlockTaxSubsidy(subsidyCache,
		nextBlockHeight, params)

	// output
	// Create the script to pay to the provided payment address if one was
	// specified.  Otherwise create a script that allows the coinbase to be
	// redeemable by anyone.
	var pksSubsidy []byte
	var err error
	if addr != nil {
		pksSubsidy, err = txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}
	} else {
		scriptBuilder := txscript.NewScriptBuilder()
		pksSubsidy, err = scriptBuilder.AddOp(txscript.OP_TRUE).Script()
		if err != nil {
			return nil, err
		}
	}
	if !hasTax {
		subsidy+=uint64(tax)
		tax=0
	}
	// Subsidy paid to miner.
	tx.AddTxOut(&types.TxOutput{
		Amount:   subsidy,
		PkScript: pksSubsidy,
	})

	// Tax output.
	if hasTax {
		tx.AddTxOut(&types.TxOutput{
				Amount:    uint64(tax),
				PkScript: params.OrganizationPkScript,
			})
	}
	// nulldata.
	if opReturnPkScript != nil {
		tx.AddTxOut(&types.TxOutput{
			Amount:    0,
			PkScript: opReturnPkScript,
		})
	}
	return types.NewTx(tx), nil
}

func BlockVersion(net protocol.Network) uint32  {
	blockVersion := uint32(GeneratedBlockVersion)
	if net != protocol.MainNet {
		blockVersion = GeneratedBlockVersionTest
	}
	return blockVersion
}
