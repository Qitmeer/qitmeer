package client

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	j "github.com/Qitmeer/qng-core/core/json"
	"github.com/Qitmeer/qng-core/core/types"
	"github.com/Qitmeer/qng-core/core/types/pow"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
	"strings"
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

type FutureGetRemoteGBTCmdResult chan *response

func (r FutureGetRemoteGBTCmdResult) Receive() (*types.BlockHeader, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}
	serialized, err := hex.DecodeString(strings.ReplaceAll(string(res), `"`, ""))
	if err != nil {
		fmt.Println(err.Error(), len(string(res)), string(res))
		return nil, fmt.Errorf(err.Error())
	}
	var header types.BlockHeader
	err = header.Deserialize(bytes.NewReader(serialized))
	if err != nil {
		return nil, err
	}
	return &header, nil
}

func (c *Client) GetRemoteGBTAsync(powType pow.PowType) FutureGetRemoteGBTCmdResult {
	cmd := cmds.NewGetRemoteGBTCmd(powType)
	return c.sendCmd(cmd)
}

func (c *Client) GetRemoteGBT(powType pow.PowType) (*types.BlockHeader, error) {
	return c.GetRemoteGBTAsync(powType).Receive()
}

type FutureSubmitBlockHeaderResult chan *response

func (r FutureSubmitBlockHeaderResult) Receive() (*j.SubmitBlockResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}
	var result j.SubmitBlockResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) SubmitBlockHeaderAsync(header *types.BlockHeader) FutureSubmitBlockHeaderResult {
	cmd := cmds.NewSubmitBlockHeaderCmd(header)
	return c.sendCmd(cmd)
}

func (c *Client) SubmitBlockHeader(header *types.BlockHeader) (*j.SubmitBlockResult, error) {
	return c.SubmitBlockHeaderAsync(header).Receive()
}
