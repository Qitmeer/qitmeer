/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:traderecord.go
 * Date:6/10/20 9:08 AM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package main

import "github.com/Qitmeer/qng-core/common/hash"

type TradeRecord struct {
	blockHash    *hash.Hash
	blockId      uint
	blockOrder   uint
	blockConfirm uint
	blockStatus  byte
	blockBlue    int // 0:not blue;  1：blue  2：Cannot confirm
	blockHeight  uint
	txHash       *hash.Hash
	txFullHash   *hash.Hash
	txUIndex     int
	txValid      bool
	txIsIn       bool
	amount       uint64
	isCoinbase   bool
}
