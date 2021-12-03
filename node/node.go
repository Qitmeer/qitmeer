// Copyright (c) 2017-2018 The qitmeer developers
package node

import (
	"github.com/Qitmeer/qng-core/common/roughtime"
	"github.com/Qitmeer/qng-core/common/util"
	"github.com/Qitmeer/qng-core/config"
	"github.com/Qitmeer/qitmeer/core/event"
	"github.com/Qitmeer/qng-core/database"
	"github.com/Qitmeer/qitmeer/node/service"
	"github.com/Qitmeer/qng-core/params"
	"sync"
)

// Node works as a server container for all service can be registered.
// such as p2p, rpc, ws etc.
type Node struct {
	service.Service
	wg   util.WaitGroupWrapper
	quit chan struct{}
	lock sync.RWMutex

	startupTime int64

	// config
	Config *config.Config
	Params *params.Params

	// database layer
	DB database.DB

	// event system
	events event.Feed

	shutdownRequestChannel chan struct{}
}

func NewNode(cfg *config.Config, database database.DB, chainParams *params.Params, shutdownRequestChannel chan struct{}) (*Node, error) {

	n := Node{
		Config:                 cfg,
		DB:                     database,
		Params:                 chainParams,
		quit:                   make(chan struct{}),
		shutdownRequestChannel: shutdownRequestChannel,
	}
	n.InitServices()
	return &n, nil
}

func (n *Node) Stop() error {
	log.Info("Stopping Server")
	if err := n.Service.Stop(); err != nil {
		return err
	}
	// Signal the node quit.
	close(n.quit)

	return nil
}

// WaitForShutdown blocks until the main listener and peer handlers are stopped.
func (n *Node) WaitForShutdown() {
	log.Info("Waiting for server shutdown")
	n.wg.Wait()
}

func (n *Node) nodeEventHandler() {
	<-n.quit
	log.Trace("node stop event (quit) received")
}

func (n *Node) Start() error {
	n.lock.Lock()
	defer n.lock.Unlock()
	log.Info("Starting Node")
	// Already started?
	if err := n.Service.Start(); err != nil {
		return err
	}

	// Finished node start
	// Server startup time. Used for the uptime command for uptime calculation.
	n.startupTime = roughtime.Now().Unix()
	n.wg.Wrap(n.nodeEventHandler)

	return nil
}

func (n *Node) RegisterService() error {
	if n.Config.LightNode {
		return n.registerQitmeerLight()
	}
	return n.registerQitmeerFull()
}

// register services as qitmeer Full node
func (n *Node) registerQitmeerFull() error {
	fullNode, err := newQitmeerFullNode(n)
	if err != nil {
		return err
	}
	n.Services().RegisterService(fullNode)
	return nil
}

// register services as the qitmeer Light node
func (n *Node) registerQitmeerLight() error {
	lightNode, err := newQitmeerLight(n)
	if err != nil {
		return err
	}
	n.Services().RegisterService(lightNode)
	return nil
}

// return qitmeer full
func (n *Node) GetQitmeerFull() *QitmeerFull {
	var qm *QitmeerFull
	if err := n.Services().FetchService(&qm); err != nil {
		log.Error(err.Error())
		return nil
	}
	return qm
}
