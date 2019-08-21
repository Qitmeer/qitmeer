// Copyright (c) 2017-2018 The qitmeer developers
package node

import (
	"github.com/Qitmeer/qitmeer-lib/config"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer-lib/rpc"
	"github.com/Qitmeer/qitmeer/p2p/peerserver"
)

// QitmeerLight implements the qitmeer light node service.
type QitmeerLight struct {
	// database
	db               database.DB
	config           *config.Config
}

func (light *QitmeerLight) Start(server *peerserver.PeerServer) error {
	log.Debug("Starting Qitmeer light node service")
	return nil
}

func (light *QitmeerLight) Stop() error {
	log.Debug("Stopping Qitmeer light node service")
	return nil
}

func (light *QitmeerLight)	APIs() []rpc.API {
	return []rpc.API{}
}

func newQitmeerLight(n *Node) (*QitmeerLight, error){
	light := QitmeerLight{
		config : n.Config,
		db : n.DB,
	}
	return &light, nil
}
