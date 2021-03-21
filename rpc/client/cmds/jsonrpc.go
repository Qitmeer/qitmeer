/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package cmds

import (
	"encoding/json"
	"fmt"
)

// These are all service namespace in node
const (
	DefaultServiceNameSpace = "qitmeer"
	MinerNameSpace          = "miner"
	TestNameSpace           = "test"
	LogNameSpace            = "log"
	NotifyNameSpace         = ""
)

type RPCErrorCode int

type RPCError struct {
	Code    RPCErrorCode `json:"code,omitempty"`
	Message string       `json:"message,omitempty"`
}

var _, _ error = RPCError{}, (*RPCError)(nil)

func (e RPCError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

func NewRPCError(code RPCErrorCode, message string) *RPCError {
	return &RPCError{
		Code:    code,
		Message: message,
	}
}

func IsValidIDType(id interface{}) bool {
	switch id.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		string,
		nil:
		return true
	default:
		return false
	}
}

type Request struct {
	Jsonrpc string            `json:"jsonrpc"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
	ID      interface{}       `json:"id"`
}

func NewRequest(id interface{}, method string, params []interface{}) (*Request, error) {
	if !IsValidIDType(id) {
		return nil, fmt.Errorf("the id of type '%T' is invalid", id)
	}

	rawParams := make([]json.RawMessage, 0, len(params))
	for _, param := range params {
		marshalledParam, err := json.Marshal(param)
		if err != nil {
			return nil, err
		}
		rawMessage := json.RawMessage(marshalledParam)
		rawParams = append(rawParams, rawMessage)
	}

	return &Request{
		Jsonrpc: "1.0",
		ID:      id,
		Method:  method,
		Params:  rawParams,
	}, nil
}

type Response struct {
	Result json.RawMessage `json:"result"`
	Error  *RPCError       `json:"error"`
	ID     *interface{}    `json:"id"`
}

func NewResponse(id interface{}, marshalledResult []byte, rpcErr *RPCError) (*Response, error) {
	if !IsValidIDType(id) {
		return nil, fmt.Errorf("the id of type '%T' is invalid", id)
	}

	pid := &id
	return &Response{
		Result: marshalledResult,
		Error:  rpcErr,
		ID:     pid,
	}, nil
}

func MarshalResponse(id interface{}, result interface{}, rpcErr *RPCError) ([]byte, error) {
	marshalledResult, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	response, err := NewResponse(id, marshalledResult, rpcErr)
	if err != nil {
		return nil, err
	}
	return json.Marshal(&response)
}

var (
	ErrRPCInvalidRequest = &RPCError{
		Code:    -32600,
		Message: "Invalid request",
	}
	ErrRPCMethodNotFound = &RPCError{
		Code:    -32601,
		Message: "Method not found",
	}
	ErrRPCInvalidParams = &RPCError{
		Code:    -32602,
		Message: "Invalid parameters",
	}
	ErrRPCInternal = &RPCError{
		Code:    -32603,
		Message: "Internal error",
	}
	// ErrRPCDatabase indicates a database error.
	ErrRPCDatabase = &RPCError{
		Code:    -32001,
		Message: "Database error",
	}
	ErrRPCBlockNotFound = &RPCError{
		Code:    -32002,
		Message: "Block Not Found error",
	}

	ErrInvalidNode = &RPCError{
		Code:    -32003,
		Message: "Invalid Node",
	}
	ErrRPCParse = &RPCError{
		Code:    -32700,
		Message: "Parse error",
	}
	ErrRPCDecodeHexString = &RPCError{
		Code:    -32701,
		Message: "Hex decode error",
	}
)

func InternalRPCError(errStr, context string) *RPCError {
	logStr := errStr
	if context != "" {
		logStr = context + ": " + errStr
	}
	log.Error(logStr)
	return NewRPCError(ErrRPCInternal.Code, errStr)
}

type JsonRequestStatus struct {
	Name        string `json:"name"`
	TotalCalls  int    `json:"totalcalls"`
	TotalTime   string `json:"totaltime"`
	AverageTime string `json:"averagetime"`
	RunningNum  int    `json:"runningnum"`
}
