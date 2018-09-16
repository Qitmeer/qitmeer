package handler

import (
	"github.com/noxproject/nox/middleware/config"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
)

/**
 * Author: 傅令杰
 * Date: 2018/9/10
 */

/**
根据交易hash查询交易记录
*/
func GetTransaction(transactionHashStr string) *btcjson.GetTransactionResult {
	//noinspection SpellCheckingInspection
	walletHomeDir := btcutil.AppDataDir("btcwallet", false)
	certs, err := ioutil.ReadFile(filepath.Join(walletHomeDir, "rpc.cert"))
	if err != nil {
		log.Fatal(err)
	}

	//certsStr := string(certs)
	//log.Println(certsStr)

	connCfg := &rpcclient.ConnConfig{
		Host:         config.WalletHost,
		Endpoint:     config.EndpointWS,
		User:         config.RpcUser,
		Pass:         config.RpcPassword,
		Certificates: certs,
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {

		log.Fatal(err)
	}
	defer client.Shutdown()

	blockHash, err := chainhash.NewHashFromStr(transactionHashStr)
	transactionResult, err := client.GetTransaction(blockHash)
	if err != nil {
		log.Println(err)
	}

	return transactionResult
}
