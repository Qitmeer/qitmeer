// Copyright 2017-2018 The nox developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/noxproject/nox/core/json"
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/params"
	"github.com/noxproject/nox/services/common/marshal"
)

func txDecode(network string, rawTxStr string) {
	var param *params.Params
	switch network {
	case "mainnet":
		param = &params.MainNetParams
	case "testnet":
		param = &params.TestNetParams
	case "privnet":
		param = &params.PrivNetParams
	}
	if len(rawTxStr)%2 != 0 {
		errExit(fmt.Errorf("invaild raw transaction : %s",rawTxStr))
	}
	serializedTx, err := hex.DecodeString(rawTxStr)
	if err != nil {
		errExit(err)
	}
	var tx types.Transaction
	err = tx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		errExit(err)
	}

	jsonTx := &json.OrderedResult{
		{"txid", tx.TxHash().String()},
		{"txhash", tx.TxHashFull().String()},
		{"version",  int32(tx.Version)},
		{"locktime", tx.LockTime},
		{"vin",      marshal.MarshJsonVin(&tx)},
		{"vout",     marshal.MarshJsonVout(&tx, nil,param)},
	}
	marshaledTx, err := jsonTx.MarshalJSON()
	if err != nil {
		errExit(err)
	}

	fmt.Printf("%s",marshaledTx)
}




