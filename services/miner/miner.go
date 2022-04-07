package miner

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/event"
	"github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/rpc"
	"github.com/Qitmeer/qitmeer/services/blkmgr"
	"github.com/Qitmeer/qitmeer/services/mempool"
	"github.com/Qitmeer/qitmeer/services/mining"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// gbtRegenerateSeconds is the number of seconds that must pass before
	// a new template is generated when the previous block hash has not
	// changed and there have been changes to the available transactions
	// in the memory pool.
	gbtRegenerateSeconds = 60
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
	worker       IWorker

	template        *types.BlockTemplate
	lastTxUpdate    time.Time
	lastTemplate    time.Time
	minTimestamp    time.Time
	coinbaseAddress types.Address
	powType         pow.PowType

	sync.Mutex
	submitLocker sync.Mutex

	totalSubmit   int
	successSubmit int
}

func (m *Miner) Start() error {
	if !m.cfg.Miner {
		return nil
	}
	// Already started?
	if atomic.AddInt32(&m.started, 1) != 1 {
		return nil
	}

	//
	log.Info("Start Miner...")

	m.subscribe()

	m.wg.Add(1)
	go m.handler()

	if m.cfg.Generate {
		m.StartCPUMining()
	}
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
		case mc := <-m.msgChan:
			switch msg := mc.(type) {
			case *StartCPUMiningMsg:
				if m.worker != nil {
					if m.worker.GetType() == CPUWorkerType {
						continue
					}
					m.worker.Stop()
					m.worker = nil
				}
				m.worker = NewCPUWorker(m)
				if m.worker.Start() != nil {
					m.worker = nil
					continue
				}
				m.worker.Update()

			case *CPUMiningGenerateMsg:
				if msg.discreteNum <= 0 {
					if msg.block != nil {
						close(msg.block)
					}
					continue
				}
				if m.worker != nil {
					if m.worker.GetType() == CPUWorkerType {
						m.worker.(*CPUWorker).generateDiscrete(msg.discreteNum, msg.block)
						if m.powType != msg.powType {
							m.powType = msg.powType
						}
						if m.updateBlockTemplate(true) == nil {
							m.worker.Update()
						} else {
							if msg.block != nil {
								close(msg.block)
							}
						}
						continue
					}
					m.worker.Stop()
					m.worker = nil
				}
				worker := NewCPUWorker(m)
				m.worker = worker
				if m.worker.Start() != nil {
					m.worker = nil
					if msg.block != nil {
						close(msg.block)
					}
					continue
				}
				worker.generateDiscrete(msg.discreteNum, msg.block)
				worker.Update()

			case *BlockChainChangeMsg:
				if m.updateBlockTemplate(false) == nil {
					if m.worker != nil {
						m.worker.Update()
					}
				}
			case *MempoolChangeMsg:
				if m.updateBlockTemplate(false) == nil {
					if m.worker != nil {
						m.worker.Update()
					}
				}

			case *GBTMiningMsg:
				if m.worker != nil {
					if m.worker.GetType() == GBTWorkerType {
						m.worker.(*GBTWorker).GetRequest(msg.request, msg.reply)
						continue
					}
					m.worker.Stop()
					m.worker = nil
				}
				worker := NewGBTWorker(m)
				m.worker = worker
				err := m.worker.Start()
				if err != nil {
					log.Error(err.Error())
					m.worker = nil
					if msg.reply != nil {
						msg.reply <- &gbtResponse{nil, err}
					}
					continue
				}
				worker.Update()
				worker.GetRequest(msg.request, msg.reply)

			case *RemoteMiningMsg:
				if m.worker != nil {
					if m.worker.GetType() == RemoteWorkerType {
						m.worker.(*RemoteWorker).GetRequest(msg.powType, msg.reply)
						continue
					}
					m.worker.Stop()
					m.worker = nil
				}
				worker := NewRemoteWorker(m)
				m.worker = worker
				err := m.worker.Start()
				if err != nil {
					log.Error(err.Error())
					m.worker = nil
					if msg.reply != nil {
						msg.reply <- &gbtResponse{nil, err}
					}
					continue
				}
				worker.Update()
				worker.GetRequest(msg.powType, msg.reply)

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

	if m.worker != nil {
		m.worker.Stop()
	}

	m.wg.Done()
	log.Trace("Miner handler done")
}

func (m *Miner) updateBlockTemplate(force bool) error {

	reCreate := false
	//
	if force {
		reCreate = true
	} else if m.template == nil {
		reCreate = true
	}
	if !reCreate {
		hasCoinbaseAddr := m.coinbaseAddress != nil
		if hasCoinbaseAddr != m.template.ValidPayAddress {
			reCreate = true
		}
	}
	if !reCreate {
		parentsSet := blockdag.NewHashSet()
		parentsSet.AddList(m.blockManager.GetChain().GetMiningTips(blockdag.MaxPriority))

		tparentSet := blockdag.NewHashSet()
		tparentSet.AddList(m.template.Block.Parents)
		if !parentsSet.IsEqual(tparentSet) {
			reCreate = true
		} else {
			lastTxUpdate := m.txSource.LastUpdated()
			if lastTxUpdate.IsZero() {
				lastTxUpdate = roughtime.Now()
			}
			if lastTxUpdate != m.lastTxUpdate && roughtime.Now().After(m.lastTemplate.Add(time.Second*gbtRegenerateSeconds)) {
				reCreate = true
			}
		}
	}

	if reCreate {
		template, err := mining.NewBlockTemplate(m.policy, params.ActiveNetParams.Params, m.sigCache, m.txSource, m.timeSource, m.blockManager, m.coinbaseAddress, nil, m.powType)
		if err != nil {
			e := fmt.Errorf("Failed to create new block template: %s", err.Error())
			log.Error(e.Error())
			return e
		}
		m.template = template
		m.lastTxUpdate = m.txSource.LastUpdated()
		m.lastTemplate = time.Now()

		// Get the minimum allowed timestamp for the block based on the
		// median timestamp of the last several blocks per the chain
		// consensus rules.
		m.minTimestamp = mining.MinimumMedianTime(m.blockManager.GetChain())

		return nil
	} else {
		err := mining.UpdateBlockTime(m.template.Block, m.blockManager.GetChain(), m.timeSource, params.ActiveNetParams.Params)
		if err != nil {
			log.Warn(fmt.Sprintf("%s unable to update block template time: %v", m.worker.GetType(), err))
			return err
		}
	}
	return nil
}

func (m *Miner) subscribe() {
	ch := make(chan *event.Event)
	sub := m.events.Subscribe(ch)
	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case ev := <-ch:
				if ev.Data != nil {
					switch value := ev.Data.(type) {
					case *blockchain.Notification:
						m.handleNotifyMsg(value)
					case int:
						if value == mempool.MempoolTxAdd {
							go m.MempoolChange()
						}
					}
				}
				if ev.Ack != nil {
					ev.Ack <- struct{}{}
				}
			case <-m.quit:
				log.Info("Close Miner Event Subscribe")
				return
			}
		}
	}()
}
func (m *Miner) handleNotifyMsg(notification *blockchain.Notification) {
	if m.worker == nil {
		return
	}
	switch notification.Type {
	case blockchain.BlockAccepted:
		band, ok := notification.Data.(*blockchain.BlockAcceptedNotifyData)
		if !ok {
			return
		}
		if band.IsMainChainTipChange {
			go m.BlockChainChange()
		}
	}
}

// submitBlock submits the passed block to network after ensuring it passes all
// of the consensus validation rules.
func (m *Miner) submitBlock(block *types.SerializedBlock) (interface{}, error) {
	if m.worker == nil {
		return nil, fmt.Errorf("You must enable miner by --miner.")
	}
	m.submitLocker.Lock()
	defer m.submitLocker.Unlock()
	m.totalSubmit++

	// Process this block using the same rules as blocks coming from other
	// nodes. This will in turn relay it to the network like normal.
	isOrphan, err := m.blockManager.ProcessBlock(block, blockchain.BFRPCAdd)
	if err != nil {
		// Anything other than a rule violation is an unexpected error,
		// so log that error as an internal error.
		rErr, ok := err.(blockchain.RuleError)
		if !ok {
			return nil, fmt.Errorf(fmt.Sprintf("Unexpected error while processing block submitted miner: %v (%s)", err, m.worker.GetType()))
		}
		// Occasionally errors are given out for timing errors with
		// ReduceMinDifficulty and high block works that is above
		// the target. Feed these to debug.
		if params.ActiveNetParams.Params.ReduceMinDifficulty &&
			rErr.ErrorCode == blockchain.ErrHighHash {
			return nil, fmt.Errorf(fmt.Sprintf("Block submitted via miner rejected "+
				"because of ReduceMinDifficulty time sync failure: %v (%s)",
				err, m.worker.GetType()))
		}
		// Other rule errors should be reported.
		return nil, fmt.Errorf(fmt.Sprintf("Block submitted via %s rejected: %v ", m.worker.GetType(), err))
	}
	if isOrphan {
		return nil, fmt.Errorf(fmt.Sprintf("Block submitted via %s is an orphan building "+
			"on parent %v", m.worker.GetType(), block.Block().Header.ParentRoot))
	}

	m.successSubmit++

	// The block was accepted.
	coinbaseTxOuts := block.Block().Transactions[0].TxOut
	coinbaseTxGenerated := uint64(0)
	for _, out := range coinbaseTxOuts {
		coinbaseTxGenerated += uint64(out.Amount.Value)
	}
	return fmt.Sprintf("Block submitted accepted hash:%s order:%s height:%d amount:%d miner:%s", block.Hash(),
		blockdag.GetOrderLogStr(uint(block.Order())), block.Height(), coinbaseTxGenerated, m.worker.GetType()), nil
}

func (m *Miner) submitBlockHeader(header *types.BlockHeader) (interface{}, error) {
	if !m.IsEnable() || m.template == nil {
		return nil, fmt.Errorf("You must enable miner by --miner.")
	}
	tHeader := &m.template.Block.Header
	if !IsEqualForMiner(tHeader, header) {
		return nil, fmt.Errorf("You're overdue")
	}
	tHeader.Difficulty = header.Difficulty
	tHeader.Timestamp = header.Timestamp
	tHeader.Pow = header.Pow
	block := types.NewBlock(m.template.Block)
	block.SetHeight(uint(m.template.Height))
	return m.submitBlock(block)
}

func (m *Miner) CanMining() error {
	currentOrder := m.blockManager.GetChain().BestSnapshot().GraphState.GetTotal() - 1
	if currentOrder != 0 && !m.blockManager.IsCurrent() {
		log.Trace("Client in initial download, qitmeer is downloading blocks...")
		return rpc.RPCClientInInitialDownloadError("Client in initial download ",
			"qitmeer is downloading blocks...")
	}
	return nil
}

func (m *Miner) IsEnable() bool {
	if !m.cfg.Miner {
		return false
	}
	if atomic.LoadInt32(&m.shutdown) != 0 {
		return false
	}
	if atomic.LoadInt32(&m.started) == 0 {
		return false
	}
	return true
}

func (m *Miner) initCoinbase() error {
	if m.coinbaseAddress != nil {
		return nil
	}
	mAddrs := m.cfg.GetMinningAddrs()
	if len(mAddrs) <= 0 {
		// Respond with an error if there are no addresses to pay the
		// created blocks to.
		return fmt.Errorf("No payment addresses specified via --miningaddr.")
	}
	// Choose a payment address at random.
	if len(mAddrs) == 1 {
		m.coinbaseAddress = mAddrs[0]
	} else {
		m.coinbaseAddress = mAddrs[rand.Intn(len(mAddrs))]
	}
	log.Info(fmt.Sprintf("Init Coinbase Address:%s", m.coinbaseAddress.String()))
	return nil
}

func (m *Miner) handleStallSample() {
	//if atomic.LoadInt32(&m.shutdown) != 0 {
	//	return
	//}
	//log.Debug("Miner stall sample")
}

func (m *Miner) StartCPUMining() {
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&m.shutdown) != 0 {
		return
	}

	m.msgChan <- &StartCPUMiningMsg{}
}

func (m *Miner) CPUMiningGenerate(discreteNum int, block chan *hash.Hash, powType pow.PowType) error {
	if err := m.CanMining(); err != nil {
		return err
	}
	if atomic.LoadInt32(&m.started) == 0 {
		if !m.cfg.Miner {
			m.cfg.Miner = true
		}
		if err := m.Start(); err != nil {
			log.Error(err.Error())
			return err
		}
	}
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&m.shutdown) != 0 {
		return fmt.Errorf("Miner is quit")
	}
	m.msgChan <- &CPUMiningGenerateMsg{discreteNum: discreteNum, block: block, powType: powType}
	return nil
}

func (m *Miner) BlockChainChange() {
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&m.shutdown) != 0 {
		return
	}
	if err := m.CanMining(); err != nil {
		return
	}

	m.msgChan <- &BlockChainChangeMsg{}
}

func (m *Miner) MempoolChange() {
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&m.shutdown) != 0 {
		return
	}
	if m.worker == nil {
		return
	}
	if err := m.CanMining(); err != nil {
		return
	}
	m.msgChan <- &MempoolChangeMsg{}
}

func (m *Miner) GBTMining(request *json.TemplateRequest, reply chan *gbtResponse) error {
	if atomic.LoadInt32(&m.started) == 0 {
		if !m.cfg.Miner {
			m.cfg.Miner = true
		}
		if err := m.Start(); err != nil {
			log.Error(err.Error())
			return err
		}
	}
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&m.shutdown) != 0 {
		return fmt.Errorf("Miner is shutdown")
	}
	if err := m.CanMining(); err != nil {
		return err
	}

	m.msgChan <- &GBTMiningMsg{request: request, reply: reply}
	return nil
}

func (m *Miner) RemoteMining(powType pow.PowType, reply chan *gbtResponse) error {
	if !m.cfg.Miner {
		return fmt.Errorf("Miner is disable. You can enable by --miner.")
	}
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&m.shutdown) != 0 {
		return fmt.Errorf("Miner is shutdown")
	}
	if err := m.CanMining(); err != nil {
		return err
	}

	m.msgChan <- &RemoteMiningMsg{powType: powType, reply: reply}
	return nil
}

func NewMiner(cfg *config.Config, policy *mining.Policy,
	sigCache *txscript.SigCache,
	txSource mining.TxSource, tsource blockchain.MedianTimeSource, blkMgr *blkmgr.BlockManager, events *event.Feed) *Miner {
	m := Miner{
		msgChan:      make(chan interface{}),
		quit:         make(chan struct{}),
		cfg:          cfg,
		policy:       policy,
		sigCache:     sigCache,
		txSource:     txSource,
		timeSource:   tsource,
		blockManager: blkMgr,
		powType:      pow.MEERXKECCAKV1,
		events:       events,
	}

	return &m
}

func IsEqualForMiner(header *types.BlockHeader, other *types.BlockHeader) bool {
	if header.Version != other.Version ||
		!header.ParentRoot.IsEqual(&other.ParentRoot) ||
		!header.StateRoot.IsEqual(&other.StateRoot) ||
		!header.TxRoot.IsEqual(&other.TxRoot) {
		return false
	}
	return true
}
