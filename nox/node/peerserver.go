// Copyright (c) 2017-2018 The nox developers
package node

import (
	"github.com/noxproject/nox/database"
	"github.com/noxproject/nox/params"
	"github.com/noxproject/nox/rpc"
	"sync"
	"time"
	"github.com/noxproject/nox/log"
)

// Node works as a server container for all service can be registered.
// such as p2p, rpc, ws etc.
type Node struct {
	peerServer    *peerServer
	rpcServer     *rpc.RpcServer
	wg            sync.WaitGroup
	startupTime   int64
}

// newNode returns a new nox node which configured to listen on addr for the
// nox network type specified by the network Params.
func NewNode(listenAddrs []string, db database.DB, chainParams *params.Params, interrupt <-chan struct{}) (*Node, error) {
	node := Node{
	}
	return &node,nil
}

func (n *Node) Stop() error {
	return nil
}

// WaitForShutdown blocks until the main listener and peer handlers are stopped.
func (n *Node) WaitForShutdown() {
	n.wg.Wait()
}

func (s *Node) Start() {
	log.Trace("Starting server")

	// Server startup time. Used for the uptime command for uptime calculation.
	s.startupTime = time.Now().Unix()
}

// Use start to begin accepting connections from peers.

// peer server handling communications to and from nox peers.
type peerServer struct{
	// The following variables must only be used atomically.
	// Putting the uint64s first makes them 64-bit aligned for 32-bit systems.
	bytesReceived uint64 // Total bytes received from all peers since start.
	bytesSent     uint64 // Total bytes sent by all peers since start.
	started       int32
	shutdown      int32
	shutdownSched int32
}
