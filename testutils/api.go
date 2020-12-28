package testutils

import "github.com/Qitmeer/qitmeer/core/json"

func (c *Client) GetNodeInfo() (json.InfoNodeResult, error) {
	var result json.InfoNodeResult
	if err := c.Call(&result, "getNodeInfo"); err != nil {
		return result, err
	}
	return result, nil
}
