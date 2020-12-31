/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package client

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
	"time"
)

type NotificationHandlers struct {
	OnClientConnected     func()
	OnBlockConnected      func(hash *hash.Hash, order int64, t time.Time, txs []*types.Transaction)
	OnBlockDisconnected   func(hash *hash.Hash, order int64, t time.Time, txs []*types.Transaction)
	OnUnknownNotification func(method string, params []json.RawMessage)
}

func parseChainNtfnParams(params []json.RawMessage) (*hash.Hash, int64, time.Time, []*types.Transaction, error) {
	if len(params) != 4 {
		return nil, 0, time.Time{}, nil, wrongNumParams(len(params))
	}

	// Unmarshal first parameter as a string.
	var blockHashStr string
	err := json.Unmarshal(params[0], &blockHashStr)
	if err != nil {
		return nil, 0, time.Time{}, nil, err
	}

	// Unmarshal second parameter as an integer.
	var blockOrder int64
	err = json.Unmarshal(params[1], &blockOrder)
	if err != nil {
		return nil, 0, time.Time{}, nil, err
	}

	// Unmarshal third parameter as unix time.
	var blockTimeUnix int64
	err = json.Unmarshal(params[2], &blockTimeUnix)
	if err != nil {
		return nil, 0, time.Time{}, nil, err
	}

	var txHexs []string
	err = json.Unmarshal(params[3], &txHexs)
	if err != nil {
		return nil, 0, time.Time{}, nil, err
	}
	txs := []*types.Transaction{}
	for _, txHex := range txHexs {
		serializedTx, err := hex.DecodeString(txHex)
		if err != nil {
			return nil, 0, time.Time{}, nil, err
		}
		var tx types.Transaction
		err = tx.Deserialize(bytes.NewReader(serializedTx))
		if err != nil {
			return nil, 0, time.Time{}, nil, err
		}
		txs = append(txs, &tx)
	}

	// Create hash from block hash string.
	blockHash, err := hash.NewHashFromStr(blockHashStr)
	if err != nil {
		return nil, 0, time.Time{}, nil, err
	}

	// Create time.Time from unix time.
	blockTime := time.Unix(blockTimeUnix, 0)

	return blockHash, blockOrder, blockTime, txs, nil
}
