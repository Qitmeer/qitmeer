// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2014-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package miner

import (
	"sync"
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/config"
	"math/rand"
	"time"
	"fmt"
	"errors"
	"github.com/noxproject/nox/params"
	"github.com/noxproject/nox/log"
	"github.com/noxproject/nox/services/blkmgr"
	"github.com/noxproject/nox/core/blockchain"
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/engine/txscript"
	"github.com/noxproject/nox/services/mining"
)

const (
	// maxNonce is the maximum value a nonce can be in a block header.
	maxNonce = ^uint64(0) // 2^64 - 1

	// TODO, decided if th extra nonce for coinbase-tx need
	// maxExtraNonce is the maximum value an extra nonce used in a coinbase
	// transaction can be.
	maxExtraNonce = ^uint64(0) // 2^64 - 1

	// hpsUpdateSecs is the number of seconds to wait in between each
	// update to the hashes per second monitor.
	hpsUpdateSecs = 10

	// hashUpdateSec is the number of seconds each worker waits in between
	// notifying the speed monitor with how many hashes have been completed
	// while they are actively searching for a solution.  This is done to
	// reduce the amount of syncs between the workers that must be done to
	// keep track of the hashes per second.
	hashUpdateSecs = 15

	// maxSimnetToMine is the maximum number of blocks to mine on HEAD~1
	// for simnet so that you don't run out of memory if tickets for
	// some reason run out during simulations.
	maxSimnetToMine uint8 = 4
)


// CPUMiner provides facilities for solving blocks (mining) using the CPU in
// a concurrency-safe manner.  It consists of two main goroutines -- a speed
// monitor and a controller for worker goroutines which generate and solve
// blocks.  The number of goroutines can be set via the SetMaxGoRoutines
// function, but the default is based on the number of processor cores in the
// system which is typically sufficient.
type CPUMiner struct {
	sync.Mutex
	params            *params.Params
	config            *config.Config
	policy            *mining.Policy
	sigCache          *txscript.SigCache
	txSource          mining.TxSource
	timeSource        blockchain.MedianTimeSource
	blockManager      *blkmgr.BlockManager
	numWorkers        uint32
	started           bool
	discreteMining    bool
	submitBlockLock   sync.Mutex
	wg                sync.WaitGroup
	workerWg          sync.WaitGroup
	updateNumWorkers  chan struct{}
	queryHashesPerSec chan float64
	updateHashes      chan uint64
	speedMonitorQuit  chan struct{}
	quit              chan struct{}

	// This is a map that keeps track of how many blocks have
	// been mined on each parent by the CPUMiner. It is only
	// for use in simulation networks, to diminish memory
	// exhaustion. It should not race because it's only
	// accessed in a single threaded loop below.
	minedOnParents map[hash.Hash]uint8
}

// newCPUMiner returns a new instance of a CPU miner for the provided server.
// Use Start to begin the mining process.  See the documentation for CPUMiner
// type for more details.
func NewCPUMiner(cfg *config.Config,par *params.Params, policy *mining.Policy,
	cache *txscript.SigCache,
	source mining.TxSource,tsource blockchain.MedianTimeSource,blkMgr *blkmgr.BlockManager,   numWorkers uint32) *CPUMiner {
	return &CPUMiner{
		config:            cfg,
		params:            par,
		policy:            policy,
		sigCache:          cache,
		txSource:          source,
		timeSource:        tsource,
		blockManager:      blkMgr,
		numWorkers:        numWorkers,
		updateNumWorkers:  make(chan struct{}),
		queryHashesPerSec: make(chan float64),
		updateHashes:      make(chan uint64),
		minedOnParents:    make(map[hash.Hash]uint8),
	}
}

// GenerateNBlocks generates the requested number of blocks. It is self
// contained in that it creates block templates and attempts to solve them while
// detecting when it is performing stale work and reacting accordingly by
// generating a new block template.  When a block is solved, it is submitted.
// The function returns a list of the hashes of generated blocks.
func (m *CPUMiner) GenerateNBlocks(n uint32) ([]*hash.Hash, error) {
	m.Lock()

	// Respond with an error if there's virtually 0 chance of CPU-mining a block.
	if !m.params.GenerateSupported {
		m.Unlock()
		return nil, errors.New("no support for `generate` on the current " +
			"network, " + m.params.Net.String() +
			", as it's unlikely to be possible to CPU-mine a block.")
	}

	// Respond with an error if server is already mining.
	if m.started || m.discreteMining {
		m.Unlock()
		return nil, errors.New("server is already CPU mining. Please call " +
			"`setgenerate 0` before calling discrete `generate` commands.")
	}

	m.started = true
	m.discreteMining = true

	m.speedMonitorQuit = make(chan struct{})
	m.wg.Add(1)
	go m.speedMonitor()

	m.Unlock()

	log.Trace("Generating blocks", n)

	i := uint32(0)
	blockHashes := make([]*hash.Hash, n)

	// Start a ticker which is used to signal checks for stale work and
	// updates to the speed monitor.
	ticker := time.NewTicker(time.Second * hashUpdateSecs)
	defer ticker.Stop()

	for {
		// Read updateNumWorkers in case someone tries a `setgenerate` while
		// we're generating. We can ignore it as the `generate` RPC call only
		// uses 1 worker.
		select {
		case <-m.updateNumWorkers:
		default:
		}

		// Grab the lock used for block submission, since the current block will
		// be changing and this would otherwise end up building a new block
		// template on a block that is in the process of becoming stale.
		m.submitBlockLock.Lock()

		// Choose a payment address at random.
		rand.Seed(time.Now().UnixNano())
		payToAddr := m.config.GetMinningAddrs()[rand.Intn(len(m.config.GetMinningAddrs()))]

		// Create a new block template using the available transactions
		// in the memory pool as a source of transactions to potentially
		// include in the block.
		// TODO, refactor NewBlockTemplate input dependencies
		template, err := mining.NewBlockTemplate(m.policy,m.config,m.params,m.sigCache,m.txSource,m.timeSource,m.blockManager,payToAddr)
		m.submitBlockLock.Unlock()
		if err != nil {
			errStr := fmt.Sprintf("Failed to create new block "+
				"template: %v", err)
			log.Error(errStr)
			continue
		}
		if template == nil {
			errStr := fmt.Sprintf("Not enough voters on parent block " +
				"and failed to pull parent template")
			log.Debug(errStr)
			continue
		}

		// Attempt to solve the block.  The function will exit early
		// with false when conditions that trigger a stale block, so
		// a new block template can be generated.  When the return is
		// true a solution was found, so submit the solved block.
		if m.solveBlock(template.Block, ticker, nil) {
			block := types.NewBlock(template.Block)
			m.submitBlock(block)
			blockHashes[i] = block.Hash()
			i++
			if i == n {
				log.Trace("Generated %d blocks", i)
				m.Lock()
				close(m.speedMonitorQuit)
				m.wg.Wait()
				m.started = false
				m.discreteMining = false
				m.Unlock()
				return blockHashes, nil
			}
		}
	}
}

// speedMonitor handles tracking the number of hashes per second the mining
// process is performing.  It must be run as a goroutine.
func (m *CPUMiner) speedMonitor() {
	log.Trace("CPU miner speed monitor started")

	var hashesPerSec float64
	var totalHashes uint64
	ticker := time.NewTicker(time.Second * hpsUpdateSecs)
	defer ticker.Stop()

out:
	for {
		select {
		// Periodic updates from the workers with how many hashes they
		// have performed.
		case numHashes := <-m.updateHashes:
			totalHashes += numHashes

		// Time to update the hashes per second.
		case <-ticker.C:
			curHashesPerSec := float64(totalHashes) / hpsUpdateSecs
			if hashesPerSec == 0 {
				hashesPerSec = curHashesPerSec
			}
			hashesPerSec = (hashesPerSec + curHashesPerSec) / 2
			totalHashes = 0
			if hashesPerSec != 0 {
				log.Debug("Hash speed: %6.0f kilohashes/s",
					hashesPerSec/1000)
			}

		// Request for the number of hashes per second.
		case m.queryHashesPerSec <- hashesPerSec:
			// Nothing to do.

		case <-m.speedMonitorQuit:
			break out
		}
	}

	m.wg.Done()
	log.Trace("CPU miner speed monitor done")
}

// solveBlock attempts to find some combination of a nonce, extra nonce, and
// current timestamp which makes the passed block hash to a value less than the
// target difficulty.  The timestamp is updated periodically and the passed
// block is modified with all tweaks during this process.  This means that
// when the function returns true, the block is ready for submission.
//
// This function will return early with false when conditions that trigger a
// stale block such as a new block showing up or periodically when there are
// new transactions and enough time has elapsed without finding a solution.
func (m *CPUMiner) solveBlock(msgBlock *types.Block, ticker *time.Ticker, quit chan struct{}) bool {

	// TODO, decided if need extra nonce for coinbase-tx
	// Choose a random extra nonce offset for this block template and
	// worker.
	/*
	enOffset, err := s.RandomUint64()
	if err != nil {
		log.Error("Unexpected error while generating random "+
			"extra nonce offset: %v", err)
		enOffset = 0
	}
	*/

	// Create a couple of convenience variables.
	header := &msgBlock.Header
	targetDifficulty := blockchain.CompactToBig(header.Difficulty)

	// Initial state.
	lastGenerated := time.Now()
	lastTxUpdate := m.txSource.LastUpdated()
	hashesCompleted := uint64(0)

		// Search through the entire nonce range for a solution while
		// periodically checking for early quit and stale block
		// conditions along with updates to the speed monitor.
		for i := uint64(0); i <= maxNonce; i++ {
			select {
			case <-quit:
				return false

			case <-ticker.C:
				m.updateHashes <- hashesCompleted
				hashesCompleted = 0

				// The current block is stale if the memory pool
				// has been updated since the block template was
				// generated and it has been at least 3 seconds,
				// or if it's been one minute.
				if (lastTxUpdate != m.txSource.LastUpdated() &&
					time.Now().After(lastGenerated.Add(3*time.Second))) ||
					time.Now().After(lastGenerated.Add(60*time.Second)) {

					return false
				}

				err := mining.UpdateBlockTime(msgBlock,m.blockManager, m.blockManager.GetChain(), m.timeSource, m.params, m.config)
				if err != nil {
					log.Warn("CPU miner unable to update block template "+
						"time: %v", err)
					return false
				}

			default:
				// Non-blocking select to fall through
			}

			// Update the nonce and hash the block header.
			header.Nonce = i
			hash := header.BlockHash()
			hashesCompleted++

			// The block is solved when the new block hash is less
			// than the target difficulty.  Yay!
			if blockchain.HashToBig(&hash).Cmp(targetDifficulty) <= 0 {
				m.updateHashes <- hashesCompleted
				return true
			}
		}

	return false
}

// submitBlock submits the passed block to network after ensuring it passes all
// of the consensus validation rules.
func (m *CPUMiner) submitBlock(block *types.SerializedBlock) bool {
	m.submitBlockLock.Lock()
	defer m.submitBlockLock.Unlock()

	// Process this block using the same rules as blocks coming from other
	// nodes. This will in turn relay it to the network like normal.
	isOrphan, err := m.blockManager.ProcessBlock(block, blockchain.BFNone)
	if err != nil {
		// Anything other than a rule violation is an unexpected error,
		// so log that error as an internal error.
		rErr, ok := err.(blockchain.RuleError)
		if !ok {
			log.Error("Unexpected error while processing "+
				"block submitted via CPU miner: %v", err)
			return false
		}
		// Occasionally errors are given out for timing errors with
		// ReduceMinDifficulty and high block works that is above
		// the target. Feed these to debug.
		if m.params.ReduceMinDifficulty &&
			rErr.ErrorCode == blockchain.ErrHighHash {
			log.Debug("Block submitted via CPU miner rejected "+
				"because of ReduceMinDifficulty time sync failure: %v",
				err)
			return false
		}
		// Other rule errors should be reported.
		log.Error("Block submitted via CPU miner rejected: %v", err)
		return false

	}
	if isOrphan {
		log.Error("Block submitted via CPU miner is an orphan building "+
			"on parent %v", block.Block().Header.ParentRoot)
		return false
	}

	// The block was accepted.
	coinbaseTxOuts := block.Block().Transactions[0].TxOut
	coinbaseTxGenerated := uint64(0)
	for _, out := range coinbaseTxOuts {
		coinbaseTxGenerated += out.Amount
	}
	log.Info("Block submitted via CPU miner accepted (hash %s, "+
		"height %v, amount %v)", block.Hash(), block.Height(),
		int64(coinbaseTxGenerated))
	return true
}

