package testutils

import (
	"github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"strconv"
)

func (c *Client) NodeInfo() (json.InfoNodeResult, error) {
	var result json.InfoNodeResult
	if err := c.Call(&result, "getNodeInfo"); err != nil {
		return result, err
	}
	return result, nil
}

func (c *Client) BlockCount() (uint64, error) {
	var result uint64
	if err := c.Call(&result, "getBlockCount"); err != nil {
		return result, err
	}
	return result, nil
}

func (c *Client) BlockTotal() (uint64, error) {
	var result uint64
	if err := c.Call(&result, "getBlockTotal"); err != nil {
		return result, err
	}
	return result, nil
}

func (c *Client) MainHeight() (uint64, error) {
	var result string
	if err := c.Call(&result, "getMainChainHeight"); err != nil {
		return 0, err
	}
	if height, err := strconv.Atoi(result); err != nil {
		return 0, err
	} else {
		return uint64(height), nil
	}
}

func (c *Client) Generate(num int) ([]string, error) {
	var result []string
	if err := c.Call(&result, "miner_generate", num, pow.PowType(0)); err != nil {
		return result, err
	}
	return result, nil
}
