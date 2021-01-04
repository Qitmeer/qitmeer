package client

import (
	"encoding/json"
	j "github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
)

type FutureGetBlockTemplateResult chan *response

func (r FutureGetBlockTemplateResult) Receive() (*j.GetBlockTemplateResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}
	var template j.GetBlockTemplateResult
	err = json.Unmarshal(res, &template)
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (c *Client) GetBlockTemplateAsync(capabilities []string, powType byte) FutureGetBlockTemplateResult {
	cmd := cmds.NewGetBlockTemplateCmd(capabilities, powType)
	return c.sendCmd(cmd)
}

func (c *Client) GetBlockTemplate(capabilities []string, powType byte) (*j.GetBlockTemplateResult, error) {
	return c.GetBlockTemplateAsync(capabilities, powType).Receive()
}

type FutureSubmitBlockResult chan *response

func (r FutureSubmitBlockResult) Receive() (string, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return "", err
	}
	var result string
	err = json.Unmarshal(res, &result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (c *Client) SubmitBlockAsync(hexBlock string) FutureSubmitBlockResult {
	cmd := cmds.NewSubmitBlockCmd(hexBlock)
	return c.sendCmd(cmd)
}

func (c *Client) SubmitBlock(hexBlock string) (string, error) {
	return c.SubmitBlockAsync(hexBlock).Receive()
}

type FutureGenerateCmdResult chan *response

func (r FutureGenerateCmdResult) Receive() ([]string, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}
	var result []string
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GenerateAsync(numBlocks uint32, powType pow.PowType) FutureGenerateCmdResult {
	cmd := cmds.NewGenerateCmd(numBlocks, powType)
	return c.sendCmd(cmd)
}

func (c *Client) Generate(numBlocks uint32, powType pow.PowType) ([]string, error) {
	return c.GenerateAsync(numBlocks, powType).Receive()
}
