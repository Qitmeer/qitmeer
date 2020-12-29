package client

import (
	"encoding/json"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
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
