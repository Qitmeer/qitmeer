// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package testutils

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"testing"
)

// GenerateBlock will generate a number of blocks by the input number for
// the appointed test harness.
// It will return the hashes of the generated blocks or an error
func GenerateBlock(t *testing.T, h *Harness, num uint64) []*hash.Hash {
	result := make([]*hash.Hash, 0)
	if blocks, err := h.Client.Generate(num); err != nil {
		t.Errorf("generate block failed : %v", err)
		return nil
	} else {
		for _, b := range blocks {
			result = append(result, b)
			t.Logf("%v: generate block [%v] ok", h.Node.Id(), b)
		}
	}
	return result
}

// AssertBlockOrderAndHeight will verify the current block order, total block number
// and current main-chain height of the appointed test harness and assert it ok or
// cause the test failed.
func AssertBlockOrderAndHeight(t *testing.T, h *Harness, order, total, height uint64) {
	// order
	if c, err := h.Client.BlockCount(); err != nil {
		t.Errorf("test failed : %v", err)
	} else {
		expect := order
		if c != expect {
			t.Errorf("test failed, expect %v , but got %v", expect, c)
		}
	}
	// total block
	if tal, err := h.Client.BlockTotal(); err != nil {
		t.Errorf("test failed : %v", err)
	} else {
		expect := total
		if tal != expect {
			t.Errorf("test failed, expect %v , but got %v", expect, tal)
		}
	}
	// main height
	if h, err := h.Client.MainHeight(); err != nil {
		t.Errorf("test failed : %v", err)
	} else {
		expect := height
		if h != expect {
			t.Errorf("test failed, expect %v , but got %v", expect, h)
		}
	}
}

// Spend amount from the wallet of the test harness and return tx hash
func Spend(t *testing.T, h *Harness, amt types.Amount) *hash.Hash {
	addr, err := h.Wallet.newAddress()
	if err != nil {
		t.Fatalf("failed to generate new address for test wallet: %v", err)
	}
	t.Logf("test wallet generated new address %v ok", addr.Encode())
	addrScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		t.Fatalf("failed to generated addr script: %v", err)
	}
	output := types.NewTxOutput(amt, addrScript)

	feeRate := types.Amount{10, amt.Id}
	txId, err := h.Wallet.PayAndSend([]*types.TxOutput{output}, feeRate)
	if err != nil {
		t.Fatalf("failed to pay the output: %v", err)
	}
	return txId
}

// TODO, order and height not work for the SerializedBlock
func AssertTxMined(t *testing.T, h *Harness, txId *hash.Hash, blockHash *hash.Hash) {
	block, err := h.Client.GetBlock(blockHash)
	if err != nil {
		t.Fatalf("failed to find block by hash %x : %v", blockHash, err)
	}
	numBlockTxns := len(block.Transactions())
	if numBlockTxns < 2 {
		t.Fatalf("the tx has not been mined, the block should at least 2 tx, but got %v", numBlockTxns)
	}
	minedTx := block.Transactions()[1]
	txHash := minedTx.Tx.TxHash()
	if txHash != *txId {
		t.Fatalf("txId %v not match vs block.tx[1] %v", txId, txHash)
	}
	t.Logf("txId %v minted in block, hash=%v, order=%v, height=%v", txId, blockHash, block.Order(), block.Height())
}

func AssertTxMinedUseNotifierAPI(t *testing.T, h *Harness, txId *hash.Hash, blockHash *hash.Hash) {
	block, err := h.Notifier.GetBlockV2FullTx(blockHash.String(), true)
	if err != nil {
		t.Fatalf("failed to find block by hash %x : %v", blockHash, err)
	}
	numBlockTxns := len(block.Tx)
	if numBlockTxns < 2 {
		t.Fatalf("the tx has not been mined, the block should at least 2 tx, but got %v", numBlockTxns)
	}
	minedTx := block.Tx[1]
	txHash := minedTx.Txid
	if txHash != txId.String() {
		t.Fatalf("txId %v not match vs block.tx[1] %v", txId, txHash)
	}
	t.Logf("txId %v minted in block, hash=%v, order=%v, height=%v", txId, blockHash, block.Order, block.Height)
}
