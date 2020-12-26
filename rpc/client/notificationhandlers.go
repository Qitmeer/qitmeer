/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package client

import (
	"encoding/json"
	"github.com/Qitmeer/qitmeer/common/hash"
	"time"
)

type NotificationHandlers struct {
	OnClientConnected     func()
	OnBlockConnected      func(hash *hash.Hash, height int32, t time.Time)
	OnBlockDisconnected   func(hash *hash.Hash, height int32, t time.Time)
	OnUnknownNotification func(method string, params []json.RawMessage)
}

const (
	BlockConnectedNtfnMethod    = "blockconnected"
	BlockDisconnectedNtfnMethod = "blockdisconnected"
)

func parseChainNtfnParams(params []json.RawMessage) (*hash.Hash, int32, time.Time, error) {

	if len(params) != 3 {
		return nil, 0, time.Time{}, wrongNumParams(len(params))
	}

	// Unmarshal first parameter as a string.
	var blockHashStr string
	err := json.Unmarshal(params[0], &blockHashStr)
	if err != nil {
		return nil, 0, time.Time{}, err
	}

	// Unmarshal second parameter as an integer.
	var blockHeight int32
	err = json.Unmarshal(params[1], &blockHeight)
	if err != nil {
		return nil, 0, time.Time{}, err
	}

	// Unmarshal third parameter as unix time.
	var blockTimeUnix int64
	err = json.Unmarshal(params[2], &blockTimeUnix)
	if err != nil {
		return nil, 0, time.Time{}, err
	}

	// Create hash from block hash string.
	blockHash, err := hash.NewHashFromStr(blockHashStr)
	if err != nil {
		return nil, 0, time.Time{}, err
	}

	// Create time.Time from unix time.
	blockTime := time.Unix(blockTimeUnix, 0)

	return blockHash, blockHeight, blockTime, nil
}
