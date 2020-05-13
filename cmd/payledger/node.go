package main

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/services/common"
	"github.com/Qitmeer/qitmeer/services/mining"
	"os"
	"path"
)

type Node struct {
	name     string
	bc       *blockchain.BlockChain
	db       database.DB
	cfg      *Config
	endPoint blockdag.IBlock
}

func (node *Node) init(cfg *Config, srcnode *SrcNode, endPoint blockdag.IBlock) error {
	node.cfg = cfg
	node.endPoint = endPoint
	//
	err := CleanupBlockDB(cfg)
	if err != nil {
		return err
	}
	// Load the block database.
	db, err := LoadBlockDB(cfg.DbType, cfg.DataDir, true)
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
		DB:           db,
		ChainParams:  params.ActiveNetParams.Params,
		TimeSource:   blockchain.NewMedianTime(),
		DAGType:      cfg.DAGType,
		BlockVersion: mining.BlockVersion(params.ActiveNetParams.Params.Net),
	})
	if err != nil {
		log.Error(err.Error())
		return err
	}
	node.bc = bc
	node.name = path.Base(cfg.DataDir)

	log.Info(fmt.Sprintf("Load Data:%s", cfg.DataDir))

	return node.processBlockDAG(srcnode)
}

func (node *Node) exit() {
	if node.db != nil {
		log.Info(fmt.Sprintf("Gracefully shutting down the database:%s", node.name))
		node.db.Close()
	}
}

func (node *Node) BlockChain() *blockchain.BlockChain {
	return node.bc
}

func (node *Node) DB() database.DB {
	return node.db
}

func (node *Node) processBlockDAG(srcnode *SrcNode) error {
	genesisHash := node.bc.BlockDAG().GetGenesisHash()
	srcgenesisHash := srcnode.BlockChain().BlockDAG().GetGenesisHash()
	if !genesisHash.IsEqual(srcgenesisHash) {
		return fmt.Errorf("Different genesis!")
	}
	common.Glogger().Verbosity(log.LvlCrit)
	srcTotal := srcnode.bc.BlockDAG().GetBlockTotal()
	i := uint(0)
	bar := ProgressBar{}
	bar.init()
	bar.reset(int(node.endPoint.GetID() + 1))
	bar.refresh()
	for ; i < srcTotal; i++ {
		blockHash := srcnode.bc.BlockDAG().GetBlockHash(i)
		if blockHash == nil {
			return fmt.Errorf(fmt.Sprintf("Can't find block id (%d)!", i))
		}
		if blockHash.IsEqual(srcgenesisHash) {
			continue
		}
		block, err := srcnode.bc.FetchBlockByHash(blockHash)
		if err != nil {
			return err
		}
		//fmt.Printf("%d %s\n", i, blockHash.String())
		err = node.bc.FastAcceptBlock(block)
		if err != nil {
			return err
		}
		if blockHash.IsEqual(node.endPoint.GetHash()) {
			break
		}
		bar.add()
	}
	bar.setMax()
	fmt.Println()

	common.Glogger().Verbosity(log.LvlInfo)
	log.Info(fmt.Sprintf("End process block DAG:(%d/%d)", i, srcTotal))
	return nil
}

// removeBlockDB removes the existing database
func removeBlockDB(dbPath string) error {
	// Remove the old database if it already exists.
	fi, err := os.Stat(dbPath)
	if err == nil {
		log.Info(fmt.Sprintf("Removing block database from '%s'", dbPath))
		if fi.IsDir() {
			err := os.RemoveAll(dbPath)
			if err != nil {
				return err
			}
		} else {
			err := os.Remove(dbPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func CleanupBlockDB(cfg *Config) error {
	dbPath := blockDbPath(cfg.DbType, cfg.DataDir)
	err := removeBlockDB(dbPath)
	if err != nil {
		return err
	}
	log.Info("Finished cleanup")
	return nil
}
