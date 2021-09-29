// Copyright (c) 2017-2018 The qitmeer developers

package blkmgr

import (
	"container/list"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/event"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/node/notify"
	"github.com/Qitmeer/qitmeer/p2p"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/services/common/progresslog"
	"github.com/Qitmeer/qitmeer/services/zmq"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// maxStallDuration is the time after which we will disconnect our
	// current sync peer if we haven't made progress.
	MaxStallDuration = 3 * time.Minute

	// stallSampleInterval the interval at which we will check to see if our
	// sync has stalled.
	StallSampleInterval = 3 * time.Second

	// maxStallDuration is the time after which we will disconnect our
	// current sync peer if we haven't made progress.
	MaxBlockStallDuration = 3 * time.Second
)

// BlockManager provides a concurrency safe block manager for handling all
// incoming blocks.
type BlockManager struct {
	started  int32
	shutdown int32

	config *config.Config
	params *params.Params

	notify notify.Notify

	chain *blockchain.BlockChain

	progressLogger *progresslog.BlockProgressLogger

	msgChan chan interface{}

	wg   sync.WaitGroup
	quit chan struct{}

	// The following fields are used for headers-first mode.
	headersFirstMode bool
	headerList       *list.List
	startHeader      *list.Element
	nextCheckpoint   *params.Checkpoint

	//block template cache
	cachedCurrentTemplate *types.BlockTemplate
	cachedParentTemplate  *types.BlockTemplate

	lastProgressTime time.Time

	// zmq notification
	zmqNotify zmq.IZMQNotification

	sync.Mutex

	//tx manager
	txManager TxManager

	// network server
	peerServer *p2p.Service
}

// NewBlockManager returns a new block manager.
// Use Start to begin processing asynchronous block and inv updates.
func NewBlockManager(ntmgr notify.Notify, indexManager blockchain.IndexManager, db database.DB,
	timeSource blockchain.MedianTimeSource, sigCache *txscript.SigCache,
	cfg *config.Config, par *params.Params,
	interrupt <-chan struct{}, events *event.Feed, peerServer *p2p.Service) (*BlockManager, error) {
	bm := BlockManager{
		config:         cfg,
		params:         par,
		notify:         ntmgr,
		progressLogger: progresslog.NewBlockProgressLogger("Processed", log),
		msgChan:        make(chan interface{}, cfg.MaxPeers*3),
		headerList:     list.New(),
		quit:           make(chan struct{}),
		peerServer:     peerServer,
	}

	// Create a new block chain instance with the appropriate configuration.
	var err error
	bm.chain, err = blockchain.New(&blockchain.Config{
		DB:             db,
		Interrupt:      interrupt,
		ChainParams:    par,
		TimeSource:     timeSource,
		Events:         events,
		SigCache:       sigCache,
		IndexManager:   indexManager,
		DAGType:        cfg.DAGType,
		CacheInvalidTx: cfg.CacheInvalidTx,
	})
	if err != nil {
		return nil, err
	}

	best := bm.chain.BestSnapshot()
	bm.chain.DisableCheckpoints(cfg.DisableCheckpoints)
	if !cfg.DisableCheckpoints {
		// Initialize the next checkpoint based on the current height.
		bm.nextCheckpoint = bm.findNextHeaderCheckpoint(uint64(best.GraphState.GetMainHeight()))
		if bm.nextCheckpoint != nil {
			bm.resetHeaderState(&best.Hash, uint64(best.GraphState.GetMainHeight()))
		}
	} else {
		log.Info("Checkpoints are disabled")
	}

	if cfg.DumpBlockchain != "" {
		err = bm.chain.DumpBlockChain(cfg.DumpBlockchain, par, uint64(best.GraphState.GetTotal())-1)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("closing after dumping blockchain")
	}

	bm.zmqNotify = zmq.NewZMQNotification(cfg)

	bm.subscribe(events)
	return &bm, nil
}

// handleNotifyMsg handles notifications from blockchain.  It does things such
// as request orphan block parents and relay accepted blocks to connected peers.
func (b *BlockManager) handleNotifyMsg(notification *blockchain.Notification) {
	switch notification.Type {
	// A block has been accepted into the block chain.  Relay it to other peers
	// and possibly notify RPC clients with the winning tickets.
	case blockchain.BlockAccepted:
		band, ok := notification.Data.(*blockchain.BlockAcceptedNotifyData)
		if !ok {
			log.Warn("Chain accepted notification is not " +
				"BlockAcceptedNotifyData.")
			break
		}
		block := band.Block
		if band.Flags&blockchain.BFP2PAdd == blockchain.BFP2PAdd {
			b.progressLogger.LogBlockHeight(block)
			// reset last progress time
			b.lastProgressTime = roughtime.Now()
		}
		b.zmqNotify.BlockAccepted(block)
		// Don't relay if we are not current. Other peers that are current
		// should already know about it
		if !b.peerServer.PeerSync().IsCurrent() {
			log.Trace("we are not current")
			return
		}
		log.Trace("we are current, can do relay")

		// Send a winning tickets notification as needed.  The notification will
		// only be sent when the following conditions hold:
		//
		// - The RPC server is running
		// - The block that would build on this one is at or after the height
		//   voting begins
		// - The block that would build on this one would not cause a reorg
		//   larger than the max reorg notify depth
		// - This block is after the final checkpoint height
		// - A notification for this block has not already been sent
		//
		// To help visualize the math here, consider the following two competing
		// branches:
		//
		// 100 -> 101  -> 102  -> 103 -> 104 -> 105 -> 106
		//    \-> 101' -> 102'
		//
		// Further, assume that this is a notification for block 103', or in
		// other words, it is extending the shorter side chain.  The reorg depth
		// would be 106 - (103 - 3) = 6.  This should intuitively make sense,
		// because if the side chain were to be extended enough to become the
		// best chain, it would result in a a reorg that would remove 6 blocks,
		// namely blocks 101, 102, 103, 104, 105, and 106.
		b.notify.RelayInventory(block.Block().Header, nil)

	// A block has been connected to the main block chain.
	case blockchain.BlockConnected:
		log.Trace("Chain connected notification.")
		blockSlice, ok := notification.Data.([]*types.SerializedBlock)
		if !ok {
			log.Warn("Chain connected notification is not a block slice.")
			break
		}

		if len(blockSlice) != 1 {
			log.Warn("Chain connected notification is wrong size slice.")
			break
		}

		block := blockSlice[0]
		// Remove all of the transactions (except the coinbase) in the
		// connected block from the transaction pool.  Secondly, remove any
		// transactions which are now double spends as a result of these
		// new transactions.  Finally, remove any transaction that is
		// no longer an orphan. Transactions which depend on a confirmed
		// transaction are NOT removed recursively because they are still
		// valid.
		for _, tx := range block.Transactions()[1:] {
			b.GetTxManager().MemPool().RemoveTransaction(tx, false)
			b.GetTxManager().MemPool().RemoveDoubleSpends(tx)
			b.GetTxManager().MemPool().RemoveOrphan(tx.Hash())
			b.notify.TransactionConfirmed(tx)
			acceptedTxs := b.GetTxManager().MemPool().ProcessOrphans(tx.Hash())
			b.notify.AnnounceNewTransactions(acceptedTxs, nil)
		}

		/*
			if r := b.server.rpcServer; r != nil {
				// Notify registered websocket clients of incoming block.
				r.ntfnMgr.NotifyBlockConnected(block)
			}
		*/

		b.zmqNotify.BlockConnected(block)

	// A block has been disconnected from the main block chain.
	case blockchain.BlockDisconnected:
		log.Trace("Chain disconnected notification.")
		block, ok := notification.Data.(*types.SerializedBlock)
		if !ok {
			log.Warn("Chain disconnected notification is not a block slice.")
			break
		}
		b.zmqNotify.BlockDisconnected(block)
	// The blockchain is reorganizing.
	case blockchain.Reorganization:
		log.Trace("Chain reorganization notification")
		/*
			rd, ok := notification.Data.(*blockchain.ReorganizationNotifyData)
			if !ok {
				log.Warn("Chain reorganization notification is malformed")
				break
			}

			// Notify registered websocket clients.
			if r := b.server.rpcServer; r != nil {
				r.ntfnMgr.NotifyReorganization(rd)
			}

			// Drop the associated mining template from the old chain, since it
			// will be no longer valid.
			b.cachedCurrentTemplate = nil
		*/
	}
}

func (b *BlockManager) IsCurrent() bool {
	return b.peerServer.PeerSync().IsCurrent()
}

// Start begins the core block handler which processes block and inv messages.
func (b *BlockManager) Start() {
	// Already started?
	if atomic.AddInt32(&b.started, 1) != 1 {
		return
	}

	log.Trace("Starting block manager")
	b.wg.Add(1)
	go b.blockHandler()
}

func (b *BlockManager) Stop() error {
	if atomic.AddInt32(&b.shutdown, 1) != 1 {
		log.Warn("Block manager is already in the process of " +
			"shutting down")
		return nil
	}
	log.Info("Block manager shutting down")
	close(b.quit)

	// shutdown zmq
	b.zmqNotify.Shutdown()
	return nil
}

func (b *BlockManager) WaitForStop() {
	log.Info("Wait For Block manager stop ...")
	b.wg.Wait()
	log.Info("Block manager stopped")
}

// findNextHeaderCheckpoint returns the next checkpoint after the passed layer.
// It returns nil when there is not one either because the height is already
// later than the final checkpoint or some other reason such as disabled
// checkpoints.
func (b *BlockManager) findNextHeaderCheckpoint(layer uint64) *params.Checkpoint {
	// There is no next checkpoint if checkpoints are disabled or there are
	// none for this current network.
	if b.config.DisableCheckpoints {
		return nil
	}
	checkpoints := b.params.Checkpoints
	if len(checkpoints) == 0 {
		return nil
	}

	// There is no next checkpoint if the height is already after the final
	// checkpoint.
	finalCheckpoint := &checkpoints[len(checkpoints)-1]
	if layer >= finalCheckpoint.Layer {
		return nil
	}

	// Find the next checkpoint.
	nextCheckpoint := finalCheckpoint
	for i := len(checkpoints) - 2; i >= 0; i-- {
		if layer >= checkpoints[i].Layer {
			break
		}
		nextCheckpoint = &checkpoints[i]
	}
	return nextCheckpoint
}

// resetHeaderState sets the headers-first mode state to values appropriate for
// syncing from a new peer.
func (b *BlockManager) resetHeaderState(newestHash *hash.Hash, newestHeight uint64) {
	b.headersFirstMode = false
	b.headerList.Init()
	b.startHeader = nil

	// When there is a next checkpoint, add an entry for the latest known
	// block into the header pool.  This allows the next downloaded header
	// to prove it links to the chain properly.
	if b.nextCheckpoint != nil {
		node := headerNode{height: newestHeight, hash: newestHash}
		b.headerList.PushBack(&node)
	}
}

func (b *BlockManager) blockHandler() {
	stallTicker := time.NewTicker(StallSampleInterval)
	defer stallTicker.Stop()

out:
	for {
		select {
		case m := <-b.msgChan:
			log.Trace("blkmgr msgChan received ...", "msg", m)
			switch msg := m.(type) {

			case tipGenerationMsg:
				log.Trace("blkmgr msgChan tipGenerationMsg", "msg", msg)
				g, err := b.chain.TipGeneration()
				msg.reply <- tipGenerationResponse{
					hashes: g,
					err:    err,
				}

			case processBlockMsg:
				log.Trace("blkmgr msgChan processBlockMsg", "msg", msg)

				if msg.flags.Has(blockchain.BFRPCAdd) {
					_, ok := b.chain.BlockDAG().CheckSubMainChainTip(msg.block.Block().Parents)
					if !ok {
						msg.reply <- processBlockResponse{
							isOrphan: false,
							err:      fmt.Errorf("The tips of block is expired:%s\n", msg.block.Hash().String()),
						}
						continue
					}
				}

				isOrphan, err := b.chain.ProcessBlock(
					msg.block, msg.flags)
				if err != nil {
					msg.reply <- processBlockResponse{
						isOrphan: isOrphan,
						err:      err,
					}
					continue
				}

				// If the block added to the dag chain, then we need to
				// update the tip locally on block manager.
				if !isOrphan {
					// TODO, decoupling mempool with bm
					b.GetTxManager().MemPool().PruneExpiredTx()
				}

				// Allow any clients performing long polling via the
				// getblocktemplate RPC to be notified when the new block causes
				// their old block template to become stale.
				// TODO, re-impl the client notify by subscript/publish
				/*
					rpcServer := b.rpcServer
					if rpcServer != nil {
						rpcServer.gbtWorkState.NotifyBlockConnected(msg.block.Hash())
					}
				*/

				msg.reply <- processBlockResponse{
					isOrphan: isOrphan,
					err:      nil,
				}
				b.peerServer.Rebroadcast().RegainMempool()

			case processTransactionMsg:
				log.Trace("blkmgr msgChan processTransactionMsg", "msg", msg)
				acceptedTxs, err := b.GetTxManager().MemPool().ProcessTransaction(msg.tx,
					msg.allowOrphans, msg.rateLimit, msg.allowHighFees)
				msg.reply <- processTransactionResponse{
					acceptedTxs: acceptedTxs,
					err:         err,
				}
			case isCurrentMsg:
				log.Trace("blkmgr msgChan isCurrentMsg", "msg", msg)
				msg.isCurrentReply <- b.IsCurrent()
				/*
					case pauseMsg:
						// Wait until the sender unpauses the manager.
						<-msg.unpause
				*/
			case getCurrentTemplateMsg:
				log.Trace("blkmgr msgChan getCurrentTemplateMsg", "msg", msg)
				cur := deepCopyBlockTemplate(b.cachedCurrentTemplate)
				msg.reply <- getCurrentTemplateResponse{
					Template: cur,
				}

			case setCurrentTemplateMsg:
				log.Trace("blkmgr msgChan setCurrentTemplateMsg", "msg", msg)
				b.cachedCurrentTemplate = deepCopyBlockTemplate(msg.Template)
				msg.reply <- setCurrentTemplateResponse{}

			case getParentTemplateMsg:
				log.Trace("blkmgr msgChan getParentTemplateMsg", "msg", msg)
				par := deepCopyBlockTemplate(b.cachedParentTemplate)
				msg.reply <- getParentTemplateResponse{
					Template: par,
				}

			case setParentTemplateMsg:
				log.Trace("blkmgr msgChan setParentTemplateMsg", "msg", msg)
				b.cachedParentTemplate = deepCopyBlockTemplate(msg.Template)
				msg.reply <- setParentTemplateResponse{}
			default:
				log.Error("Unknown message type", "msg", msg)
			}

		case <-stallTicker.C:
			b.handleStallSample()

		case <-b.quit:
			log.Trace("blkmgr quit received, break out")
			break out
		}
	}
	b.wg.Done()
	log.Trace("Block handler done")
}

// processBlockResponse is a response sent to the reply channel of a
// processBlockMsg.
type processBlockResponse struct {
	isOrphan bool
	err      error
}

// processBlockMsg is a message type to be sent across the message channel
// for requested a block is processed.  Note this call differs from blockMsg
// above in that blockMsg is intended for blocks that came from peers and have
// extra handling whereas this message essentially is just a concurrent safe
// way to call ProcessBlock on the internal block chain instance.
type processBlockMsg struct {
	block *types.SerializedBlock
	flags blockchain.BehaviorFlags
	reply chan processBlockResponse
}

// ProcessBlock makes use of ProcessBlock on an internal instance of a block
// chain.  It is funneled through the block manager since blockchain is not safe
// for concurrent access.
func (b *BlockManager) ProcessBlock(block *types.SerializedBlock, flags blockchain.BehaviorFlags) (bool, error) {
	reply := make(chan processBlockResponse, 1)
	b.msgChan <- processBlockMsg{block: block, flags: flags, reply: reply}
	response := <-reply
	return response.isOrphan, response.err
}

// processTransactionResponse is a response sent to the reply channel of a
// processTransactionMsg.
type processTransactionResponse struct {
	acceptedTxs []*types.TxDesc
	err         error
}

// processTransactionMsg is a message type to be sent across the message
// channel for requesting a transaction to be processed through the block
// manager.
type processTransactionMsg struct {
	tx            *types.Tx
	allowOrphans  bool
	rateLimit     bool
	allowHighFees bool
	reply         chan processTransactionResponse
}

// ProcessTransaction makes use of ProcessTransaction on an internal instance of
// a block chain.  It is funneled through the block manager since blockchain is
// not safe for concurrent access.
func (b *BlockManager) ProcessTransaction(tx *types.Tx, allowOrphans bool,
	rateLimit bool, allowHighFees bool) ([]*types.TxDesc, error) {
	reply := make(chan processTransactionResponse, 1)
	b.msgChan <- processTransactionMsg{tx, allowOrphans, rateLimit,
		allowHighFees, reply}
	response := <-reply
	return response.acceptedTxs, response.err
}

// isCurrentMsg is a message type to be sent across the message channel for
// requesting whether or not the block manager believes it is synced with
// the currently connected peers.
type isCurrentMsg struct {
	isCurrentReply chan bool
}

// IsCurrent returns whether or not the block manager believes it is synced with
// the connected peers.
func (b *BlockManager) Current() bool {
	reply := make(chan bool)
	log.Trace("send isCurrentMsg to blkmgr msgChan")
	b.msgChan <- isCurrentMsg{isCurrentReply: reply}
	return <-reply
}

// tipGenerationResponse is a response sent to the reply channel of a
// tipGenerationMsg query.
type tipGenerationResponse struct {
	hashes []hash.Hash
	err    error
}

// tipGenerationMsg is a message type to be sent across the message
// channel for requesting the required the entire generation of a
// block node.
type tipGenerationMsg struct {
	reply chan tipGenerationResponse
}

// TipGeneration returns the hashes of all the children of the current best
// chain tip.  It is funneled through the block manager since blockchain is not
// safe for concurrent access.
func (b *BlockManager) TipGeneration() ([]hash.Hash, error) {
	reply := make(chan tipGenerationResponse)
	b.msgChan <- tipGenerationMsg{reply: reply}
	response := <-reply
	return response.hashes, response.err
}

// handleStallSample will switch to a new sync peer if the current one has
// stalled. This is detected when by comparing the last progress timestamp with
// the current time, and disconnecting the peer if we stalled before reaching
// their highest advertised block.
func (b *BlockManager) handleStallSample() {
	if atomic.LoadInt32(&b.shutdown) != 0 {
		return
	}
}

// Return chain params
func (b *BlockManager) ChainParams() *params.Params {
	return b.params
}

// DAGSync
func (b *BlockManager) DAGSync() *blockdag.DAGSync {
	return nil
}

func (b *BlockManager) SetTxManager(txManager TxManager) {
	b.txManager = txManager
}

func (b *BlockManager) GetTxManager() TxManager {
	return b.txManager
}

func (b *BlockManager) subscribe(events *event.Feed) {
	ch := make(chan *event.Event)
	sub := events.Subscribe(ch)
	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case ev := <-ch:
				if ev.Data != nil {
					switch value := ev.Data.(type) {
					case *blockchain.Notification:
						b.handleNotifyMsg(value)
					}
				}
				if ev.Ack != nil {
					ev.Ack <- struct{}{}
				}
			case <-b.quit:
				log.Info("Close BlockManager Event Subscribe")
				return
			}
		}
	}()
}

// headerNode is used as a node in a list of headers that are linked together
// between checkpoints.
type headerNode struct {
	height uint64
	hash   *hash.Hash
}
