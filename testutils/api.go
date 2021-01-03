// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package testutils

import (
	"encoding/hex"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"strconv"
)

func (c *Client) NodeInfo() (json.InfoNodeResult, error) {
	var result json.InfoNodeResult
	if err := c.Call(&result, "getNodeInfo"); err != nil {
		return result, err
	}
	return result, nil
}

func (c *Client) BlockCount() (uint64, error) {
	var result uint64
	if err := c.Call(&result, "getBlockCount"); err != nil {
		return result, err
	}
	return result, nil
}

func (c *Client) BlockTotal() (uint64, error) {
	var result uint64
	if err := c.Call(&result, "getBlockTotal"); err != nil {
		return result, err
	}
	return result, nil
}

func (c *Client) MainHeight() (uint64, error) {
	var result string
	if err := c.Call(&result, "getMainChainHeight"); err != nil {
		return 0, err
	}
	if height, err := strconv.Atoi(result); err != nil {
		return 0, err
	} else {
		return uint64(height), nil
	}
}

func (c *Client) Generate(num uint64) ([]*hash.Hash, error) {
	var result []*hash.Hash
	if err := c.Call(&result, "miner_generate", num, pow.PowType(0)); err != nil {
		return result, err
	}
	return result, nil
}

func (c *Client) SendRawTx(tx *types.Transaction, allowHighFees bool) (*hash.Hash, error) {
	txByte, err := tx.Serialize()
	if err != nil {
		return nil, err
	}
	txHex := hex.EncodeToString(txByte[:])

	var result *hash.Hash
	if err := c.Call(&result, "sendRawTransaction", txHex, allowHighFees); err != nil {
		return nil, err
	}
	return result, nil
}
