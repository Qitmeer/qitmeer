// Copyright (c) 2017-2018 The qitmeer developers
package node

import (
	"context"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/common/util"
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/core/event"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer/node/service"
	"github.com/Qitmeer/qitmeer/p2p"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/rpc"
	"github.com/Qitmeer/qitmeer/rpc/api"
	"reflect"
	"sync"
	"sync/atomic"
)

// Node works as a server container for all service can be registered.
// such as p2p, rpc, ws etc.
type Node struct {
	started  int32
	shutdown int32
	wg       util.WaitGroupWrapper
	quit     chan struct{}
	lock     sync.RWMutex

	startupTime int64

	// config
	Config *config.Config
	Params *params.Params

	// database layer
	DB database.DB

	// network server
	peerServer *p2p.Service

	services *service.ServiceRegistry

	// api server
	rpcServer *rpc.RpcServer

	// event system
	events event.Feed
}

func NewNode(cfg *config.Config, database database.DB, chainParams *params.Params, shutdownRequestChannel chan struct{}) (*Node, error) {

	n := Node{
		Config:   cfg,
		DB:       database,
		Params:   chainParams,
		quit:     make(chan struct{}),
		services: service.NewServiceRegistry(),
	}

	server, err := p2p.NewService(cfg, &n.events, chainParams)
	if err != nil {
		return nil, err
	}
	n.peerServer = server

	if !cfg.DisableRPC {
		n.rpcServer, err = rpc.NewRPCServer(cfg, &n.events)
		if err != nil {
			return nil, err
		}
		go func() {
			<-n.rpcServer.RequestedProcessShutdown()
			shutdownRequestChannel <- struct{}{}
		}()
	}

	return &n, nil
}

func (n *Node) Stop() error {
	log.Info("Stopping Server")
	// Signal the node quit.
	close(n.quit)

	// stop rpc server
	if n.rpcServer != nil {
		n.rpcServer.Stop()
	}

	// stop p2p server
	n.peerServer.Stop()

	return n.services.StopAll()
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
	// Already started?
	if atomic.AddInt32(&n.started, 1) != 1 {
		return nil
	}

	log.Info("Starting Server")
	// start service one by one
	n.services.StartAll(context.Background())

	// start p2p server
	if err := n.peerServer.Start(); err != nil {
		return err
	}
	// start RPC by service
	if !n.Config.DisableRPC {
		if err := n.startRPC(); err != nil {
			return err
		}
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

// startRPC is a helper method to start all the various RPC endpoint during node
// startup. It's not meant to be called at any time afterwards as it makes certain
// assumptions about the state of the node.
func (n *Node) startRPC() error {
	// Gather all the possible APIs to surface
	apis := []api.API{}
	for _, service := range n.services.GetServices() {
		apis = append(apis, service.APIs()...)
	}
	// Generate the whitelist based on the allowed modules
	whitelist := make(map[string]bool)
	for _, module := range n.Config.Modules {
		whitelist[module] = true
	}

	// Register all the APIs exposed by the services
	for _, api := range apis {
		if whitelist[api.NameSpace] || (len(whitelist) == 0 && api.Public) {
			if err := n.rpcServer.RegisterService(api.NameSpace, api.Service); err != nil {
				return err
			}
			log.Debug(fmt.Sprintf("RPC Service API registered. NameSpace:%s     %s", api.NameSpace, reflect.TypeOf(api.Service)))
		}
	}
	if err := n.rpcServer.Start(); err != nil {
		return err
	}
	return nil
}

// register services as qitmeer Full node
func (n *Node) registerQitmeerFull() error {
	fullNode, err := newQitmeerFullNode(n)
	if err != nil {
		return err
	}
	n.services.RegisterService(fullNode)
	return nil
}

// register services as the qitmeer Light node
func (n *Node) registerQitmeerLight() error {
	lightNode, err := newQitmeerLight(n)
	if err != nil {
		return err
	}
	n.services.RegisterService(lightNode)
	return nil
}

// return qitmeer full
func (n *Node) GetQitmeerFull() *QitmeerFull {
	var qm *QitmeerFull
	if err := n.services.FetchService(&qm); err != nil {
		log.Error(err.Error())
		return nil
	}
	return qm
}
