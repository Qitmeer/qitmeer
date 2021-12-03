// Copyright (c) 2017-2018 The qitmeer developers
package node

import (
	"github.com/Qitmeer/qng-core/config"
	"github.com/Qitmeer/qng-core/database"
	"github.com/Qitmeer/qitmeer/node/service"
)

// QitmeerLight implements the qitmeer light node service.
type QitmeerLight struct {
	service.Service
	// database
	db     database.DB
	config *config.Config
}

func newQitmeerLight(n *Node) (*QitmeerLight, error) {
	light := QitmeerLight{
		config: n.Config,
		db:     n.DB,
	}
	return &light, nil
}
