package p2p

import (
	"github.com/noxproject/nox/params"
	"github.com/noxproject/nox/database"
	"github.com/noxproject/nox/config"
	"github.com/noxproject/nox/log"
)

// Use start to begin accepting connections from peers.
// peer server handling communications to and from nox peers.
type PeerServer struct{
    // address manager caching the peers
	addrManager          *AddrManager
	// conn manager handles network connections.
	connManager          *ConnManager
}

func NewPeerServer(cfg *config.Config, db database.DB, chainParams *params.Params) (*PeerServer, error){
	s := PeerServer{}
	return &s, nil
}

func (p *PeerServer) Start() error {
	log.Debug("Starting P2P server")
	return nil
}
func (p *PeerServer) Stop() error {
	log.Debug("Stopping P2P server")
	return nil
}

type AddrManager struct {}
type ConnManager struct {}
