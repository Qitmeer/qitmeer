package miner

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/event"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/rpc"
	"github.com/Qitmeer/qitmeer/services/blkmgr"
	"github.com/Qitmeer/qitmeer/services/mining"
	"sync"
	"sync/atomic"
	"time"
)

// Miner creates blocks and searches for proof-of-work values.
type Miner struct {
	started  int32
	shutdown int32
	msgChan  chan interface{}
	wg       sync.WaitGroup
	quit     chan struct{}

	cfg          *config.Config
	events       *event.Feed
	txSource     mining.TxSource
	timeSource   blockchain.MedianTimeSource
	blockManager *blkmgr.BlockManager
	policy       *mining.Policy
	sigCache     *txscript.SigCache
}

func (m *Miner) Start() error {
	if !m.cfg.Miner {
		return nil
	}
	// Already started?
	if atomic.AddInt32(&m.started, 1) != 1 {
		return nil
	}

	log.Info("Start Miner...")

	m.wg.Add(1)
	go m.handler()
	return nil
}

func (m *Miner) Stop() {
	if !m.cfg.Miner {
		return
	}
	if atomic.AddInt32(&m.shutdown, 1) != 1 {
		log.Warn(fmt.Sprintf("Miner is already in the process of shutting down"))
		return
	}
	log.Info("Stop Miner...")

	close(m.quit)
	m.wg.Wait()
}

func (m *Miner) handler() {
	stallTicker := time.NewTicker(params.ActiveNetParams.TargetTimePerBlock)
	defer stallTicker.Stop()

out:
	for {
		select {
		case m := <-m.msgChan:
			switch msg := m.(type) {

			default:
				log.Warn("Invalid message type in task handler: %T", msg)
			}

		case <-stallTicker.C:
			m.handleStallSample()

		case <-m.quit:
			break out
		}
	}

cleanup:
	for {
		select {
		case <-m.msgChan:
		default:
			break cleanup
		}
	}

	m.wg.Done()
	log.Trace("Miner handler done")
}

func (m *Miner) handleStallSample() {
	if atomic.LoadInt32(&m.shutdown) != 0 {
		return
	}
	log.Debug("Miner stall sample")
}

func NewMiner(cfg *config.Config, policy *mining.Policy,
	sigCache *txscript.SigCache,
	txSource mining.TxSource, tsource blockchain.MedianTimeSource, blkMgr *blkmgr.BlockManager, numWorkers uint32) *Miner {
	m := Miner{
		msgChan:      make(chan interface{}),
		quit:         make(chan struct{}),
		cfg:          cfg,
		policy:       policy,
		sigCache:     sigCache,
		txSource:     txSource,
		timeSource:   tsource,
		blockManager: blkMgr,
	}

	return &m
}

func (m *Miner) APIs() []rpc.API {
	return []rpc.API{}
}
