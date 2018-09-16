package handler

import (
	"github.com/noxproject/nox/middleware/config"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/btcsuite/btcutil"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
)

//通过WebSocket的方式
func GetBlockInfo(blockHashStr string) *btcjson.GetBlockVerboseResult {

	btcdHomeDir := btcutil.AppDataDir("noxd", false)
	certs, err := ioutil.ReadFile(filepath.Join(btcdHomeDir, "rpc.cert"))
	if err != nil {
		log.Fatal(err)
	}

	connCfg := &rpcclient.ConnConfig{
		Host:         config.NoxdHost,
		User:         config.RpcUser,
		Pass:         config.RpcPassword,
		Endpoint:     config.EndpointWS,
		Certificates: certs,
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

	return blockJsonStrut
}
