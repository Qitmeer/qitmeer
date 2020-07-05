/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:Service.go
 * Date:7/2/20 8:04 PM
 * Author:Jin
 * Email:lochjin@gmail.com
 */
package p2p

import (
	"context"
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/message"
	"github.com/Qitmeer/qitmeer/services/blkmgr"
	"github.com/Qitmeer/qitmeer/services/mempool"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"time"
)

type Service struct {
	started      bool
	isPreGenesis bool
	pingMethod   func(ctx context.Context, id peer.ID) error
	cancel       context.CancelFunc
	//peers                 *peers.Status
	//privKey               *ecdsa.PrivateKey
	startupErr            error
	ctx                   context.Context
	host                  host.Host
	genesisTime           time.Time
	genesisValidatorsRoot []byte

	TimeSource   blockchain.MedianTimeSource
	BlockManager *blkmgr.BlockManager
	TxMemPool    *mempool.TxPool
}

func (s *Service) Start() error {
	log.Info("P2P Service Start")
	return nil
}

func (s *Service) Stop() error {
	log.Info("P2P Service Stop")
	return nil
}

func (s *Service) ConnectedCount() int32 {
	return 0
}

// ConnectedPeers returns an array consisting of all connected peers.
func (s *Service) ConnectedPeers() []int {
	return nil
}

func (s *Service) GetBanlist() map[string]time.Time {
	return nil
}

func (s *Service) RemoveBan(host string) {

}

func (s *Service) RelayInventory(invVect *message.InvVect, data interface{}) {

}

func (s *Service) BroadcastMessage(msg message.Message) {

}

func NewService(cfg *config.Config) (*Service, error) {
	s := Service{}

	return &s, nil
}
