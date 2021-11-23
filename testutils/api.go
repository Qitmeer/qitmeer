// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package testutils

import (
	"encoding/hex"
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/core/json"
	"github.com/Qitmeer/qng-core/core/types"
	"github.com/Qitmeer/qng-core/core/types/pow"
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

func (c *Client) SendRawTx(txHex string, allowHighFees bool) (*hash.Hash, error) {

	//fmt.Printf("send rawtx=%s\n", txHex)
	var result *hash.Hash
	if err := c.Call(&result, "sendRawTransaction", txHex, allowHighFees); err != nil {
		return nil, err
	}
	return result, nil
}

// TODO, the SerializedBlock not work for order and height
func (c *Client) GetSerializedBlock(h *hash.Hash) (*types.SerializedBlock, error) {
	var blockHex string
	if err := c.Call(&blockHex, "getBlock", h.String(), false); err != nil {
		return nil, err
	}
	// Decode the serialized block hex to raw bytes.
	serializedBlock, err := hex.DecodeString(blockHex)
	if err != nil {
		return nil, err
	}
	// Deserialize the block and return it.
	block, err := types.NewBlockFromBytes(serializedBlock)
	if err != nil {
		return nil, err
	}
	return block, nil
}

// TODO, the api is not easy to use when doing the internal-test
func (c *Client) GetBlock(h *hash.Hash) (*json.BlockVerboseResult, error) {
	var result json.BlockVerboseResult
	if err := c.Call(&result, "getBlockV2", h.String(), true, true, true); err != nil {
		return nil, err
	}
	return &result, nil
}
