// Copyright (c) 2017-2018 The nox developers
package node

import (
	"github.com/HalalChain/qitmeer/config"
	"github.com/HalalChain/qitmeer/database"
	"github.com/HalalChain/qitmeer/rpc"
	"github.com/HalalChain/qitmeer/p2p/peerserver"
)

// NoxLight implements the nox light node service.
type NoxLight struct {
	// database
	db               database.DB
	config           *config.Config
}

func (light *NoxLight) Start(server *peerserver.PeerServer) error {
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

func newNoxLight(n *Node) (*NoxLight, error){
	light := NoxLight{
		config : n.Config,
		db : n.DB,
	}
	return &light, nil
}
