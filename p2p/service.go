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
	"crypto/ecdsa"
	"fmt"
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/message"
	"github.com/Qitmeer/qitmeer/services/blkmgr"
	"github.com/Qitmeer/qitmeer/services/mempool"
	"github.com/dgraph-io/ristretto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"time"
)

type Service struct {
	cfg           *Config
	ctx           context.Context
	cancel        context.CancelFunc
	exclusionList *ristretto.Cache
	started       bool
	isPreGenesis  bool
	privKey       *ecdsa.PrivateKey

	pingMethod func(ctx context.Context, id peer.ID) error
	//peers                 *peers.Status

	startupErr            error
	host                  host.Host
	genesisTime           time.Time
	genesisValidatorsRoot []byte

	TimeSource   blockchain.MedianTimeSource
	BlockManager *blkmgr.BlockManager
	TxMemPool    *mempool.TxPool
}

func (s *Service) Start() error {
	if s.started {
		return fmt.Errorf("Attempted to start p2p service when it was already started")
	}
	log.Info("P2P Service Start")

	s.isPreGenesis = false

	s.started = true

	return nil
}

// Started returns true if the p2p service has successfully started.
func (s *Service) Started() bool {
	return s.started
}

func (s *Service) Stop() error {
	log.Info("P2P Service Stop")

	defer s.cancel()
	s.started = false

	return nil
}

// Status of the p2p service. Will return an error if the service is considered unhealthy to
// indicate that this node should not serve traffic until the issue has been resolved.
func (s *Service) Status() error {
	if s.isPreGenesis {
		return nil
	}
	if !s.started {
		return errors.New("not running")
	}
	if s.startupErr != nil {
		return s.startupErr
	}
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
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,
		MaxCost:     1000,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}

	bootnodesTemp := cfg.BootstrapNodes
	bootnodeAddrs := make([]string, 0) //dest of final list of nodes
	for _, addr := range bootnodesTemp {
		if filepath.Ext(addr) == ".yaml" {
			fileNodes, err := readbootNodes(addr)
			if err != nil {
				return nil, err
			}
			bootnodeAddrs = append(bootnodeAddrs, fileNodes...)
		} else {
			bootnodeAddrs = append(bootnodeAddrs, addr)
		}
	}

	s := Service{
		cfg: &Config{
			StaticPeers:          cfg.AddPeers,
			DataDir:              cfg.DataDir,
			BootstrapNodeAddr:    bootnodeAddrs,
			NoDiscovery:          cfg.NoDiscovery,
			MaxPeers:             uint(cfg.MaxPeers),
			ReadWritePermissions: 0600, //-rw------- Read and Write permissions for user
		},
		ctx:           ctx,
		cancel:        cancel,
		exclusionList: cache,
		isPreGenesis:  true,
	}

	dv5Nodes := parseBootStrapAddrs(s.cfg.BootstrapNodeAddr)
	s.cfg.Discv5BootStrapAddr = dv5Nodes

	ipAddr := ipAddr()
	s.privKey, err = privKey(s.cfg)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to generate p2p private key:%v", err))
		return nil, err
	}
	s.metaData, err = metaDataFromConfig(s.cfg)
	if err != nil {
		log.WithError(err).Error("Failed to create peer metadata")
		return nil, err
	}
	s.addrFilter, err = configureFilter(s.cfg)
	if err != nil {
		log.WithError(err).Error("Failed to create address filter")
		return nil, err
	}
	opts := s.buildOptions(ipAddr, s.privKey)
	h, err := libp2p.New(s.ctx, opts...)
	if err != nil {
		log.WithError(err).Error("Failed to create p2p host")
		return nil, err
	}

	s.host = h
	return &s, nil
}

func readbootNodes(fileName string) ([]string, error) {
	fileContent, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	listNodes := make([]string, 0)
	err = yaml.Unmarshal(fileContent, &listNodes)
	if err != nil {
		return nil, err
	}
	return listNodes, nil
}
