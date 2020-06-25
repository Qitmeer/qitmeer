package main

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/services/common"
	"github.com/Qitmeer/qitmeer/services/index"
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
		BlockVersion: mining.BlockVersion(params.ActiveNetParams.Params.Net),
		IndexManager: indexManager,
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
	srcTotal := srcnode.bc.BlockDAG().GetBlockTotal()
	if node.endPoint.GetHash().IsEqual(genesisHash) {
		return nil
	}

	common.Glogger().Verbosity(log.LvlCrit)
	var bar *ProgressBar
	i := uint(1)
	if !node.cfg.DisableBar {

		bar = &ProgressBar{}
		bar.init("Process:")
		bar.reset(int(node.endPoint.GetID() + 1))
		bar.add()
	} else {
		log.Info("Process...")
	}

	defer func() {
		common.Glogger().Verbosity(log.LvlInfo)
		if bar != nil {
			fmt.Println()
		}
		log.Info(fmt.Sprintf("End process block DAG:(%d/%d)", i, srcTotal))
	}()
	for ; i < srcTotal; i++ {
		blockHash := srcnode.bc.BlockDAG().GetBlockHash(i)
		if blockHash == nil {
			return fmt.Errorf(fmt.Sprintf("Can't find block id (%d)!", i))
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
		if bar != nil {
			bar.add()
		}
		if blockHash.IsEqual(node.endPoint.GetHash()) {
			break
		}
	}
	if bar != nil {
		bar.setMax()
	}
	err := node.dataVerification(srcnode)
	if err != nil {
		return err
	}

	return nil
}

func (node *Node) dataVerification(srcnode *SrcNode) error {
	fmt.Println()
	total := node.bc.BlockDAG().GetBlockTotal()

	var bar *ProgressBar
	i := uint(1)
	if !node.cfg.DisableBar {
		bar = &ProgressBar{}
		bar.init("Validate:")
		bar.reset(int(node.endPoint.GetID() + 1))
		bar.add()
	} else {
		log.Info("Validate...")
	}

	for ; i < total; i++ {
		srcIB := srcnode.bc.BlockDAG().GetBlockById(i)
		if srcIB == nil {
			return fmt.Errorf(fmt.Sprintf("Can't find block id (%d) from src node!", i))
		}
		ib := node.bc.BlockDAG().GetBlockById(i)
		if ib == nil {
			return fmt.Errorf(fmt.Sprintf("Can't find block id (%d) from node!", i))
		}
		if srcIB.GetStatus() != ib.GetStatus() ||
			srcIB.GetHeight() != ib.GetHeight() ||
			!srcIB.GetHash().IsEqual(ib.GetHash()) {
			return fmt.Errorf(fmt.Sprintf("Validate fail (%s)!", srcIB.GetHash().String()))
		}
		if bar != nil {
			bar.add()
		}
	}
	if bar != nil {
		bar.setMax()
		fmt.Println()
	}
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
