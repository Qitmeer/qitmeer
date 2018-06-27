// Copyright (c) 2017-2018 The nox developers
package node

import (
	"github.com/noxproject/nox/database"
	"github.com/noxproject/nox/params"
	"github.com/noxproject/nox/rpc"
	"time"
	"github.com/noxproject/nox/log"
	"github.com/noxproject/nox/common/util"
	"github.com/noxproject/nox/p2p"
)

// Node works as a server container for all service can be registered.
// such as p2p, rpc, ws etc.
type Node struct {

	startupTime   int64

	// config
	params        params.Params

	// database layer
	db            database.DB

	// service layer

	// network layer
	peerServer    *p2p.PeerServer

	/// api layer
	rpcServer     *rpc.RpcServer

	wg            util.WaitGroupWrapper
	quit          chan struct{}

}

func (n *Node) Stop() error {
	return nil
}

// WaitForShutdown blocks until the main listener and peer handlers are stopped.
func (n *Node) WaitForShutdown() {
	n.wg.Wait()
}

func (s *Node) Start() {
	log.Info("Starting server")

	// Server startup time. Used for the uptime command for uptime calculation.
	s.startupTime = time.Now().Unix()
}



