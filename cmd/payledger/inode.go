package main

import (
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/database"
)

type INode interface {
	BlockChain() *blockchain.BlockChain
	DB() database.DB
}
