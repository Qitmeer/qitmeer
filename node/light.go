// Copyright (c) 2017-2018 The qitmeer developers
package node

import (
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer/rpc"
)

// QitmeerLight implements the qitmeer light node service.
type QitmeerLight struct {
	// database
	db     database.DB
	config *config.Config
}

func (light *QitmeerLight) Start() error {
	log.Debug("Starting Qitmeer light node service")
	return nil
}

func (light *QitmeerLight) Stop() error {
	log.Debug("Stopping Qitmeer light node service")
	return nil
}

func (light *QitmeerLight) APIs() []rpc.API {
	return []rpc.API{}
}

func newQitmeerLight(n *Node) (*QitmeerLight, error) {
	light := QitmeerLight{
		config: n.Config,
		db:     n.DB,
	}
	return &light, nil
}
