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
	"reflect"
	"fmt"
	"github.com/noxproject/nox/config"
	"sync/atomic"
	"sync"
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
	params        *params.Params
	// database layer
	db            database.DB
	// service layer
	// Service constructors (in dependency order)
	svcConstructors   []ServiceConstructor
    // Currently running services
	runningSvcs   map[reflect.Type]Service
	// network layer
	peerServer    *p2p.PeerServer
	/// api layer
	rpcServer     *rpc.RpcServer
}

func NewNode(conf *config.Config, db database.DB, chainParams *params.Params) (*Node,error) {
    return &Node{
		params: chainParams,
		quit:   make(chan struct{}),
	},nil
}

func (n *Node) Stop() error {
	log.Info("Stopping Server")

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

	log.Info("Starting server")

	// Initialize every service by calling the registered service constructors & save to services
	services := make(map[reflect.Type]Service)
	for _, c := range n.svcConstructors {
		ctx := &ServiceContext{}
		// Construct and save the service
		service, err := c(ctx)
		if err != nil {
			return err
		}
		kind := reflect.TypeOf(service)
		if _, exists := services[kind]; exists {
			return fmt.Errorf("Duplicate Service, Kind=%s}",kind)
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
	}

	// Finished node start
	n.runningSvcs = services
	// Server startup time. Used for the uptime command for uptime calculation.
	n.startupTime = time.Now().Unix()
	n.wg.Wrap(n.nodeEventHandler)
	return nil
}


func (n *Node) Register(sc ServiceConstructor) error {
	n.lock.Lock()
	defer n.lock.Unlock()
	// Already started?
	if atomic.AddInt32(&n.started, 1) != 1 {
		return fmt.Errorf("node has been started")
	}
	n.svcConstructors = append(n.svcConstructors, sc)
	return nil
}
