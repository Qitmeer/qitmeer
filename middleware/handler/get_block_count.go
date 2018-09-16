package handler

import (
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
	"github.com/noxproject/nox/middleware/config"
	"io/ioutil"
	"log"
	"path/filepath"
)

func GetBlockCount() int64 {
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

	blockCount, err := client.GetBlockCount()
	if err != nil {
		log.Println("GetBlockCount error : " + err.Error())
	}

	log.Printf("Block count: %d", blockCount)

	return blockCount
}
