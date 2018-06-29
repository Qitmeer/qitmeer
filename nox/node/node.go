// Copyright (c) 2017-2018 The nox developers
package node

import (
	"github.com/noxproject/nox/database"
	"github.com/noxproject/nox/params"
	"github.com/noxproject/nox/rpc"
	"github.com/noxproject/nox/log"
	"github.com/noxproject/nox/common/util"
	"github.com/noxproject/nox/p2p"
	"reflect"
	"fmt"
	"github.com/noxproject/nox/config"
	"sync/atomic"
	"sync"
	"time"
)

// Node works as a server container for all service can be registered.
// such as p2p, rpc, ws etc.
type Node struct {

	started       int32
	shutdown      int32
	wg            util.WaitGroupWrapper
	quit          chan struct{}
	lock          sync.RWMutex

	startupTime   int64

	// config
	config        *config.Config
	params        *params.Params

	// database layer
	db            database.DB

	// network server
	peerServer    *p2p.PeerServer

	// service layer
	// Service constructors (in dependency order)
	svcConstructors   []ServiceConstructor
    // Currently registered & running services
	runningSvcs   map[reflect.Type]Service

	// api server
	rpcServer     *rpc.RpcServer


}

func NewNode(cfg *config.Config, database database.DB, chainParams *params.Params) (*Node,error) {

	n := Node{
		config: cfg,
		db    : database,
		params: chainParams,
		quit:   make(chan struct{}),
	}

	server, err := p2p.NewPeerServer(cfg,database,chainParams)
	if err != nil {
		return nil, err
	}
	n.peerServer = server

	if !cfg.DisableRPC {
		n.rpcServer, err = rpc.NewRPCServer(cfg)
		if err != nil {
			return nil, err
		}
	}

    return &n, nil
}

func (n *Node) Stop() error {
	log.Info("Stopping Server")

	failure := &ServiceStopError{
		Services: make(map[reflect.Type]error),
	}
	for kind, service := range n.runningSvcs {
		if err := service.Stop(); err != nil {
			failure.Services[kind] = err
		}
		log.Debug("Service stopped", "service",kind)
	}
	// Signal the node quit.
	close(n.quit)

	if len(failure.Services) > 0 {
		return failure
	}
	return nil
}

// WaitForShutdown blocks until the main listener and peer handlers are stopped.
func (n *Node) WaitForShutdown() {
	log.Info("Waiting for server shutdown")
	n.wg.Wait()
}

func (n *Node) nodeEventHandler() {
	for {
		select {
			case <-n.quit:
				log.Trace("node stop event (quit) received")
				return
		}
	}
}


func (n *Node) Start() error {
	n.lock.Lock()
	defer n.lock.Unlock()
	// Already started?
	if atomic.AddInt32(&n.started, 1) != 1 {
		return nil
	}

	log.Info("Starting Server")

	// start p2p server
	if err :=n.peerServer.Start(); err != nil {
		return err
	}
	log.Info("P2P server started")

	// Initialize every service by calling the registered service constructors & save to services
	services := make(map[reflect.Type]Service)
	for _, c := range n.svcConstructors {
		ctx := &ServiceContext{}
		// Construct and save the service
		service, err := c.initFunc(ctx)
		if err != nil {
			return err
		}
		kind := reflect.TypeOf(service)
		if _, exists := services[kind]; exists {
			return fmt.Errorf("duplicate Service, kind=%s}",kind)
		}
		services[kind] = service
	}
	// start service one by one
	startedSvs := []reflect.Type{}
	for kind, service := range services {
		if err := service.Start(n.peerServer); err != nil {
			// stopping all started service if upon failure
			for _, kind := range startedSvs {
				services[kind].Stop()
			}
			return err
		}
		// Mark the service has been started
		startedSvs = append(startedSvs, kind)
		log.Debug("Node service started", "service",kind)
	}
	n.runningSvcs = services

	// start RPC by service
	if !n.config.DisableRPC {
		if err:= n.startRPC(services); err != nil {
			for _, service := range services {
				service.Stop()
			}
			n.peerServer.Stop()
			return err
		}
	}

	// Finished node start
	// Server startup time. Used for the uptime command for uptime calculation.
	n.startupTime = time.Now().Unix()
	n.wg.Wrap(n.nodeEventHandler)

	return nil
}


func (n *Node) Register(sc ServiceConstructor) error {
	n.lock.Lock()
	defer n.lock.Unlock()
	// Already started?
	if atomic.LoadInt32(&n.started) == 1 {
		return fmt.Errorf("node has already been started")
	}
	n.svcConstructors = append(n.svcConstructors, sc)
	log.Debug("Register service to node","service",sc)
	return nil
}

// startRPC is a helper method to start all the various RPC endpoint during node
// startup. It's not meant to be called at any time afterwards as it makes certain
// assumptions about the state of the node.
func (n *Node) startRPC(services map[reflect.Type]Service) error {
	// Gather all the possible APIs to surface
	apis := []rpc.API{}
	for _, service := range services {
		apis = append(apis, service.APIs()...)
	}
	// Register all the APIs exposed by the services
	for _, api := range apis {
		if err := n.rpcServer.RegisterService(api.NameSpace, api.Service); err != nil {
			return err
		}
		log.Debug("RPC Service API registered", "api", reflect.TypeOf(api.Service))
	}
	if err := n.rpcServer.Start(); err != nil {
		return err
	}
	return nil
}
