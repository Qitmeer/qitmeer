package miner

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/services/mining"
	"sync"
	"sync/atomic"
	"time"
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
)

var (
	// defaultNumWorkers is the default number of workers to use for mining
	// and is based on the number of processor cores.  This helps ensure the
	// system stays reasonably responsive under heavy load.
	defaultNumWorkers = uint32(params.CPUMinerThreads) //TODO, move to config
)

type CPUWorker struct {
	started  int32
	shutdown int32
	wg       sync.WaitGroup
	quit     chan struct{}

	discrete      bool
	discreteNum   int
	discreteBlock chan *hash.Hash

	updateHashes      chan uint64
	queryHashesPerSec chan float64
	workWg            sync.WaitGroup
	updateNumWorks    chan struct{}
	numWorks          uint32
	updateWork        chan struct{}
	hasNewWork        bool

	miner *Miner

	sync.Mutex
}

func (w *CPUWorker) GetType() string {
	return CPUWorkerType
}

func (w *CPUWorker) Start() error {
	err := w.miner.initCoinbase()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	// Already started?
	if atomic.AddInt32(&w.started, 1) != 1 {
		return nil
	}

	log.Info("Start CPU Worker...")

	w.miner.updateBlockTemplate(false)

	w.wg.Add(2)
	go w.speedMonitor()
	go w.workController()

	return nil
}

func (w *CPUWorker) Stop() {
	if atomic.AddInt32(&w.shutdown, 1) != 1 {
		log.Warn(fmt.Sprintf("CPU Worker is already in the process of shutting down"))
		return
	}
	log.Info("Stop CPU Worker...")

	close(w.quit)
	w.wg.Wait()
}

// speedMonitor handles tracking the number of hashes per second the mining
// process is performing.  It must be run as a goroutine.
func (w *CPUWorker) speedMonitor() {
	log.Trace("CPU Worker speed monitor started")

	var hashesPerSec float64
	var totalHashes uint64
	ticker := time.NewTicker(time.Second * hpsUpdateSecs)
	defer ticker.Stop()

out:
	for {
		select {
		// Periodic updates from the workers with how many hashes they
		// have performed.
		case numHashes := <-w.updateHashes:
			if w.discrete {
				continue
			}
			totalHashes += numHashes

		// Time to update the hashes per second.
		case <-ticker.C:
			if w.discrete {
				continue
			}
			curHashesPerSec := float64(totalHashes) / hpsUpdateSecs
			if hashesPerSec == 0 {
				hashesPerSec = curHashesPerSec
			}
			hashesPerSec = (hashesPerSec + curHashesPerSec) / 2
			totalHashes = 0
			if hashesPerSec != 0 {
				log.Debug(fmt.Sprintf("Hash speed: %6.0f kilohashes/s",
					hashesPerSec/1000))
			}

		// Request for the number of hashes per second.
		case w.queryHashesPerSec <- hashesPerSec:
			// Nothing to do.

		case <-w.quit:
			break out
		}
	}

	w.wg.Done()
	log.Trace("CPU Worker speed monitor done")
}

// HashesPerSecond returns the number of hashes per second the mining process
// is performing.  0 is returned if the miner is not currently running.
//
// This function is safe for concurrent access.
func (w *CPUWorker) HashesPerSecond() float64 {
	w.Lock()
	defer w.Unlock()

	// Nothing to do if the miner is not currently running.
	if atomic.LoadInt32(&w.started) == 0 {
		return 0
	}

	return <-w.queryHashesPerSec
}

// controller launches the worker goroutines that are used to
// generate block templates and solve them.  It also provides the ability to
// dynamically adjust the number of running worker goroutines.
//
// It must be run as a goroutine.
func (w *CPUWorker) workController() {
	// launchWorkers groups common code to launch a specified number of
	// workers for generating blocks.
	var runningWorks []chan struct{}
	launchWorkers := func(numWorkers uint32) {
		for i := uint32(0); i < numWorkers; i++ {
			quit := make(chan struct{})
			runningWorks = append(runningWorks, quit)

			w.workWg.Add(1)
			go w.generateBlocks()
		}
	}

	// Launch the current number of workers by default.
	runningWorks = make([]chan struct{}, 0, w.numWorks)
	launchWorkers(w.numWorks)

out:
	for {
		select {
		// Update the number of running workers.
		case <-w.updateNumWorks:
			// No change.
			numRunning := uint32(len(runningWorks))
			if w.numWorks == numRunning {
				continue
			}

			// Add new workers.
			if w.numWorks > numRunning {
				launchWorkers(w.numWorks - numRunning)
				continue
			}

			// Signal the most recently created goroutines to exit.
			for i := numRunning - 1; i >= w.numWorks; i-- {
				close(runningWorks[i])
				runningWorks[i] = nil
				runningWorks = runningWorks[:i]
			}

		case <-w.quit:
			for _, quit := range runningWorks {
				close(quit)
			}
			break out
		}
	}

	// Wait until all workers shut down to stop the speed monitor since
	// they rely on being able to send updates to it.
	w.workWg.Wait()
	w.wg.Done()
}

// SetNumWorkers sets the number of workers to create which solve blocks.  Any
// negative values will cause a default number of workers to be used which is
// based on the number of processor cores in the system.  A value of 0 will
// cause all CPU mining to be stopped.
//
// This function is safe for concurrent access.
func (w *CPUWorker) SetNumWorkers(numWorkers int32) {
	if numWorkers <= 0 {
		return
	}

	// Don't lock until after the first check since Stop does its own
	// locking.
	w.Lock()
	defer w.Unlock()

	// Use default if provided value is negative.
	if numWorkers <= 0 {
		w.numWorks = defaultNumWorkers
	} else {
		w.numWorks = uint32(numWorkers)
	}

	// When the miner is already running, notify the controller about the
	// the change.

	if atomic.LoadInt32(&w.started) != 0 {
		w.updateNumWorks <- struct{}{}
	}
}

// NumWorkers returns the number of workers which are running to solve blocks.
//
// This function is safe for concurrent access.
func (w *CPUWorker) NumWorkers() int32 {
	w.Lock()
	defer w.Unlock()

	return int32(w.numWorks)
}

// IsRunning returns whether or not the CPU miner has been started and is
// therefore currenting mining.
//
// This function is safe for concurrent access.
func (w *CPUWorker) IsRunning() bool {
	return atomic.LoadInt32(&w.started) != 0
}

func (w *CPUWorker) Update() {
	if atomic.LoadInt32(&w.shutdown) != 0 {
		return
	}
	w.Lock()
	defer w.Unlock()

	if w.discrete && w.discreteNum <= 0 {
		return
	}
	w.hasNewWork = true
	w.updateWork <- struct{}{}
	w.hasNewWork = false
}

func (w *CPUWorker) generateDiscrete(num int, block chan *hash.Hash) {
	if atomic.LoadInt32(&w.shutdown) != 0 {
		if block != nil {
			close(block)
		}
		return
	}
	w.Lock()
	defer w.Unlock()
	if w.discrete && w.discreteNum > 0 {
		if block != nil {
			close(block)
		}
		log.Info(fmt.Sprintf("It already exists generate blocks by discrete: left=%d", w.discreteNum))
		return
	}
	w.discrete = true
	w.discreteNum = num
	w.discreteBlock = block
}

func (w *CPUWorker) generateBlocks() {
	log.Trace(fmt.Sprintf("Starting generate blocks worker:%s", w.GetType()))
out:
	for {
		// Quit when the miner is stopped.
		select {
		case <-w.updateWork:
			if w.discrete && w.discreteNum <= 0 {
				continue
			}
			if w.solveBlock() {
				block := types.NewBlock(w.miner.template.Block)
				block.SetHeight(uint(w.miner.template.Height))
				info, err := w.miner.submitBlock(block)
				if err != nil {
					log.Error(fmt.Sprintf("Failed to submit new block:%s ,%v", block.Hash().String(), err))
					w.cleanDiscrete()
					continue
				}
				log.Info(fmt.Sprintf("%v", info))

				if w.discrete && w.discreteNum > 0 {
					if w.discreteBlock != nil {
						w.discreteBlock <- block.Hash()
					}
					w.discreteNum--
					if w.discreteNum <= 0 {
						w.cleanDiscrete()
					}
				}
			} else {
				w.cleanDiscrete()
			}
		case <-w.quit:
			break out
		}
	}

	w.workWg.Done()
	log.Trace(fmt.Sprintf("Generate blocks worker done:%s", w.GetType()))
}

func (w *CPUWorker) cleanDiscrete() {
	if w.discrete {
		w.discreteNum = 0
		if w.discreteBlock != nil {
			close(w.discreteBlock)
			w.discreteBlock = nil
		}
	}
}

func (w *CPUWorker) solveBlock() bool {
	if w.miner.template == nil {
		return false
	}
	// Start a ticker which is used to signal checks for stale work and
	// updates to the speed monitor.
	ticker := time.NewTicker(333 * time.Millisecond)
	defer ticker.Stop()

	// Create a couple of convenience variables.
	block := w.miner.template.Block
	header := &block.Header

	// Initial state.
	lastGenerated := roughtime.Now()
	hashesCompleted := uint64(0)
	// TODO, decided if need extra nonce for coinbase-tx
	// Note that the entire extra nonce range is iterated and the offset is
	// added relying on the fact that overflow will wrap around 0 as
	// provided by the Go spec.
	// for extraNonce := uint64(0); extraNonce < maxExtraNonce; extraNonce++ {

	// Update the extra nonce in the block template with the
	// new value by regenerating the coinbase script and
	// setting the merkle root to the new value.
	// TODO, decided if need extra nonce for coinbase-tx
	// updateExtraNonce(msgBlock, extraNonce+enOffset)

	// Update the extra nonce in the block template header with the
	// new value.
	// binary.LittleEndian.PutUint64(header.ExtraData[:], extraNonce+enOffset)

	// Search through the entire nonce range for a solution while
	// periodically checking for early quit and stale block
	// conditions along with updates to the speed monitor.
	for i := uint64(0); i <= maxNonce; i++ {
		select {
		case <-w.quit:
			return false

		case <-ticker.C:
			w.updateHashes <- hashesCompleted
			hashesCompleted = 0

			// The current block is stale if the memory pool
			// has been updated since the block template was
			// generated and it has been at least 3 seconds,
			// or if it's been one minute.
			if w.hasNewWork || roughtime.Now().After(lastGenerated.Add(gbtRegenerateSeconds*time.Second)) {
				return false
			}

			err := mining.UpdateBlockTime(block, w.miner.blockManager.GetChain(), w.miner.timeSource, params.ActiveNetParams.Params)
			if err != nil {
				log.Warn(fmt.Sprintf("CPU miner unable to update block template time: %v", err))
				return false
			}

		default:
			// Non-blocking select to fall through
		}
		instance := pow.GetInstance(w.miner.powType, 0, []byte{})
		instance.SetNonce(uint64(i))
		instance.SetMainHeight(pow.MainHeight(w.miner.template.Height))
		instance.SetParams(params.ActiveNetParams.Params.PowConfig)
		hashesCompleted += 2
		header.Pow = instance
		if header.Pow.FindSolver(header.BlockData(), header.BlockHash(), header.Difficulty) {
			w.updateHashes <- hashesCompleted
			return true
		}
		// Each hash is actually a double hash (tow hashes), so
	}
	return false
}

func NewCPUWorker(miner *Miner) *CPUWorker {
	w := CPUWorker{
		quit:              make(chan struct{}),
		discrete:          false,
		discreteNum:       0,
		miner:             miner,
		updateHashes:      make(chan uint64),
		queryHashesPerSec: make(chan float64),
		updateNumWorks:    make(chan struct{}),
		numWorks:          defaultNumWorkers,
		updateWork:        make(chan struct{}),
	}

	return &w
}
