// Copyright (c) 2017-2018 The qitmeer developers
package node

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/common/util"
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/core/event"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer/p2p"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/rpc"
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

	// service layer
	// Service constructors (in dependency order)
	svcConstructors []ServiceConstructor
	// Currently registered & running services
	runningSvcs map[reflect.Type]Service

	// api server
	rpcServer *rpc.RpcServer

	// event system
	events event.Feed
}

func NewNode(cfg *config.Config, database database.DB, chainParams *params.Params, shutdownRequestChannel chan struct{}) (*Node, error) {

	n := Node{
		Config: cfg,
		DB:     database,
		Params: chainParams,
		quit:   make(chan struct{}),
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

	// stop rpc server
	if n.rpcServer != nil {
		n.rpcServer.Stop()
	}

	// stop p2p server
	n.peerServer.Stop()

	failure := &ServiceStopError{
		Services: make(map[reflect.Type]error),
	}
	// stop all service
	for kind, service := range n.runningSvcs {
		if err := service.Stop(); err != nil {
			failure.Services[kind] = err
		}
		log.Debug("Service stopped", "service", kind)
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
			return fmt.Errorf("duplicate Service, kind=%s}", kind)
		}
		services[kind] = service
	}
	// start service one by one
	startedSvs := []reflect.Type{}
	for kind, service := range services {
		if err := service.Start(); err != nil {
			// stopping all started service if upon failure
			for _, kind := range startedSvs {
				services[kind].Stop()
			}
			return err
		}
		// Mark the service has been started
		startedSvs = append(startedSvs, kind)
		log.Debug("Node service started", "service", kind)
	}
	n.runningSvcs = services

	// start p2p server
	if err := n.peerServer.Start(); err != nil {
		return err
	}
	// start RPC by service
	if !n.Config.DisableRPC {
		if err := n.startRPC(services); err != nil {
			for _, service := range services {
				service.Stop()
			}
			n.peerServer.Stop()
			return err
		}
	}

	// Finished node start
	// Server startup time. Used for the uptime command for uptime calculation.
	n.startupTime = roughtime.Now().Unix()
	n.wg.Wrap(n.nodeEventHandler)

	return nil
}

func (n *Node) register(sc ServiceConstructor) error {
	n.lock.Lock()
	defer n.lock.Unlock()
	// Already started?
	if atomic.LoadInt32(&n.started) == 1 {
		return fmt.Errorf("node has already been started")
	}
	n.svcConstructors = append(n.svcConstructors, sc)
	log.Debug("Register service to node", "service", sc)
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
func (n *Node) startRPC(services map[reflect.Type]Service) error {
	// Gather all the possible APIs to surface
	apis := []rpc.API{}
	for _, service := range services {
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
	err := n.register(NewServiceConstructor("qitmeer",
		func(ctx *ServiceContext) (Service, error) {
			fullNode, err := newQitmeerFullNode(n)
			return fullNode, err
		}))
	return err
}

// register services as the qitmeer Light node
func (n *Node) registerQitmeerLight() error {
	err := n.register(NewServiceConstructor("qitmeer-light",
		func(ctx *ServiceContext) (Service, error) {
			lightNode, err := newQitmeerLight(n)
			return lightNode, err
		}))
	return err
}

// return qitmeer full
func (n *Node) GetQitmeerFull() *QitmeerFull {
	for _, server := range n.runningSvcs {
		fullqm := server.(*QitmeerFull)
		if fullqm != nil {
			return fullqm
		}
	}
	return nil
}
