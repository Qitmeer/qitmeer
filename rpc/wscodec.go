package rpc

import (
	"encoding/json"
	"reflect"
	"sync"
)

type WSCodec struct {
	msg    json.RawMessage
	client *wsClient
	closer sync.Once        // close closed channel once
	closed chan interface{} // closed on Close
}

func (c *WSCodec) ReadRequestHeaders() ([]rpcRequest, bool, Error) {
	return parseRequest(c.msg)
}

func (c *WSCodec) ParseRequestArguments(argTypes []reflect.Type, params interface{}) ([]reflect.Value, Error) {
	if args, ok := params.(json.RawMessage); !ok {
		return nil, &invalidParamsError{"Invalid params supplied"}
	} else {
		return parsePositionalArguments(args, argTypes)
	}
}

// CreateResponse will create a JSON-RPC success response with the given id and reply as result.
func (c *WSCodec) CreateResponse(id interface{}, reply interface{}) interface{} {
	return &jsonSuccessResponse{Version: jsonrpcVersion, Id: id, Result: reply}
}

// CreateErrorResponse will create a JSON-RPC error response with the given id and error.
func (c *WSCodec) CreateErrorResponse(id interface{}, err Error) interface{} {
	return &jsonErrResponse{Version: jsonrpcVersion, Id: id, Error: jsonError{Code: err.ErrorCode(), Message: err.Error()}}
}

// CreateErrorResponseWithInfo will create a JSON-RPC error response with the given id and error.
// info is optional and contains additional information about the error. When an empty string is passed it is ignored.
func (c *WSCodec) CreateErrorResponseWithInfo(id interface{}, err Error, info interface{}) interface{} {
	return &jsonErrResponse{Version: jsonrpcVersion, Id: id,
		Error: jsonError{Code: err.ErrorCode(), Message: err.Error(), Data: info}}
}

// CreateNotification will create a JSON-RPC notification with the given subscription id and event as params.
func (c *WSCodec) CreateNotification(subid, namespace string, event interface{}) interface{} {
	return &jsonNotification{Version: jsonrpcVersion, Method: namespace + notificationMethodSuffix,
		Params: jsonSubscription{Subscription: subid, Result: event}}
}

// Write message to client
func (c *WSCodec) Write(res interface{}) error {
	result, err := json.Marshal(res)
	if err != nil {
		return err
	}
	c.client.SendMessage(result, nil)
	return nil
}

// Close the underlying connection
func (c *WSCodec) Close() {
	c.closer.Do(func() {
		close(c.closed)
	})
}

// Closed returns a channel which will be closed when Close is called
func (c *WSCodec) Closed() <-chan interface{} {
	return c.closed
}

func NewWSCodec(msg []byte, client *wsClient) *WSCodec {
	return &WSCodec{msg: json.RawMessage(msg), client: client, closed: make(chan interface{})}
}
