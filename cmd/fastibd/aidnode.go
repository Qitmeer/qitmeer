package main

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/services/common"
	"github.com/Qitmeer/qitmeer/services/index"
	"github.com/schollz/progressbar/v3"
	"path"
	"runtime"
	"time"
)

type AidNode struct {
	name  string
	bc    *blockchain.BlockChain
	db    database.DB
	cfg   *Config
	total uint64
}

func (node *AidNode) init(cfg *Config) error {
	runtime.GOMAXPROCS(cfg.CPUNum)
	log.Info(fmt.Sprintf("Start first aid mode. (CPU Num:%d)", cfg.CPUNum))
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
	node.name = path.Base(cfg.DataDir)

	err = db.Update(func(dbTx database.Tx) error {
		// Fetch the stored chain state from the database metadata.
		// When it doesn't exist, it means the database hasn't been
		// initialized for use with chain yet, so break out now to allow
		// that to happen under a writable database transaction.
		meta := dbTx.Metadata()
		serializedData := meta.Get(dbnamespace.ChainStateKeyName)
		if serializedData == nil {
			return nil
		}
		log.Info("Serialized chain state: ", "serializedData", fmt.Sprintf("%x", serializedData))
		state, err := blockchain.DeserializeBestChainState(serializedData)
		if err != nil {
			return err
		}
		log.Info(fmt.Sprintf("blocks:%d", state.GetTotal()))
		node.total = state.GetTotal()
		return nil
	})
	if err != nil {
		return err
	}
	if node.total <= 0 {
		return fmt.Errorf("No blocks in database")
	}
	log.Info(fmt.Sprintf("Load Data:%s", cfg.DataDir))

	return nil
}

func (node *AidNode) exit() error {
	if node.db != nil {
		log.Info(fmt.Sprintf("Gracefully shutting down the database:%s", node.name))
		node.db.Close()
	}
	return nil
}

func (node *AidNode) DB() database.DB {
	return node.db
}

func (node *AidNode) Upgrade() error {
	if node.total <= 0 {
		return fmt.Errorf("No blocks in database")
	}

	endNum := uint(node.total - 1)

	var bar *progressbar.ProgressBar
	if !node.cfg.DisableBar {
		bar = progressbar.Default(int64(endNum), "Export:")
		bar.Add(1)
	} else {
		log.Info("Export...")
	}

	var i uint
	var blockHash *hash.Hash
	blocks := []*types.SerializedBlock{}

	for i = uint(1); i <= endNum; i++ {
		blockHash = nil
		err := node.db.View(func(dbTx database.Tx) error {

			block := &blockdag.Block{}
			block.SetID(i)
			ib := &blockdag.PhantomBlock{Block: block}
			err := blockdag.DBGetDAGBlock(dbTx, ib)
			if err != nil {
				return err
			}
			blockHash = ib.GetHash()

			return nil
		})
		if err != nil {
			return err
		}

		if blockHash == nil {
			return fmt.Errorf(fmt.Sprintf("Can't find block (%d)!", i))
		}

		var blockBytes []byte
		err = node.db.View(func(dbTx database.Tx) error {
			bb, er := dbTx.FetchBlock(blockHash)
			if er != nil {
				return er
			}
			blockBytes = bb
			return nil
		})
		if err != nil {
			return err
		}

		block, err := types.NewBlockFromBytes(blockBytes)
		if err != nil {
			return err
		}
		blocks = append(blocks, block)

		if bar != nil {
			bar.Add(1)
		}
	}

	if node.db != nil {
		log.Info(fmt.Sprintf("Gracefully shutting down the last database:%s", node.name))
		node.db.Close()
	}
	time.Sleep(time.Second * 1)

	common.CleanupBlockDB(&config.Config{DbType: node.cfg.DbType, DataDir: node.cfg.DataDir})

	time.Sleep(time.Second * 2)

	db, err := LoadBlockDB(node.cfg.DbType, node.cfg.DataDir, true)
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
		DAGType:      node.cfg.DAGType,
		IndexManager: indexManager,
	})
	if err != nil {
		log.Error(err.Error())
		return err
	}
	node.bc = bc
	node.name = path.Base(node.cfg.DataDir)

	log.Info(fmt.Sprintf("Load new data:%s", node.cfg.DataDir))

	if bar != nil {
		bar = progressbar.Default(int64(len(blocks)), "Upgrade:")
		bar.Add(1)
	} else {
		log.Info("Upgrade...")
	}
	//
	addNum := int(0)
	lastBH := ""
	defer func() {
		if bar != nil {
			bar.Add(1)
			fmt.Println()
		}
		log.Info(fmt.Sprintf("Finish upgrade: blocks(%d/%d), %s", addNum, endNum, lastBH))
	}()

	for _, block := range blocks {
		err := node.bc.FastAcceptBlock(block, blockchain.BFNone)
		if err != nil {
			fmt.Println()
			log.Info(fmt.Sprintf("The block stopped because of an error:%s", block.Hash().String()))
			return nil
		}
		if bar != nil {
			bar.Add(1)
		}
		addNum++
		lastBH = block.Hash().String()
	}

	return nil
}
