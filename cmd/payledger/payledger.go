package main

import (
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/database"
	_ "github.com/Qitmeer/qitmeer/database/ffldb"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/ledger"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/services/common"
	"github.com/Qitmeer/qitmeer/services/mining"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultSuffixFilename = "payouts.go"
	defaultPayoutDirPath  = "./../../ledger/"
)

func main() {
	// Load configuration and parse command line.  This function also
	// initializes logging and configures it accordingly.
	cfg, _, err := common.LoadConfig()
	if err != nil {
		log.Error(err.Error())
		return
	}

	defer func() {
		if common.LogWrite() != nil {
			common.LogWrite().Close()
		}
	}()

	// Load the block database.
	db, err := common.LoadBlockDB(cfg)
	if err != nil {
		log.Error("load block database", "error", err)
		return
	}
	defer func() {
		// Ensure the database is sync'd and closed on shutdown.
		log.Info("Gracefully shutting down the database...")
		db.Close()
	}()

	// ledger
	buildLedger(cfg, db, params.ActiveNetParams.Params)
}

func buildLedger(cfg *config.Config, db database.DB, params *params.Params) error {

	var err error
	bc, err := blockchain.New(&blockchain.Config{
		DB:           db,
		ChainParams:  params,
		TimeSource:   blockchain.NewMedianTime(),
		DAGType:      cfg.DAGType,
		BlockVersion: mining.BlockVersion(params.Net),
	})
	if err != nil {
		log.Error(err.Error())
		return err
	}

	genesisLedger := map[string]*ledger.TokenPayout{}
	var totalAmount uint64

	log.Info(fmt.Sprintf("Cur main tip:%s", bc.BlockDAG().GetMainChainTip().GetHash().String()))
	err = db.View(func(dbTx database.Tx) error {
		meta := dbTx.Metadata()
		utxoBucket := meta.Bucket(dbnamespace.UtxoSetBucketName)
		cursor := utxoBucket.Cursor()
		for ok := cursor.First(); ok; ok = cursor.Next() {
			serializedUtxo := utxoBucket.Get(cursor.Key())

			// Deserialize the utxo entry and return it.
			entry, err := blockchain.DeserializeUtxoEntry(serializedUtxo)
			if err != nil {
				return err
			}
			if entry.IsSpent() {
				continue
			}
			confir := bc.BlockDAG().GetConfirmations(bc.BlockIndex().GetDAGBlockID(entry.BlockHash()))
			if confir < blockdag.StableConfirmations && !entry.BlockHash().IsEqual(params.GenesisHash) {
				continue
			}
			_, addr, _, err := txscript.ExtractPkScriptAddrs(entry.PkScript(), params)
			if err != nil {
				return err
			}
			var addrStr string
			if len(addr) > 0 {
				for i := 0; i < len(addr); i++ {
					if i > 0 {
						addrStr += "-"
					}
					addrStr += addr[i].String()
				}
			}
			if _, ok := genesisLedger[addrStr]; ok {
				genesisLedger[addrStr].Amount += entry.Amount()
			} else {
				tp := ledger.TokenPayout{Address: addrStr, PkScript: entry.PkScript(), Amount: entry.Amount()}
				genesisLedger[addrStr] = &tp
			}
			totalAmount += entry.Amount()
			if !cfg.BuildLedger {
				log.Trace(fmt.Sprintf("Process Address:%s Amount:%d Block Hash:%s", addrStr, entry.Amount(), entry.BlockHash().String()))
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(genesisLedger) == 0 {
		log.Info("No payouts need to deal with.")
		return nil
	}
	if !cfg.BuildLedger {
		fmt.Println("Show Ledger:")
		for k, v := range genesisLedger {
			fmt.Printf("Address:%s Amount:%d PkScript:%v\n", k, v.Amount, v.PkScript)
		}
	}

	log.Info(fmt.Sprintf("Total Ledger:%d   Amount:%d", len(genesisLedger), totalAmount))

	if !cfg.BuildLedger {
		return nil
	}
	return savePayoutsFile(params, genesisLedger)
}

func savePayoutsFile(params *params.Params, genesisLedger map[string]*ledger.TokenPayout) error {
	if len(genesisLedger) == 0 {
		log.Info("No payouts need to deal with.")
		return nil
	}
	netName := ""
	switch params.Net {
	case protocol.MainNet:
		netName = "main"
	case protocol.TestNet:
		netName = "test"
	case protocol.PrivNet:
		netName = "priv"
	}

	fileName := filepath.Join(defaultPayoutDirPath, netName+defaultSuffixFilename)

	f, err := os.Create(fileName)

	if err != nil {
		log.Error(fmt.Sprintf("Save error:%s  %s", fileName, err))
		return err
	}
	defer func() {
		err = f.Close()
	}()

	funName := fmt.Sprintf("%s%s", strings.ToUpper(string(netName[0])), netName[1:])
	fileContent := fmt.Sprintf("package ledger\nfunc init%s() {\n", funName)

	for k, v := range genesisLedger {
		fileContent += fmt.Sprintf("	addPayout(\"%s\",%d,\"%s\")\n", k, v.Amount, hex.EncodeToString(v.PkScript))
	}
	fileContent += "}"

	f.WriteString(fileContent)

	log.Info(fmt.Sprintf("Finish save %s", fileName))

	return nil
}
