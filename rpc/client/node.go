package client

import (
	"encoding/json"
	j "github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
)

type FutureGetNodeInfoResult chan *response

func (r FutureGetNodeInfoResult) Receive() (*j.InfoNodeResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a getinfo result object.
	var infoRes j.InfoNodeResult
	err = json.Unmarshal(res, &infoRes)
	if err != nil {
		return nil, err
	}

	return &infoRes, nil
}

func (c *Client) GetNodeInfoAsync() FutureGetNodeInfoResult {
	cmd := cmds.NewGetNodeInfoCmd()
	return c.sendCmd(cmd)
}

func (c *Client) GetNodeInfo() (*j.InfoNodeResult, error) {
	return c.GetNodeInfoAsync().Receive()
}

type FutureGetPeerInfoResult chan *response

func (r FutureGetPeerInfoResult) Receive() ([]j.GetPeerInfoResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a getinfo result object.
	var infoRes []j.GetPeerInfoResult
	err = json.Unmarshal(res, &infoRes)
	if err != nil {
		return nil, err
	}

	return infoRes, nil
}

func (c *Client) GetPeerInfoAsync() FutureGetPeerInfoResult {
	cmd := cmds.NewGetPeerInfoCmd()
	return c.sendCmd(cmd)
}

func (c *Client) GetPeerInfo() ([]j.GetPeerInfoResult, error) {
	return c.GetPeerInfoAsync().Receive()
}
