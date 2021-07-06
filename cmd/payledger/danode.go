/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:danode.go
 * Date:6/6/20 9:28 AM
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
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/services/index"
	"path"
)

type DebugAddressNode struct {
	name     string
	bc       *blockchain.BlockChain
	db       database.DB
	cfg      *Config
	endPoint blockdag.IBlock

	info []DebugAddrInfo
}

func (node *DebugAddressNode) init(cfg *Config) error {
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

	// process address
	blueMap := map[uint]bool{}
	if !cfg.DebugAddrUTXO {
		err = node.processAddress(&blueMap)
		if err != nil {
			return err
		}
	}

	return node.checkUTXO(&blueMap)
}

func (node *DebugAddressNode) exit() {
	if node.db != nil {
		log.Info(fmt.Sprintf("Gracefully shutting down the database:%s", node.name))
		node.db.Close()
	}
}

func (node *DebugAddressNode) BlockChain() *blockchain.BlockChain {
	return node.bc
}

func (node *DebugAddressNode) DB() database.DB {
	return node.db
}

// Load from database
func (node *DebugAddressNode) LoadInfo() error {
	err := node.db.View(func(dbTx database.Tx) error {
		meta := dbTx.Metadata()
		serializedData := meta.Get(DebugAddrInfoBucketName)
		if serializedData == nil {
			log.Info("No Debug Address Data")
			return nil
		}

		info, err := decodeDebugAddrInfo(serializedData)
		if err != nil {
			return err
		}
		node.info = info
		return nil
	})
	if node.info != nil {
		log.Info(fmt.Sprintf("Load Address info: total=%d", len(node.info)))
	}
	return err
}

func (node *DebugAddressNode) processAddress(blueM *map[uint]bool) error {
	db := node.db
	par := params.ActiveNetParams.Params
	checkAddress := node.cfg.DebugAddress
	tradeRecord := []*TradeRecord{}
	tradeRecordMap := map[types.TxOutPoint]*TradeRecord{}
	blueMap := *blueM
	mainTip := node.bc.BlockDAG().GetMainChainTip()
	fmt.Printf("Start analysis:%s  mainTip:%s mainOrder:%d total:%d \n", checkAddress, mainTip.GetHash(), mainTip.GetOrder(), node.bc.BlockDAG().GetBlockTotal())
	for i := uint(0); i < node.bc.BlockDAG().GetBlockTotal(); i++ {
		ib := node.bc.BlockDAG().GetBlockById(i)
		if ib == nil {
			return fmt.Errorf("Error：%d", i)
		}
		block, err := node.bc.FetchBlockByHash(ib.GetHash())
		if err != nil {
			return fmt.Errorf("Can't find：%s", err)
		}
		confims := node.bc.BlockDAG().GetConfirmations(ib.GetID())

		for _, tx := range block.Transactions() {
			txHash := tx.Hash()
			txFullHash := tx.Tx.TxHashFull()

			txValid := isTxValid(db, txHash, &txFullHash, ib.GetHash())
			if node.cfg.DebugAddrValid {
				if !txValid {
					continue
				}
			}
			if !tx.Tx.IsCoinBase() {
				for txInIndex, txIn := range tx.Tx.TxIn {
					pretr, ok := tradeRecordMap[txIn.PreviousOut]
					if ok {
						tr := &TradeRecord{}
						tr.blockHash = ib.GetHash()
						tr.blockId = ib.GetID()
						tr.blockOrder = ib.GetOrder()
						tr.blockConfirm = confims
						tr.blockStatus = byte(ib.GetStatus())
						tr.blockBlue = 2
						tr.blockHeight = ib.GetHeight()
						tr.txHash = txHash
						tr.txFullHash = &txFullHash
						tr.txUIndex = txInIndex
						tr.txIsIn = true
						tr.txValid = txValid
						tr.isCoinbase = false
						tr.amount = pretr.amount

						if !knownInvalid(tr.blockStatus) && tr.txValid {

							isblue, ok := blueMap[ib.GetID()]
							if !ok {
								isblue = node.bc.BlockDAG().IsBlue(ib.GetID())
								blueMap[ib.GetID()] = isblue
							}
							if isblue {
								tr.blockBlue = 1
							} else {
								tr.blockBlue = 0
							}
						}
						tradeRecord = append(tradeRecord, tr)
					}
				}

			}
			for txOutIndex, txOut := range tx.Tx.TxOut {
				_, addr, _, err := txscript.ExtractPkScriptAddrs(txOut.GetPkScript(), par)
				if err != nil {
					return err
				}
				if len(addr) != 1 {
					fmt.Printf("Ignore multiple addresses：%d\n", len(addr))
					continue
				}
				addrStr := addr[0].String()
				if addrStr != checkAddress {
					continue
				}

				tr := &TradeRecord{}
				tr.blockHash = ib.GetHash()
				tr.blockId = ib.GetID()
				tr.blockOrder = ib.GetOrder()
				tr.blockConfirm = confims
				tr.blockStatus = byte(ib.GetStatus())
				tr.blockBlue = 2
				tr.blockHeight = ib.GetHeight()
				tr.txHash = txHash
				tr.txFullHash = &txFullHash
				tr.txUIndex = txOutIndex
				tr.txIsIn = false
				tr.txValid = txValid
				tr.amount = uint64(txOut.Amount.Value)
				tr.isCoinbase = tx.Tx.IsCoinBase()

				if !knownInvalid(tr.blockStatus) && tr.txValid {

					isblue, ok := blueMap[ib.GetID()]
					if !ok {
						isblue = node.bc.BlockDAG().IsBlue(ib.GetID())
						blueMap[ib.GetID()] = isblue
					}
					if isblue {
						tr.blockBlue = 1
					} else {
						tr.blockBlue = 0
					}
				}

				tradeRecord = append(tradeRecord, tr)
				txOutPoint := types.TxOutPoint{*txHash, uint32(txOutIndex)}
				tradeRecordMap[txOutPoint] = tr
			}
		}
	}
	acc := int64(0)
	for i, tr := range tradeRecord {
		isValid := true
		if tr.isCoinbase && !tr.txIsIn && tr.blockBlue == 0 {
			isValid = false
		}
		if isValid {
			if tr.txValid {
				if tr.txIsIn {
					acc -= int64(tr.amount)
				} else {
					acc += int64(tr.amount)
				}
			}
		}

		if node.cfg.DebugAddrValid {
			if !isValid || !tr.txValid {
				continue
			}
		}

		fmt.Printf("%d Block Hash:%s Id:%d Order:%d Confirm:%d Valid:%v Blue:%s Height:%d ; Tx Hash:%s FullHash:%s UIndex:%d IsIn:%v Valid:%v Amount:%d Coinbase:%v  Acc:%d\n",
			i, tr.blockHash, tr.blockId, tr.blockOrder, tr.blockConfirm, !knownInvalid(tr.blockStatus), blueState(tr.blockBlue), tr.blockHeight, tr.txHash, tr.txFullHash, tr.txUIndex, tr.txIsIn, tr.txValid,
			tr.amount, tr.isCoinbase, acc)

	}

	fmt.Printf("Result：%s   Number of ledger records:%d    Total balance:%d\n\n", checkAddress, len(tradeRecord), acc)

	return nil
}

func isTxValid(db database.DB, txHash *hash.Hash, txFullHash *hash.Hash, blockHash *hash.Hash) bool {
	var preTx *types.Transaction
	var preBlockH *hash.Hash
	err := db.View(func(dbTx database.Tx) error {
		dtx, blockH, erro := index.DBFetchTxAndBlock(dbTx, txHash)
		if erro != nil {
			return erro
		}
		preTx = dtx
		preBlockH = blockH
		return nil
	})

	if err != nil {
		return false
	}
	ptxFullHash := preTx.TxHashFull()

	if !preBlockH.IsEqual(blockHash) || !txFullHash.IsEqual(&ptxFullHash) {
		return false
	}
	return true
}

func (node *DebugAddressNode) checkUTXO(blueM *map[uint]bool) error {

	db := node.db
	par := params.ActiveNetParams.Params
	checkAddress := node.cfg.DebugAddress

	blueMap := *blueM

	fmt.Printf("Checking UTXO:%s\n", checkAddress)

	var totalAmount uint64
	var count int
	serializedUtxos := [][]byte{}

	err := db.View(func(dbTx database.Tx) error {
		meta := dbTx.Metadata()
		utxoBucket := meta.Bucket(dbnamespace.UtxoSetBucketName)
		cursor := utxoBucket.Cursor()
		for ok := cursor.First(); ok; ok = cursor.Next() {
			serializedUtxo := utxoBucket.Get(cursor.Key())
			serializedUtxos = append(serializedUtxos, serializedUtxo)
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, serializedUtxo := range serializedUtxos {
		// Deserialize the utxo entry and return it.
		entry, err := blockchain.DeserializeUtxoEntry(serializedUtxo)
		if err != nil {
			return err
		}
		if entry.IsSpent() {
			continue
		}
		ib := node.bc.GetBlock(entry.BlockHash())
		if ib.GetOrder() == blockdag.MaxBlockOrder {
			continue
		}
		_, addr, _, err := txscript.ExtractPkScriptAddrs(entry.PkScript(), par)
		if err != nil {
			return err
		}
		addrStr := addr[0].String()
		if addrStr != checkAddress {
			continue
		}
		isValid := true
		blockBlue := 2
		if entry.IsCoinBase() {
			isblue, ok := blueMap[ib.GetID()]
			if !ok {
				isblue = node.bc.BlockDAG().IsBlue(ib.GetID())
			}
			if !isblue {
				isValid = false
				blockBlue = 0
			} else {
				blockBlue = 1
			}

		}

		if isValid {
			totalAmount += uint64(entry.Amount().Value)
		}

		count++

		if node.cfg.DebugAddrValid {
			if !isValid {
				continue
			}
		}

		fmt.Printf("BlockHash:%s Amount:%d Valid:%v Blue:%s\n", ib.GetHash(), entry.Amount(), isValid, blueState(blockBlue))

	}

	fmt.Printf("UTXO Result： Number of ledger records：%d   Total balance:%d\n", count, totalAmount)
	return nil
}

func knownInvalid(status byte) bool {
	var statusInvalid byte
	statusInvalid = 1 << 2
	return status&statusInvalid != 0
}

func blueState(blockBlue int) string {
	if blockBlue == 0 {
		return "No"
	} else if blockBlue == 1 {
		return "Yes"
	}
	return "?"
}
