package handler

import (
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/noxproject/nox/middleware/config"
	"log"
)

func GetBlockCountByHttp() int64 {

	connCfg := &rpcclient.ConnConfig{
		Host:         config.NoxdHost,
		User:         config.RpcUser,
		Pass:         config.RpcPassword,
		HTTPPostMode: true,
		DisableTLS:   true,
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Println(err)
	}
	defer client.Shutdown()

	blockCount, err := client.GetBlockCount()
	if err != nil {
		log.Println("GetBlockCount error : " + err.Error())
	}

	log.Printf("Block count: %d", blockCount)

	return blockCount
}
