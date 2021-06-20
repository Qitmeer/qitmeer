/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:srcnode.go
 * Date:5/13/20 6:45 AM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package main

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/params"
	"path"
)

type SrcNode struct {
	name string
	bc   *blockchain.BlockChain
	db   database.DB
	cfg  *Config
}

func (node *SrcNode) init(cfg *Config) error {
	node.cfg = cfg
	// Load the block database.
	srcDataDir := cfg.SrcDataDir
	if cfg.Last {
		srcDataDir = cfg.DataDir
	}
	db, err := LoadBlockDB(cfg.DbType, srcDataDir, false)
	if err != nil {
		log.Error("load block database", "error", err)
		return err
	}
	defer func() {
		// Ensure the database is sync'd and closed on shutdown.

	}()
	node.db = db
	//

	bc, err := blockchain.New(&blockchain.Config{
		DB:          db,
		ChainParams: params.ActiveNetParams.Params,
		TimeSource:  blockchain.NewMedianTime(),
		DAGType:     cfg.DAGType,
	})
	if err != nil {
		log.Error(err.Error())
		return err
	}
	node.bc = bc
	node.name = path.Base(srcDataDir)

	log.Info(fmt.Sprintf("Load Src Data:%s", srcDataDir))
	return nil
}

func (node *SrcNode) exit() {
	if node.db != nil {
		log.Info(fmt.Sprintf("Gracefully shutting down the database:%s", node.name))
		node.db.Close()
	}
}

func (node *SrcNode) BlockChain() *blockchain.BlockChain {
	return node.bc
}

func (node *SrcNode) DB() database.DB {
	return node.db
}
