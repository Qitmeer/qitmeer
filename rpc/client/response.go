package client

import (
	"encoding/json"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
)

type rawResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *cmds.RPCError  `json:"error"`
}

func (r rawResponse) result() (result []byte, err error) {
	if r.Error != nil {
		return nil, r.Error
	}
	return r.Result, nil
}

// response is the raw bytes of a JSON-RPC result, or the error if the response
// error object was non-null.
type response struct {
	result []byte
	err    error
}
