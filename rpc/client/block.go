package client

import (
	"encoding/json"
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
