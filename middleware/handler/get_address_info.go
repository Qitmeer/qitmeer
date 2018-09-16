package handler

import (
	"github.com/noxproject/nox/middleware/config"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
)

/**
 * Author: 傅令杰
 * Date: 2018/9/10
 */
func GetAddressInfo(address string) btcutil.Amount {
	//获取授权文件，第二个参数针对Windows平台，注意，这里其实调用的是btcd的内部方法，和钱包没什么关系
	//noinspection SpellCheckingInspection
	btcdHomeDir := btcutil.AppDataDir("btcwallet", false)
	certs, err := ioutil.ReadFile(filepath.Join(btcdHomeDir, "rpc.cert"))
	if err != nil {
		log.Fatal(err)
	}

	connCfg := &rpcclient.ConnConfig{
		Host:         config.NoxdHost,
		Endpoint:     config.EndpointWS,
		User:         config.RpcUser,
		Pass:         config.RpcPassword,
		Certificates: certs,
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {

		log.Println(err)
	}
	defer client.Shutdown()

	balance, err := client.GetBalance("acount1")

	return balance
}
