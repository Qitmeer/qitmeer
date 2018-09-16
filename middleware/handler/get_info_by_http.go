package handler

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/noxproject/nox/middleware/config"
	"log"
)

func GetBlockInfoByHttp(blockHashStr string) *btcjson.GetBlockVerboseResult {

	connCfg := &rpcclient.ConnConfig{
		Host:         config.NoxdHost,
		User:         config.RpcUser,
		Pass:         config.RpcPassword,
		HTTPPostMode: true,
		DisableTLS:   true,
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {

		log.Fatal(err)
	}
	defer client.Shutdown()

	blockHash, err := chainhash.NewHashFromStr(blockHashStr)
	if err != nil {
		log.Fatal(err)
	}
	blockJsonStrut, err := client.GetBlockVerbose(blockHash)
	if err != nil {
		log.Println("GetBlockInfo error : " + err.Error())
	}

	//TODO:会返回nil
	return blockJsonStrut
}
