package client

import (
	"encoding/json"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	j "github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
)

type FutureCreateRawTransactionResult chan *response

func (r FutureCreateRawTransactionResult) Receive() (string, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return "", err
	}
	var txStr string
	err = json.Unmarshal(res, &txStr)
	if err != nil {
		return "", err
	}
	return txStr, nil
}

func (c *Client) CreateRawTransactionAsync(inputs []j.TransactionInput, amounts j.Amounts, lockTime int64) FutureCreateRawTransactionResult {
	cmd := cmds.NewCreateRawTransactionCmd(inputs, amounts, lockTime)
	return c.sendCmd(cmd)
}

func (c *Client) CreateRawTransaction(inputs []j.TransactionInput, amounts j.Amounts, lockTime int64) (string, error) {
	return c.CreateRawTransactionAsync(inputs, amounts, lockTime).Receive()
}

type FutureDecodeRawTransactionResult chan *response

func (r FutureDecodeRawTransactionResult) Receive() (*j.DecodeRawTransactionResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}
	var tx j.DecodeRawTransactionResult
	err = json.Unmarshal(res, &tx)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (c *Client) DecodeRawTransactionAsync(hexTx string) FutureDecodeRawTransactionResult {
	cmd := cmds.NewDecodeRawTransactionCmd(hexTx)
	return c.sendCmd(cmd)
}

func (c *Client) DecodeRawTransaction(hexTx string) (*j.DecodeRawTransactionResult, error) {
	return c.DecodeRawTransactionAsync(hexTx).Receive()
}

type FutureSendRawTransactionResult chan *response

func (r FutureSendRawTransactionResult) Receive() (*hash.Hash, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	var txHashStr string
	err = json.Unmarshal(res, &txHashStr)
	if err != nil {
		return nil, err
	}
	return hash.NewHashFromStr(txHashStr)
}

func (c *Client) SendRawTransactionAsync(hexTx string, allowHighFees bool) FutureSendRawTransactionResult {
	cmd := cmds.NewSendRawTransactionCmd(hexTx, allowHighFees)
	return c.sendCmd(cmd)
}

func (c *Client) SendRawTransaction(hexTx string, allowHighFees bool) (*hash.Hash, error) {
	return c.SendRawTransactionAsync(hexTx, allowHighFees).Receive()
}

type FutureGetRawTransactionResult chan *response

func (r FutureGetRawTransactionResult) Receive(verbose bool) (interface{}, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}
	if !verbose {
		return string(res), nil
	}
	var tx j.TxRawResult
	err = json.Unmarshal(res, &tx)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (c *Client) GetRawTransactionAsync(txHash string, verbose bool) FutureGetRawTransactionResult {
	cmd := cmds.NewGetRawTransactionCmd(txHash, verbose)
	return c.sendCmd(cmd)
}

func (c *Client) GetRawTransaction(txHash string, verbose bool) (interface{}, error) {
	return c.GetRawTransactionAsync(txHash, verbose).Receive(verbose)
}

func (c *Client) GetRawTransactionRaw(txHash string) (string, error) {
	result, err := c.GetRawTransaction(txHash, false)
	if err != nil {
		return "", err
	}
	tx, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("type is fail")
	}
	return tx, nil
}

func (c *Client) GetRawTransactionVerbose(txHash string) (*j.TxRawResult, error) {
	result, err := c.GetRawTransaction(txHash, true)
	if err != nil {
		return nil, err
	}
	tx, ok := result.(*j.TxRawResult)
	if !ok {
		return nil, fmt.Errorf("type is fail")
	}
	return tx, nil
}

type FutureGetUtxoResult chan *response

func (r FutureGetUtxoResult) Receive() (*j.GetUtxoResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}
	var utxo j.GetUtxoResult
	err = json.Unmarshal(res, &utxo)
	if err != nil {
		return nil, err
	}
	return &utxo, nil
}

func (c *Client) GetUtxoAsync(txHash string, vout uint32, includeMempool bool) FutureGetUtxoResult {
	cmd := cmds.NewGetUtxoCmd(txHash, vout, includeMempool)
	return c.sendCmd(cmd)
}

func (c *Client) GetUtxo(txHash string, vout uint32, includeMempool bool) (*j.GetUtxoResult, error) {
	return c.GetUtxoAsync(txHash, vout, includeMempool).Receive()
}

type FutureGetRawTransactionsResult chan *response

func (r FutureGetRawTransactionsResult) Receive(verbose bool) (interface{}, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}
	if !verbose {
		var txs []string
		err = json.Unmarshal(res, &txs)
		if err != nil {
			return nil, err
		}
		return txs, nil
	}
	var txs []j.GetRawTransactionsResult
	err = json.Unmarshal(res, &txs)
	if err != nil {
		return nil, err
	}
	return txs, nil
}

func (c *Client) GetRawTransactionsAsync(addre string, vinext bool, count uint, skip uint, revers bool, verbose bool, filterAddrs []string) FutureGetRawTransactionsResult {
	cmd := cmds.NewGetRawTransactionsCmd(addre, vinext, count, skip, revers, verbose, filterAddrs)
	return c.sendCmd(cmd)
}

func (c *Client) GetRawTransactions(addre string, vinext bool, count uint, skip uint, revers bool, verbose bool, filterAddrs []string) (interface{}, error) {
	return c.GetRawTransactionsAsync(addre, vinext, count, skip, revers, verbose, filterAddrs).Receive(verbose)
}

func (c *Client) GetRawTransactionsRaw(addre string, vinext bool, count uint, skip uint, revers bool, filterAddrs []string) ([]string, error) {
	result, err := c.GetRawTransactions(addre, vinext, count, skip, revers, false, filterAddrs)
	if err != nil {
		return nil, err
	}
	txs, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("type is fail")
	}
	return txs, nil
}

func (c *Client) GetRawTransactionsVerbose(addre string, vinext bool, count uint, skip uint, revers bool, filterAddrs []string) ([]j.GetRawTransactionsResult, error) {
	result, err := c.GetRawTransactions(addre, vinext, count, skip, revers, true, filterAddrs)
	if err != nil {
		return nil, err
	}
	txs, ok := result.([]j.GetRawTransactionsResult)
	if !ok {
		return nil, fmt.Errorf("type is fail")
	}
	return txs, nil
}

type FutureTxSignResult chan *response

func (r FutureTxSignResult) Receive() (string, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return "", err
	}
	var tx string
	err = json.Unmarshal(res, &tx)
	if err != nil {
		return "", err
	}
	return tx, nil
}

func (c *Client) TxSignAsync(privkeyStr string, rawTxStr string) FutureTxSignResult {
	cmd := cmds.NewTxSignCmd(privkeyStr, rawTxStr)
	return c.sendCmd(cmd)
}

func (c *Client) TxSign(privkeyStr string, rawTxStr string) (string, error) {
	return c.TxSignAsync(privkeyStr, rawTxStr).Receive()
}

type FutureGetMempoolResult chan *response

func (r FutureGetMempoolResult) Receive() ([]string, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}
	var txs []string
	err = json.Unmarshal(res, &txs)
	if err != nil {
		return nil, err
	}
	return txs, nil
}

func (c *Client) GetMempoolAsync(txType string, verbose bool) FutureGetMempoolResult {
	cmd := cmds.NewGetMempoolCmd(txType, verbose)
	return c.sendCmd(cmd)
}

func (c *Client) GetMempool(txType string, verbose bool) ([]string, error) {
	return c.GetMempoolAsync(txType, verbose).Receive()
}
