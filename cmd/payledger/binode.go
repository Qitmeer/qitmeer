/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:binode.go
 * Date:6/6/20 9:28 AM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package main

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/services/index"
	"path"
)

type BINode struct {
	name string
	bc   *blockchain.BlockChain
	db   database.DB
	cfg  *Config
}

func (node *BINode) init(cfg *Config) error {
	node.cfg = cfg

	// Load the block database.
	db, err := LoadBlockDB(cfg.DbType, cfg.DataDir, true)
	if err != nil {
		log.Error("load block database", "error", err)
		return err
	}

	node.db = db
	//
	var indexes []index.Indexer
	txIndex := index.NewTxIndex(db)
	indexes = append(indexes, txIndex)
	// index-manager
	indexManager := index.NewManager(db, indexes, params.ActiveNetParams.Params)

	bc, err := blockchain.New(&blockchain.Config{
		DB:           db,
		ChainParams:  params.ActiveNetParams.Params,
		TimeSource:   blockchain.NewMedianTime(),
		DAGType:      cfg.DAGType,
		IndexManager: indexManager,
	})
	if err != nil {
		log.Error(err.Error())
		return err
	}
	node.bc = bc
	node.name = path.Base(cfg.DataDir)

	log.Info(fmt.Sprintf("Load Data:%s", cfg.DataDir))

	return node.statistics()
}

func (node *BINode) exit() {
	if node.db != nil {
		log.Info(fmt.Sprintf("Gracefully shutting down the database:%s", node.name))
		node.db.Close()
	}
}

func (node *BINode) BlockChain() *blockchain.BlockChain {
	return node.bc
}

func (node *BINode) DB() database.DB {
	return node.db
}

func (node *BINode) statistics() error {
	total := node.bc.BlockDAG().GetBlockTotal()
	validCount := 1
	subsidyCount := 0
	subsidy := uint64(0)
	fmt.Printf("Process...   ")
	for i := uint(1); i < total; i++ {
		ib := node.bc.BlockDAG().GetBlockById(i)
		if ib == nil {
			return fmt.Errorf("No block:%d", i)
		}
		if !knownInvalid(byte(ib.GetStatus())) {
			validCount++

			block, err := node.bc.FetchBlockByHash(ib.GetHash())
			if err != nil {
				return err
			}

			txfullHash := block.Transactions()[0].Tx.TxHashFull()

			if isTxValid(node.db, block.Transactions()[0].Hash(), &txfullHash, ib.GetHash()) {
				if node.bc.BlockDAG().IsBlue(i) {
					subsidyCount++
					subsidy += uint64(block.Transactions()[0].Tx.TxOut[0].Amount.Value)
				}
			}
		}

	}
	mainTip := node.bc.BlockDAG().GetMainChainTip().(*blockdag.PhantomBlock)
	blues := mainTip.GetBlueNum() + 1
	reds := mainTip.GetOrder() + 1 - blues
	unconfirmed := total - (mainTip.GetOrder() + 1)

	fmt.Println()
	fmt.Printf("Total:%d   Valid:%d   BlueNum:%d   RedNum:%d   SubsidyNum:%d Subsidy:%d", total, validCount, blues, reds, subsidyCount, subsidy)
	if unconfirmed > 0 {
		fmt.Printf(" Unconfirmed:%d", unconfirmed)
	}
	fmt.Println()
	fmt.Println("(Note:SubsidyNum does not include genesis.)")

	return nil
}
