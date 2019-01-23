// Copyright (c) 2017-2018 The nox developers

package blkmgr

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/config"
	"github.com/noxproject/nox/core/blockchain"
	"github.com/noxproject/nox/core/message"
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/database"
	"github.com/noxproject/nox/engine/txscript"
	"github.com/noxproject/nox/node/notify"
	"github.com/noxproject/nox/p2p/peer"
	"github.com/noxproject/nox/params"
	"github.com/noxproject/nox/params/dcr/types"
	"github.com/noxproject/nox/services/common/progresslog"
	"github.com/noxproject/nox/services/mempool"
	"sync"
	"sync/atomic"
)

// BlockManager provides a concurrency safe block manager for handling all
// incoming blocks.
type BlockManager struct {
	started             int32
	shutdown            int32

	config              *config.Config
	params              *params.Params

	notify              notify.Notify

	//TODO, decoupling mempool with bm
	txMemPool			*mempool.TxPool

	chain               *blockchain.BlockChain

	rejectedTxns        map[hash.Hash]struct{}
	requestedTxns       map[hash.Hash]struct{}
	requestedEverTxns   map[hash.Hash]uint8
	requestedBlocks     map[hash.Hash]struct{}
	requestedEverBlocks map[hash.Hash]uint8
	progressLogger      *progresslog.BlockProgressLogger

	syncPeer            *peer.ServerPeer
	msgChan             chan interface{}

	chainState          ChainState

	wg                  sync.WaitGroup
	quit                chan struct{}

	// The following fields are used for headers-first mode.
	headersFirstMode bool
	headerList       *list.List
	startHeader      *list.Element
	nextCheckpoint   *params.Checkpoint

	//block template cache
	cachedCurrentTemplate *types.BlockTemplate
	cachedParentTemplate  *types.BlockTemplate

	// The following fields are used to track the height being synced to from
	// peers.
	syncHeightMtx sync.Mutex
	syncHeight    uint64

}

// NewBlockManager returns a new block manager.
// Use Start to begin processing asynchronous block and inv updates.
func NewBlockManager(ntmgr notify.Notify,indexManager blockchain.IndexManager,db database.DB,
	timeSource blockchain.MedianTimeSource, sigCache *txscript.SigCache,
	cfg *config.Config, par *params.Params, /*server *peerserver.PeerServer,*/
	interrupt <-chan struct{}) (*BlockManager, error) {
	bm := BlockManager{
		config:              cfg,
		params:              par,
		notify:              ntmgr,
		rejectedTxns:        make(map[hash.Hash]struct{}),
		requestedTxns:       make(map[hash.Hash]struct{}),
		requestedEverTxns:   make(map[hash.Hash]uint8),
		requestedBlocks:     make(map[hash.Hash]struct{}),
		requestedEverBlocks: make(map[hash.Hash]uint8),
		progressLogger:      progresslog.NewBlockProgressLogger("Processed", log),
		msgChan:             make(chan interface{}, cfg.MaxPeers*3),
		headerList:          list.New(),
		quit:                make(chan struct{}),
	}

	// Create a new block chain instance with the appropriate configuration.
	var err error
	bm.chain, err = blockchain.New(&blockchain.Config{
		DB:            db,
		Interrupt:     interrupt,
		ChainParams:   par,
		TimeSource:    timeSource,
		Notifications: bm.handleNotifyMsg,
		SigCache:      sigCache,
		IndexManager:  indexManager,
	})
	if err != nil {
		return nil, err
	}
	best := bm.chain.BestSnapshot()
	bm.chain.DisableCheckpoints(cfg.DisableCheckpoints)
	if !cfg.DisableCheckpoints {
		// Initialize the next checkpoint based on the current height.
		bm.nextCheckpoint = bm.findNextHeaderCheckpoint(best.Height)
		if bm.nextCheckpoint != nil {
			bm.resetHeaderState(&best.Hash, best.Height)
		}
	} else {
		log.Info("Checkpoints are disabled")
	}

	if cfg.DumpBlockchain != "" {
		err = bm.chain.DumpBlockChain(cfg.DumpBlockchain, par, best.Height)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("closing after dumping blockchain")
	}

	// Retrieve the current previous block hash and next stake difficulty.
	curPrevHash := bm.chain.BestPrevHash()

	bm.GetChainState().UpdateChainState(&best.Hash,best.Height,best.MedianTime,curPrevHash)

	bm.syncHeightMtx.Lock()
	bm.syncHeight = best.Height
	bm.syncHeightMtx.Unlock()
	return &bm, nil
}

// Set the tx mem-pool to the block manager, It should call before the block manager
// getting started otherwise an error thrown
// TODO, decoupling mempool with bm
func (b *BlockManager) SetMemPool(pool *mempool.TxPool) error {
	if b.started == 1 {
		return errors.New("BlockManager already started, can't set mem pool")
	}
	b.txMemPool = pool
	return nil
}



// handleNotifyMsg handles notifications from blockchain.  It does things such
// as request orphan block parents and relay accepted blocks to connected peers.
func (b *BlockManager) handleNotifyMsg(notification *blockchain.Notification) {
	switch notification.Type {
	// A block has been accepted into the block chain.  Relay it to other peers
	// and possibly notify RPC clients with the winning tickets.
	case blockchain.BlockAccepted:
		// Don't relay if we are not current. Other peers that are current
		// should already know about it
		if !b.current() {
			return
		}

		band, ok := notification.Data.(*blockchain.BlockAcceptedNotifyData)
		if !ok {
			log.Warn("Chain accepted notification is not " +
				"BlockAcceptedNotifyData.")
			break
		}
		block := band.Block

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
		blockHash := block.Hash()

		// Generate the inventory vector and relay it.
		iv := message.NewInvVect(message.InvTypeBlock, blockHash)
		log.Info("relay inv","inv",iv)

		b.notify.RelayInventory(iv, block.Block().Header)

	// A block has been connected to the main block chain.
	case blockchain.BlockConnected:
		log.Trace("Chain connected notification.")
		blockSlice, ok := notification.Data.([]*types.SerializedBlock)
		if !ok {
			log.Warn("Chain connected notification is not a block slice.")
			break
		}

		if len(blockSlice) != 2 {
			log.Warn("Chain connected notification is wrong size slice.")
			break
		}

		block := blockSlice[0]
		parentBlock := blockSlice[1]

		// Remove the transaction and all transactions
		// that depend on it if it wasn't accepted into
		// the transaction pool. Probably this will mostly
		// throw errors, as the majority will already be
		// in the mempool.
		for _, tx := range parentBlock.Transactions()[1:] {
			_, err := b.txMemPool.MaybeAcceptTransaction(tx, false,
				true)
			if err != nil {

				b.txMemPool.RemoveTransaction(tx, true)
			}
		}
		// Remove all of the transactions (except the coinbase) in the
		// connected block from the transaction pool.  Secondly, remove any
		// transactions which are now double spends as a result of these
		// new transactions.  Finally, remove any transaction that is
		// no longer an orphan. Transactions which depend on a confirmed
		// transaction are NOT removed recursively because they are still
		// valid.
		for _, tx := range block.Transactions()[1:] {
			b.txMemPool.RemoveTransaction(tx, false)
			b.txMemPool.RemoveDoubleSpends(tx)
			b.txMemPool.RemoveOrphan(tx.Hash())
			acceptedTxs := b.txMemPool.ProcessOrphans(tx.Hash())
			b.notify.AnnounceNewTransactions(acceptedTxs)
		}

		/*
		if r := b.server.rpcServer; r != nil {
			// Notify registered websocket clients of incoming block.
			r.ntfnMgr.NotifyBlockConnected(block)
		}
		*/

	// A block has been disconnected from the main block chain.
	case blockchain.BlockDisconnected:
		log.Trace("Chain disconnected notification.")
		blockSlice, ok := notification.Data.([]*types.SerializedBlock)
		if !ok {
			log.Warn("Chain disconnected notification is not a block slice.")
			break
		}

		if len(blockSlice) != 2 {
			log.Warn("Chain disconnected notification is wrong size slice.")
			break
		}

		block := blockSlice[0]

		// Reinsert all of the transactions (except the coinbase) into
		// the transaction pool.
		for _, tx := range block.Transactions()[1:] {
			_, err := b.txMemPool.MaybeAcceptTransaction(tx,
				false, false)
			if err != nil {
				// Remove the transaction and all transactions
				// that depend on it if it wasn't accepted into
				// the transaction pool.
				b.txMemPool.RemoveTransaction(tx, true)
			}
		}

		/*
		// Notify registered websocket clients.
		if r := b.server.rpcServer; r != nil {
			r.ntfnMgr.NotifyBlockDisconnected(block)
		}
		*/

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

// current returns true if we believe we are synced with our peers, false if we
// still have blocks to check
func (b *BlockManager) current() bool {
	if !b.chain.IsCurrent() {
		return false
	}

	// if blockChain thinks we are current and we have no syncPeer it
	// is probably right.
	if b.syncPeer == nil {
		return true
	}

	// No matter what chain thinks, if we are below the block we are syncing
	// to we are not current.
	if b.chain.BestSnapshot().Height < b.syncPeer.LastBlock() {
		log.Trace("comparing the current best vs sync last",
			"current.best", b.chain.BestSnapshot().Height, "sync.last",b.syncPeer.LastBlock())
		return false
	}

	return true
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
	// drain the msg channel before send quit signal
	for len(b.msgChan) > 0 {
		log.Trace("Drain Block manager msgchan","msg", <-b.msgChan)
	}
	close(b.quit)
	return nil
}

func (b *BlockManager) WaitForStop() {
	log.Info("Wait For Block manager stop ...")
	b.wg.Wait()
	log.Info("Block manager stopped")
}

// limitMap is a helper function for maps that require a maximum limit by
// evicting a random transaction if adding a new value would cause it to
// overflow the maximum allowed.
func (b *BlockManager) limitMap(m map[hash.Hash]struct{}, limit int) {
	if len(m)+1 > limit {
		// Remove a random entry from the map.  For most compilers, Go's
		// range statement iterates starting at a random item although
		// that is not 100% guaranteed by the spec.  The iteration order
		// is not important here because an adversary would have to be
		// able to pull off preimage attacks on the hashing function in
		// order to target eviction of specific entries anyways.
		for txHash := range m {
			delete(m, txHash)
			return
		}
	}
}

// fetchHeaderBlocks creates and sends a request to the syncPeer for the next
// list of blocks to be downloaded based on the current list of headers.
func (b *BlockManager) fetchHeaderBlocks() {
	// Nothing to do if there is no start header.
	if b.startHeader == nil {
		log.Warn("fetchHeaderBlocks called with no start header")
		return
	}

	// Build up a getdata request for the list of blocks the headers
	// describe.  The size hint will be limited to wire.MaxInvPerMsg by
	// the function, so no need to double check it here.
	gdmsg := message.NewMsgGetDataSizeHint(uint(b.headerList.Len()))
	numRequested := 0
	for e := b.startHeader; e != nil; e = e.Next() {
		node, ok := e.Value.(*headerNode)
		if !ok {
			log.Warn("Header list node type is not a headerNode")
			continue
		}

		iv := message.NewInvVect(message.InvTypeBlock, node.hash)
		haveInv, err := b.haveInventory(iv)
		if err != nil {
			log.Warn("Unexpected failure when checking for "+
				"existing inventory during header block fetch",
				"error",err)
			continue
		}
		if !haveInv {
			b.requestedBlocks[*node.hash] = struct{}{}
			b.requestedEverBlocks[*node.hash] = 0
			b.syncPeer.RequestedBlocks[*node.hash] = struct{}{}
			err = gdmsg.AddInvVect(iv)
			if err != nil {
				log.Warn("Failed to add invvect while fetching block headers",
					"error",err)
			}
			numRequested++
		}
		b.startHeader = e.Next()
		if numRequested >= wire.MaxInvPerMsg {
			break
		}
	}
	if len(gdmsg.InvList) > 0 {
		b.syncPeer.QueueMessage(gdmsg, nil)
	}
}

// haveInventory returns whether or not the inventory represented by the passed
// inventory vector is known.  This includes checking all of the various places
// inventory can be when it is in different states such as blocks that are part
// of the main chain, on a side chain, in the orphan pool, and transactions that
// are in the memory pool (either the main pool or orphan pool).
func (b *BlockManager) haveInventory(invVect *message.InvVect) (bool, error) {
	switch invVect.Type {
	case message.InvTypeBlock:
		// Ask chain if the block is known to it in any form (main
		// chain, side chain, or orphan).
		return b.chain.HaveBlock(&invVect.Hash)

	case message.InvTypeTx:
		// Ask the transaction memory pool if the transaction is known
		// to it in any form (main pool or orphan).
		if b.txMemPool.HaveTransaction(&invVect.Hash) {
			return true, nil
		}

		// Check if the transaction exists from the point of view of the
		// end of the main chain.
		entry, err := b.chain.FetchUtxoEntry(&invVect.Hash)
		if err != nil {
			return false, err
		}
		return entry != nil && !entry.IsFullySpent(), nil
	}

	// The requested inventory is is an unsupported type, so just claim
	// it is known to avoid requesting it.
	return true, nil
}

// findNextHeaderCheckpoint returns the next checkpoint after the passed height.
// It returns nil when there is not one either because the height is already
// later than the final checkpoint or some other reason such as disabled
// checkpoints.
func (b *BlockManager) findNextHeaderCheckpoint(height uint64) *params.Checkpoint {
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
	if height >= finalCheckpoint.Height {
		return nil
	}

	// Find the next checkpoint.
	nextCheckpoint := finalCheckpoint
	for i := len(checkpoints) - 2; i >= 0; i-- {
		if height >= checkpoints[i].Height {
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
	candidatePeers := list.New()
out:
	for {
		select {
		case m := <-b.msgChan:
			log.Trace("blkmgr msgChan received ...", "msg", m)
			switch msg := m.(type) {
			case *newPeerMsg:
				log.Trace("blkmgr msgChan newPeer", "msg", msg)
				b.handleNewPeerMsg(candidatePeers, msg.peer)
			case *blockMsg:
				log.Trace("blkmgr msgChan blockMsg", "msg", msg)
				b.handleBlockMsg(msg)
				msg.peer.BlockProcessed <- struct{}{}
			case *invMsg:
				log.Trace("blkmgr msgChan invMsg", "msg", msg)
				b.handleInvMsg(msg)
			case *donePeerMsg:
				log.Trace("blkmgr msgChan donePeerMsg", "msg", msg)
				b.handleDonePeerMsg(candidatePeers, msg.peer)
			case getSyncPeerMsg:
				log.Trace("blkmgr msgChan getSyncPeerMsg", "msg", msg)
				msg.reply <- b.syncPeer
			case tipGenerationMsg:
				log.Trace("blkmgr msgChan tipGenerationMsg", "msg", msg)
				g, err := b.chain.TipGeneration()
				msg.reply <- tipGenerationResponse{
					hashes: g,
					err:    err,
				}
			case requestFromPeerMsg:
				err := b.requestFromPeer(msg.peer, msg.blocks)
				msg.reply <- requestFromPeerResponse{
					err: err,
				}

				/*
			case *txMsg:
				b.handleTxMsg(msg)
				msg.peer.txProcessed <- struct{}{}

			case *headersMsg:
				b.handleHeadersMsg(msg)



			case calcNextReqDiffNodeMsg:
				difficulty, err :=
					b.chain.CalcNextRequiredDiffFromNode(msg.hash,
						msg.timestamp)
				msg.reply <- calcNextReqDifficultyResponse{
					difficulty: difficulty,
					err:        err,
				}

			case calcNextReqStakeDifficultyMsg:
				stakeDiff, err := b.chain.CalcNextRequiredStakeDifficulty()
				msg.reply <- calcNextReqStakeDifficultyResponse{
					stakeDifficulty: stakeDiff,
					err:             err,
				}

			case forceReorganizationMsg:
				err := b.chain.ForceHeadReorganization(
					msg.formerBest, msg.newBest)

				// Reorganizing has succeeded, so we need to
				// update the chain state.
				if err == nil {
					// Query the db for the latest best block since
					// the block that was processed could be on a
					// side chain or have caused a reorg.
					best := b.chain.BestSnapshot()

					// Fetch the required lottery data.
					winningTickets, poolSize, finalState, err :=
						b.chain.LotteryDataForBlock(&best.Hash)

					// Update registered websocket clients on the
					// current stake difficulty.
					nextStakeDiff, errSDiff :=
						b.chain.CalcNextRequiredStakeDifficulty()
					if err != nil {
						bmgrLog.Warnf("Failed to get next stake difficulty "+
							"calculation: %v", err)
					}
					r := b.server.rpcServer
					if r != nil && errSDiff == nil {
						r.ntfnMgr.NotifyStakeDifficulty(
							&StakeDifficultyNtfnData{
								best.Hash,
								best.Height,
								nextStakeDiff,
							})
						b.server.txMemPool.PruneStakeTx(nextStakeDiff,
							best.Height)
						b.server.txMemPool.PruneExpiredTx(best.Height)
					}

					missedTickets, err := b.chain.MissedTickets()
					if err != nil {
						bmgrLog.Warnf("Failed to get missed tickets"+
							": %v", err)
					}

					// The blockchain should be updated, so fetch the
					// latest snapshot.
					best = b.chain.BestSnapshot()
					curPrevHash := b.chain.BestPrevHash()

					b.updateChainState(&best.Hash,
						best.Height,
						finalState,
						uint32(poolSize),
						nextStakeDiff,
						winningTickets,
						missedTickets,
						curPrevHash)
				}

				msg.reply <- forceReorganizationResponse{
					err: err,
				}


			*/

			case processBlockMsg:
				forkLen, isOrphan, err := b.chain.ProcessBlock(
					msg.block, msg.flags)
				if err != nil {
					msg.reply <- processBlockResponse{
						forkLen:  forkLen,
						isOrphan: isOrphan,
						err:      err,
					}
					continue
				}

				// If the block added to the main chain, then we need to
				// update the tip locally on block manager.
				onMainChain := !isOrphan && forkLen == 0
				if onMainChain {
					// Query the chain for the latest best block
					// since the block that was processed could be
					// on a side chain or have caused a reorg.
					best := b.chain.BestSnapshot()

					// TODO, decoupling mempool with bm
					b.txMemPool.PruneExpiredTx()

					curPrevHash := b.chain.BestPrevHash()

					b.GetChainState().UpdateChainState(&best.Hash,
						best.Height, best.MedianTime,
						curPrevHash)
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

			case processTransactionMsg:
				acceptedTxs, err := b.txMemPool.ProcessTransaction(msg.tx,
					msg.allowOrphans, msg.rateLimit, msg.allowHighFees)
				msg.reply <- processTransactionResponse{
					acceptedTxs: acceptedTxs,
					err:         err,
				}
			case isCurrentMsg:
				msg.reply <- b.current()
				/*
			case pauseMsg:
				// Wait until the sender unpauses the manager.
				<-msg.unpause
			*/
			case getCurrentTemplateMsg:
				cur := deepCopyBlockTemplate(b.cachedCurrentTemplate)
				msg.reply <- getCurrentTemplateResponse{
					Template: cur,
				}

			case setCurrentTemplateMsg:
				b.cachedCurrentTemplate = deepCopyBlockTemplate(msg.Template)
				msg.reply <- setCurrentTemplateResponse{}

			case getParentTemplateMsg:
				par := deepCopyBlockTemplate(b.cachedParentTemplate)
				msg.reply <- getParentTemplateResponse{
					Template: par,
				}

			case setParentTemplateMsg:
				b.cachedParentTemplate = deepCopyBlockTemplate(msg.Template)
				msg.reply <- setParentTemplateResponse{}

			default:
				log.Warn("Unknown message type", "msg", msg)
			}
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
	forkLen  int64
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
	acceptedTxs []*types.Tx
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
	rateLimit bool, allowHighFees bool) ([]*types.Tx, error) {
	reply := make(chan processTransactionResponse, 1)
	b.msgChan <- processTransactionMsg{tx, allowOrphans, rateLimit,
		allowHighFees, reply}
	response := <-reply
	return response.acceptedTxs, response.err
}

// newPeerMsg signifies a newly connected peer to the block handler.
type newPeerMsg struct {
	peer *peer.ServerPeer
}

// NewPeer informs the block manager of a newly active peer.
func (b *BlockManager) NewPeer(sp *peer.ServerPeer) {
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&b.shutdown) != 0 {
		return
	}
	b.msgChan <- &newPeerMsg{peer: sp}
}

// txMsg packages a tx message and the peer it came from together
// so the block handler has access to that information.
type txMsg struct {
	tx   *types.Tx
	peer *peer.ServerPeer
}
// QueueTx adds the passed transaction message and peer to the block handling
// queue.
func (b *BlockManager) QueueTx(tx *types.Tx, sp *peer.ServerPeer) {
	// Don't accept more transactions if we're shutting down.
	if atomic.LoadInt32(&b.shutdown) != 0 {
		sp.TxProcessed <- struct{}{}
		return
	}

	b.msgChan <- &txMsg{tx: tx, peer: sp}
}

// donePeerMsg signifies a newly disconnected peer to the block handler.
type donePeerMsg struct {
	peer *peer.ServerPeer
}

// DonePeer informs the blockmanager that a peer has disconnected.
func (b *BlockManager) DonePeer(sp *peer.ServerPeer) {
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&b.shutdown) != 0 {
		return
	}
	b.msgChan <- &donePeerMsg{peer: sp}
}

// isCurrentMsg is a message type to be sent across the message channel for
// requesting whether or not the block manager believes it is synced with
// the currently connected peers.
type isCurrentMsg struct {
	reply chan bool
}

// IsCurrent returns whether or not the block manager believes it is synced with
// the connected peers.
func (b *BlockManager) IsCurrent() bool {
	reply := make(chan bool)
	b.msgChan <- isCurrentMsg{reply: reply}
	return <-reply
}

// blockMsg packages a block message and the peer it came from together
// so the block handler has access to that information.
type blockMsg struct {
	block *types.SerializedBlock
	peer  *peer.ServerPeer
}

// QueueBlock adds the passed block message and peer to the block handling queue.
func (b *BlockManager) QueueBlock(block *types.SerializedBlock, sp *peer.ServerPeer) {
	// Don't accept more blocks if we're shutting down.
	if atomic.LoadInt32(&b.shutdown) != 0 {
		sp.BlockProcessed <- struct{}{}
		return
	}

	b.msgChan <- &blockMsg{block: block, peer: sp}
}

// invMsg packages a inv message and the peer it came from together
// so the block handler has access to that information.
type invMsg struct {
	inv  *message.MsgInv
	peer *peer.ServerPeer
}

// QueueInv adds the passed inv message and peer to the block handling queue.
func (b *BlockManager) QueueInv(inv *message.MsgInv, sp *peer.ServerPeer) {
	// No channel handling here because peers do not need to block on inv
	// messages.
	if atomic.LoadInt32(&b.shutdown) != 0 {
		return
	}

	b.msgChan <- &invMsg{inv: inv, peer: sp}
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


// requestFromPeerMsg is a message type to be sent across the message channel
// for requesting either blocks or transactions from a given peer. It routes
// this through the block manager so the block manager doesn't ban the peer
// when it sends this information back.
type requestFromPeerMsg struct {
	peer   *peer.ServerPeer
	blocks []*hash.Hash
	reply  chan requestFromPeerResponse
}

// requestFromPeerResponse is a response sent to the reply channel of a
// requestFromPeerMsg query.
type requestFromPeerResponse struct {
	err error
}

// RequestFromPeer allows an outside caller to request blocks or transactions
// from a peer. The requests are logged in the blockmanager's internal map of
// requests so they do not later ban the peer for sending the respective data.
func (b *BlockManager) RequestFromPeer(p *peer.ServerPeer, blocks[]*hash.Hash) error {
	reply := make(chan requestFromPeerResponse)
	b.msgChan <- requestFromPeerMsg{peer: p, blocks: blocks, reply: reply}
	response := <-reply

	return response.err
}

func (b *BlockManager) requestFromPeer(p *peer.ServerPeer, blocks []*hash.Hash) error {
	msgResp := message.NewMsgGetData()

	// Add the blocks to the request.
	for _, bh := range blocks {
		// If we've already requested this block, skip it.
		_, alreadyReqP := p.RequestedBlocks[*bh]
		_, alreadyReqB := b.requestedBlocks[*bh]

		if alreadyReqP || alreadyReqB {
			continue
		}

		// Check to see if we already have this block, too.
		// If so, skip.
		exists, err := b.chain.HaveBlock(bh)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		err = msgResp.AddInvVect(message.NewInvVect(message.InvTypeBlock, bh))
		if err != nil {
			return fmt.Errorf("unexpected error encountered building request "+
				"for mining state block %v: %v",
				bh, err.Error())
		}

		p.RequestedBlocks[*bh] = struct{}{}
		b.requestedBlocks[*bh] = struct{}{}
		b.requestedEverBlocks[*bh] = 0
	}

	if len(msgResp.InvList) > 0 {
		p.QueueMessage(msgResp, nil)
	}

	return nil
}




