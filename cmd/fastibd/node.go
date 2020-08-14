/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:node.go
 * Date:6/20/20 7:37 AM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package main

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/services/index"
	"github.com/Qitmeer/qitmeer/services/mining"
	"os"
	"path"
)

type Node struct {
	name string
	bc   *blockchain.BlockChain
	db   database.DB
	cfg  *Config
}

func (node *Node) init(cfg *Config) error {
	err := cfg.load()
	if err != nil {
		return err
	}
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

	return nil
}

func (node *Node) exit() error {
	if node.db != nil {
		log.Info(fmt.Sprintf("Gracefully shutting down the database:%s", node.name))
		node.db.Close()
	}
	return nil
}

func (node *Node) BlockChain() *blockchain.BlockChain {
	return node.bc
}

func (node *Node) DB() database.DB {
	return node.db
}

func (node *Node) Export() error {
	mainTip := node.bc.BlockDAG().GetMainChainTip()
	if mainTip.GetOrder() <= 0 {
		return fmt.Errorf("No blocks in database")
	}
	outFilePath, err := GetIBDFilePath(node.cfg.OutputPath)
	if err != nil {
		return err
	}

	outFile, err := os.OpenFile(outFilePath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer func() {
		outFile.Close()
	}()

	var endPoint blockdag.IBlock
	endNum := uint(0)
	if node.cfg.ByID {
		endNum = mainTip.GetID()
	} else {
		endNum = mainTip.GetOrder()
	}

	if len(node.cfg.EndPoint) > 0 {
		ephash, err := hash.NewHashFromStr(node.cfg.EndPoint)
		if err != nil {
			return err
		}
		endPoint = node.bc.BlockDAG().GetBlock(ephash)
		if endPoint != nil {
			if node.cfg.ByID {
				if endNum > endPoint.GetID() {
					endNum = endPoint.GetID()
				}
			} else {
				if endNum > endPoint.GetOrder() {
					endNum = endPoint.GetOrder()
				}
			}

			log.Info(fmt.Sprintf("End point:%s order:%d id:%d", ephash.String(), endPoint.GetOrder(), endPoint.GetID()))
		} else {
			return fmt.Errorf("End point is error")
		}

	}
	var bar *ProgressBar
	if !node.cfg.DisableBar {

		bar = &ProgressBar{}
		bar.init("Export:")
		bar.reset(int(endNum))
		bar.add()
	} else {
		log.Info("Export...")
	}

	var maxNum [4]byte
	dbnamespace.ByteOrder.PutUint32(maxNum[:], uint32(endNum))
	_, err = outFile.Write(maxNum[:])
	if err != nil {
		return err
	}
	var i uint
	var blockHash *hash.Hash
	for i = uint(1); i <= endNum; i++ {
		if node.cfg.ByID {
			ib := node.bc.BlockDAG().GetBlockById(i)
			if ib != nil {
				blockHash = ib.GetHash()
			} else {
				blockHash = nil
			}
		} else {
			blockHash = node.bc.BlockDAG().GetBlockByOrder(i)
		}

		if blockHash == nil {
			return fmt.Errorf(fmt.Sprintf("Can't find block (%d)!", i))
		}

		block, err := node.bc.FetchBlockByHash(blockHash)
		if err != nil {
			return err
		}
		bytes, err := block.Bytes()
		if err != nil {
			return err
		}
		ibdb := &IBDBlock{length: uint32(len(bytes)), bytes: bytes}
		err = ibdb.Encode(outFile)
		if err != nil {
			return err
		}
		if bar != nil {
			bar.add()
		}

		/*if endPoint != nil {
			if endPoint.GetHash().IsEqual(blockHash) {
				break
			}
		}*/
	}
	if bar != nil {
		bar.setMax()
		fmt.Println()
	}
	log.Info(fmt.Sprintf("Finish export: blocks(%d)    ------>File:%s", endNum, outFilePath))
	return nil
}

func (node *Node) Import() error {
	mainTip := node.bc.BlockDAG().GetMainChainTip()
	if mainTip.GetOrder() > 0 {
		return fmt.Errorf("Your database is not empty, please empty the database.")
	}
	inputFilePath, err := GetIBDFilePath(node.cfg.InputPath)
	if err != nil {
		return err
	}
	blocksBytes, err := ReadFile(inputFilePath)
	if err != nil {
		return err
	}
	offset := 0
	maxOrder := dbnamespace.ByteOrder.Uint32(blocksBytes[offset : offset+4])
	offset += 4

	var bar *ProgressBar
	if !node.cfg.DisableBar {

		bar = &ProgressBar{}
		bar.init("Import:")
		bar.reset(int(maxOrder))
		bar.add()
	} else {
		log.Info("Import...")
	}
	for i := uint32(1); i <= maxOrder; i++ {
		ibdb := &IBDBlock{}
		err := ibdb.Decode(blocksBytes[offset:])
		if err != nil {
			return err
		}
		offset += 4 + int(ibdb.length)

		err = node.bc.FastAcceptBlock(ibdb.blk)
		if err != nil {
			return err
		}
		if bar != nil {
			bar.add()
		}
	}

	if bar != nil {
		bar.setMax()
		fmt.Println()
	}
	mainTip = node.bc.BlockDAG().GetMainChainTip()
	log.Info(fmt.Sprintf("Finish import: blocks(%d)    ------>File:%s", mainTip.GetOrder(), inputFilePath))
	log.Info(fmt.Sprintf("New Info:%s  mainOrder=%d tips=%d", mainTip.GetHash().String(), mainTip.GetOrder(), node.bc.BlockDAG().GetTips().Size()))
	return nil
}
