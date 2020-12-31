package client

import (
	"encoding/json"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	j "github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
	"strconv"
)

type FutureGetBlockCountResult chan *response

func (r FutureGetBlockCountResult) Receive() (int64, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return 0, err
	}

	// Unmarshal the result as an int64.
	var count int64
	err = json.Unmarshal(res, &count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (c *Client) GetBlockCountAsync() FutureGetBlockCountResult {
	cmd := cmds.NewGetBlockCountCmd()
	return c.sendCmd(cmd)
}

func (c *Client) GetBlockCount() (int64, error) {
	return c.GetBlockCountAsync().Receive()
}

type FutureGetBlockHashResult chan *response

func (r FutureGetBlockHashResult) Receive() (*hash.Hash, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal the result as a string-encoded sha.
	var blkHashStr string
	err = json.Unmarshal(res, &blkHashStr)
	if err != nil {
		return nil, err
	}
	return hash.NewHashFromStr(blkHashStr)
}

func (c *Client) GetBlockHashAsync(order uint) FutureGetBlockHashResult {
	cmd := cmds.NewGetBlockhashCmd(order)
	return c.sendCmd(cmd)
}

func (c *Client) GetBlockHash(order uint) (*hash.Hash, error) {
	return c.GetBlockHashAsync(order).Receive()
}

type FutureGetBlockhashByRangeResult chan *response

func (r FutureGetBlockhashByRangeResult) Receive() ([]*hash.Hash, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal the result as a string-encoded sha.
	var blksHashStr []string
	err = json.Unmarshal(res, &blksHashStr)
	if err != nil {
		return nil, err
	}
	result := []*hash.Hash{}
	for _, blkHashStr := range blksHashStr {
		h, err := hash.NewHashFromStr(blkHashStr)
		if err != nil {
			return nil, err
		}
		result = append(result, h)
	}
	return result, nil
}

func (c *Client) GetBlockhashByRangeAsync(start uint, end uint) FutureGetBlockhashByRangeResult {
	cmd := cmds.NewGetBlockhashByRangeCmd(start, end)
	return c.sendCmd(cmd)
}

func (c *Client) GetBlockhashByRange(start uint, end uint) ([]*hash.Hash, error) {
	return c.GetBlockhashByRangeAsync(start, end).Receive()
}

type FutureGetBlockResult chan *response

func (r FutureGetBlockResult) Receive(verbose bool, fullTx bool) (interface{}, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}
	if !verbose {
		return string(res), nil
	}
	if fullTx {
		var blk j.BlockVerboseResult
		err = json.Unmarshal(res, &blk)
		if err != nil {
			return nil, err
		}
		return &blk, nil
	}
	var blk j.BlockResult
	err = json.Unmarshal(res, &blk)
	if err != nil {
		return nil, err
	}
	return &blk, nil
}

func (c *Client) GetBlockAsync(h string, verbose bool, inclTx bool, fullTx bool) FutureGetBlockResult {
	cmd := cmds.NewGetBlockCmd(h, verbose, inclTx, fullTx)
	return c.sendCmd(cmd)
}

func (c *Client) GetBlock(h string, verbose bool, inclTx bool, fullTx bool) (interface{}, error) {
	return c.GetBlockAsync(h, verbose, inclTx, fullTx).Receive(verbose, fullTx)
}

func (c *Client) GetBlockRaw(h string, inclTx bool) (string, error) {
	result, err := c.GetBlock(h, false, inclTx, false)
	if err != nil {
		return "", err
	}
	blk, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("type is fail")
	}
	return blk, nil
}

func (c *Client) GetBlockSimpleTx(h string, inclTx bool) (*j.BlockResult, error) {
	result, err := c.GetBlock(h, true, inclTx, false)
	if err != nil {
		return nil, err
	}
	blk, ok := result.(*j.BlockResult)
	if !ok {
		return nil, fmt.Errorf("type is fail")
	}
	return blk, nil
}

func (c *Client) GetBlockFullTx(h string, inclTx bool) (*j.BlockVerboseResult, error) {
	result, err := c.GetBlock(h, true, inclTx, true)
	if err != nil {
		return nil, err
	}
	blk, ok := result.(*j.BlockVerboseResult)
	if !ok {
		return nil, fmt.Errorf("type is fail")
	}
	return blk, nil
}

type FutureGetBlockByOrderResult chan *response

func (r FutureGetBlockByOrderResult) Receive(verbose bool, fullTx bool) (interface{}, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}
	if !verbose {
		return string(res), nil
	}
	if fullTx {
		var blk j.BlockVerboseResult
		err = json.Unmarshal(res, &blk)
		if err != nil {
			return nil, err
		}
		return &blk, nil
	}
	var blk j.BlockResult
	err = json.Unmarshal(res, &blk)
	if err != nil {
		return nil, err
	}
	return &blk, nil
}

func (c *Client) GetBlockByOrderAsync(order uint, verbose bool, inclTx bool, fullTx bool) FutureGetBlockByOrderResult {
	cmd := cmds.NewGetBlockByOrderCmd(order, verbose, inclTx, fullTx)
	return c.sendCmd(cmd)
}

func (c *Client) GetBlockByOrder(order uint, verbose bool, inclTx bool, fullTx bool) (interface{}, error) {
	return c.GetBlockByOrderAsync(order, verbose, inclTx, fullTx).Receive(verbose, fullTx)
}

func (c *Client) GetBlockByOrderRaw(order uint, inclTx bool) (string, error) {
	result, err := c.GetBlockByOrder(order, false, inclTx, false)
	if err != nil {
		return "", err
	}
	blk, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("type is fail")
	}
	return blk, nil
}

func (c *Client) GetBlockByOrderSimpleTx(order uint, inclTx bool) (*j.BlockResult, error) {
	result, err := c.GetBlockByOrder(order, true, inclTx, false)
	if err != nil {
		return nil, err
	}
	blk, ok := result.(*j.BlockResult)
	if !ok {
		return nil, fmt.Errorf("type is fail")
	}
	return blk, nil
}

func (c *Client) GetBlockByOrderFullTx(order uint, inclTx bool) (*j.BlockVerboseResult, error) {
	result, err := c.GetBlockByOrder(order, true, inclTx, true)
	if err != nil {
		return nil, err
	}
	blk, ok := result.(*j.BlockVerboseResult)
	if !ok {
		return nil, fmt.Errorf("type is fail")
	}
	return blk, nil
}

type FutureGetBlockV2Result chan *response

func (r FutureGetBlockV2Result) Receive(verbose bool, fullTx bool) (interface{}, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}
	if !verbose {
		return string(res), nil
	}
	if fullTx {
		var blk j.BlockVerboseResult
		err = json.Unmarshal(res, &blk)
		if err != nil {
			return nil, err
		}
		return &blk, nil
	}
	var blk j.BlockResult
	err = json.Unmarshal(res, &blk)
	if err != nil {
		return nil, err
	}
	return &blk, nil
}

func (c *Client) GetBlockV2Async(h string, verbose bool, inclTx bool, fullTx bool) FutureGetBlockResult {
	cmd := cmds.NewGetBlockV2Cmd(h, verbose, inclTx, fullTx)
	return c.sendCmd(cmd)
}

func (c *Client) GetBlockV2(h string, verbose bool, inclTx bool, fullTx bool) (interface{}, error) {
	return c.GetBlockV2Async(h, verbose, inclTx, fullTx).Receive(verbose, fullTx)
}

func (c *Client) GetBlockV2Raw(h string, inclTx bool) (string, error) {
	result, err := c.GetBlockV2(h, false, inclTx, false)
	if err != nil {
		return "", err
	}
	blk, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("type is fail")
	}
	return blk, nil
}

func (c *Client) GetBlockV2SimpleTx(h string, inclTx bool) (*j.BlockResult, error) {
	result, err := c.GetBlockV2(h, true, inclTx, false)
	if err != nil {
		return nil, err
	}
	blk, ok := result.(*j.BlockResult)
	if !ok {
		return nil, fmt.Errorf("type is fail")
	}
	return blk, nil
}

func (c *Client) GetBlockV2FullTx(h string, inclTx bool) (*j.BlockVerboseResult, error) {
	result, err := c.GetBlockV2(h, true, inclTx, true)
	if err != nil {
		return nil, err
	}
	blk, ok := result.(*j.BlockVerboseResult)
	if !ok {
		return nil, fmt.Errorf("type is fail")
	}
	return blk, nil
}

type FutureGetBestBlockHashResult chan *response

func (r FutureGetBestBlockHashResult) Receive() (*hash.Hash, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal the result as a string-encoded sha.
	var blkHashStr string
	err = json.Unmarshal(res, &blkHashStr)
	if err != nil {
		return nil, err
	}
	return hash.NewHashFromStr(blkHashStr)
}

func (c *Client) GetBestBlockHashAsync() FutureGetBestBlockHashResult {
	cmd := cmds.NewGetBestBlockHashCmd()
	return c.sendCmd(cmd)
}

func (c *Client) GetBestBlockHash() (*hash.Hash, error) {
	return c.GetBestBlockHashAsync().Receive()
}

type FutureGetBlockTotalResult chan *response

func (r FutureGetBlockTotalResult) Receive() (int64, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return 0, err
	}

	// Unmarshal the result as an int64.
	var total int64
	err = json.Unmarshal(res, &total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (c *Client) GetBlockTotalAsync() FutureGetBlockTotalResult {
	cmd := cmds.NewGetBlockTotalCmd()
	return c.sendCmd(cmd)
}

func (c *Client) GetBlockTotal() (int64, error) {
	return c.GetBlockTotalAsync().Receive()
}

type FutureGetBlockHeaderResult chan *response

func (r FutureGetBlockHeaderResult) Receive(verbose bool) (interface{}, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}
	if !verbose {
		return string(res), nil
	}
	var header j.GetBlockHeaderVerboseResult
	err = json.Unmarshal(res, &header)
	if err != nil {
		return nil, err
	}
	return &header, nil
}

func (c *Client) GetBlockHeaderAsync(hash string, verbose bool) FutureGetBlockHeaderResult {
	cmd := cmds.NewGetBlockHeaderCmd(hash, verbose)
	return c.sendCmd(cmd)
}

func (c *Client) GetBlockHeader(hash string, verbose bool) (interface{}, error) {
	return c.GetBlockHeaderAsync(hash, verbose).Receive(verbose)
}

func (c *Client) GetBlockHeaderRaw(hash string) (string, error) {
	result, err := c.GetBlockHeader(hash, false)
	if err != nil {
		return "", err
	}
	blk, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("type is fail")
	}
	return blk, nil
}

func (c *Client) GetBlockHeaderVerbose(hash string) (*j.GetBlockHeaderVerboseResult, error) {
	result, err := c.GetBlockHeader(hash, true)
	if err != nil {
		return nil, err
	}
	header, ok := result.(*j.GetBlockHeaderVerboseResult)
	if !ok {
		return nil, fmt.Errorf("type is fail")
	}
	return header, nil
}

type FutureIsOnMainChainResult chan *response

func (r FutureIsOnMainChainResult) Receive() (bool, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return false, err
	}
	var result string
	err = json.Unmarshal(res, &result)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(result)
}

func (c *Client) IsOnMainChainAsync(h string) FutureIsOnMainChainResult {
	cmd := cmds.NewIsOnMainChainCmd(h)
	return c.sendCmd(cmd)
}

func (c *Client) IsOnMainChain(h string) (bool, error) {
	return c.IsOnMainChainAsync(h).Receive()
}

type FutureGetMainChainHeightResult chan *response

func (r FutureGetMainChainHeightResult) Receive() (int64, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return 0, err
	}

	// Unmarshal the result as an int64.
	var heightStr string
	err = json.Unmarshal(res, &heightStr)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(heightStr, 10, 64)
}

func (c *Client) GetMainChainHeightAsync() FutureGetMainChainHeightResult {
	cmd := cmds.NewGetMainChainHeightCmd()
	return c.sendCmd(cmd)
}

func (c *Client) GetMainChainHeight() (int64, error) {
	return c.GetMainChainHeightAsync().Receive()
}

type FutureGetBlockWeightResult chan *response

func (r FutureGetBlockWeightResult) Receive() (int64, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return 0, err
	}

	// Unmarshal the result as an int64.
	var weightStr string
	err = json.Unmarshal(res, &weightStr)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(weightStr, 10, 64)
}

func (c *Client) GetBlockWeightAsync(h string) FutureGetBlockWeightResult {
	cmd := cmds.NewGetBlockWeightCmd(h)
	return c.sendCmd(cmd)
}

func (c *Client) GetBlockWeight(h string) (int64, error) {
	return c.GetBlockWeightAsync(h).Receive()
}
