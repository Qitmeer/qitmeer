// Copyright (c) 2017-2018 The nox developers

package rpc

import (
	"github.com/noxproject/nox/config"
	"github.com/noxproject/nox/common/util"
	"sync"
)

// RpcServer provides a concurrent safe RPC server to a chain server.
type RpcServer struct {

	started                int32
	shutdown               int32
	wg                     util.WaitGroupWrapper
	quit                   chan int
	statusLock             sync.RWMutex

	numClients             int32
	statusLines            map[int]string
	requestProcessShutdown chan struct{}

}

// newRPCServer returns a new instance of the rpcServer struct.
func NewRPCServer(cfg *config.Config) (*RpcServer, error) {
	rpc := RpcServer{
		statusLines:            make(map[int]string),
		requestProcessShutdown: make(chan struct{}),
		quit: make(chan int),
	}
	return &rpc, nil
}


