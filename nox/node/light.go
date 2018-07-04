// Copyright (c) 2017-2018 The nox developers
package node

import (
	"github.com/noxproject/nox/config"
	"github.com/noxproject/nox/database"
	"github.com/noxproject/nox/p2p"
	"github.com/noxproject/nox/rpc"
	"github.com/noxproject/nox/log"
)

// NoxLight implements the nox light node service.
type NoxLight struct {
	// database
	db               database.DB
	config           *config.Config
}

func (light *NoxLight) Start(server *p2p.PeerServer) error {
	log.Debug("Starting Nox light node service")
	return nil
}

func (light *NoxLight) Stop() error {
	log.Debug("Stopping Nox light node service")
	return nil
}

func (light *NoxLight)	APIs() []rpc.API {
	return []rpc.API{}
}

// register NoxLight service to node
func registerNoxLight(n *Node) error{
	// register acctmgr
	err := n.register(NewServiceConstructor("Nox-light",
		func(ctx *ServiceContext) (Service, error) {
		noxlight, err := newNoxLight(n.config,n.db)
		return noxlight, err
	}))
	return err
}


func newNoxLight(cfg *config.Config,db database.DB) (*NoxLight, error){
	light := NoxLight{
		config : cfg,
		db : db,
	}
	return &light, nil
}
