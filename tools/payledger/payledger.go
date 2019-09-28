package payledger

import (
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer-lib/common/util"
	"github.com/Qitmeer/qitmeer-lib/core/protocol"
	"github.com/Qitmeer/qitmeer-lib/engine/txscript"
	"github.com/Qitmeer/qitmeer-lib/params"
	"github.com/Qitmeer/qitmeer-lib/config"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer/ledger"
	"github.com/Qitmeer/qitmeer/services/mining"
	"os"
	"strings"
)

func BuildLedger(cfg *config.Config,db database.DB,params *params.Params) error {

	var err error
	bc, err := blockchain.New(&blockchain.Config{
		DB:            db,
		ChainParams:   params,
		TimeSource:    blockchain.NewMedianTime(),
		DAGType:       cfg.DAGType,
		BlockVersion:  mining.BlockVersion(params.Net),
	})
	if err != nil {
		return err
	}
	
	genesisLedger:=map[string]*ledger.TokenPayout{}
	log.Info("Show Ledger:")
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
			confir:=bc.BlockDAG().GetConfirmations(entry.BlockHash())
			if confir < blockdag.StableConfirmations {
				continue
			}
			_, addr,_, err := txscript.ExtractPkScriptAddrs(entry.PkScript(), params)
			if err != nil {
				return err
			}
			var addrStr string
			if len(addr)>0 {
				for i:=0;i<len(addr) ;i++  {
					if i>0 {
						addrStr+="-"
					}
					addrStr+=addr[i].String()
				}
			}
			if _,ok:=genesisLedger[addrStr];ok {
				genesisLedger[addrStr].Amount+=entry.Amount()
			}else{
				tp:=ledger.TokenPayout{Address:addrStr,PkScript:entry.PkScript(),Amount:entry.Amount()}
				genesisLedger[addrStr]=&tp
			}
			log.Info(fmt.Sprintf("Process Address:%s Amount:%d Block Hash:%s",addrStr,entry.Amount(),entry.BlockHash().String()))
		}
		return nil
	})
	if err != nil {
		return err
	}

	var totalAmount uint64
	for k,v:=range genesisLedger {
		totalAmount+=v.Amount
		log.Info(fmt.Sprintf("Address:%s Amount:%d PkScript:%v",k,v.Amount,v.PkScript))
	}
	log.Info(fmt.Sprintf("Total Ledger:%d   Amount:%d",len(genesisLedger),totalAmount))

	if !cfg.BuildLedger {
		return nil
	}
	return savePayoutsFile(cfg,params,genesisLedger)
}

func savePayoutsFile(cfg *config.Config,params *params.Params,genesisLedger map[string]*ledger.TokenPayout) error {
	if len(genesisLedger)==0 {
		log.Info("No payouts")
		return nil
	}
	netName:=""
	switch params.Net {
	case protocol.MainNet:
		netName="main"
	case protocol.TestNet:
		netName="test"
	case protocol.PrivNet:
		netName="priv"
	}

	dir:="./ledger/"
	if !util.FileExists(dir) {
		dir="./"
	}

	fileName:=dir+netName+"payouts.go"
	f,err:= os.Create(fileName)

	if err != nil {
		log.Error(fmt.Sprintf("Save error:%s  %s",fileName,err))
		return err
	}
	defer func() {
		err=f.Close()
	}()

	funName:=fmt.Sprintf("%s%s",strings.ToUpper(string(netName[0])),netName[1:])
	fileContent:=fmt.Sprintf("package ledger\nfunc init%s() {\n",funName)

	for k,v:=range genesisLedger {
		fileContent+=fmt.Sprintf("	addPayout(\"%s\",%d,\"%s\")\n",k,v.Amount,hex.EncodeToString(v.PkScript))
	}
	fileContent+="}"

	f.WriteString(fileContent)

	log.Info(fmt.Sprintf("Finish save %s",fileName))

	return nil
}