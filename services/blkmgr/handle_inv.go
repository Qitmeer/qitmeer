package blkmgr

import (
	"github.com/HalalChain/qitmeer-lib/core/message"
	"github.com/HalalChain/qitmeer-lib/params/dcr/types"
)
const (
	// maxRequestedBlocks is the maximum number of requested block
	// hashes to store in memory.
	maxRequestedBlocks = message.MaxInvPerMsg

	// maxRequestedTxns is the maximum number of requested transactions
	// hashes to store in memory.
	maxRequestedTxns = message.MaxInvPerMsg
)

// handleInvMsg handles inv messages from all peers.
// We examine the inventory advertised by the remote peer and act accordingly.
func (b *BlockManager) handleInvMsg(imsg *invMsg) {
	// Attempt to find the final block in the inventory list.  There may
	// not be one.
	lastBlock := -1
	invVects := imsg.inv.InvList
	for i := len(invVects) - 1; i >= 0; i-- {
		if invVects[i].Type == message.InvTypeBlock {
			lastBlock = i
			break
		}
	}

	// If our chain is current and a peer announces a block we already
	// know of, then update their current block height.
	if lastBlock != -1 {
		imsg.peer.UpdateLastGS(imsg.inv.GS)
	}
	// Ignore invs from peers that aren't the sync if we are not current.
	// Helps prevent fetching a mass of orphans.
	if imsg.peer != b.syncPeer && !b.current() {
		return
	}
	// Request the advertised inventory if we don't already have it.  Also,
	// request parent blocks of orphans if we receive one we already have.
	// Finally, attempt to detect potential stalls due to long side chains
	// we already have and request more blocks to prevent them.
	gs:=b.chain.BestSnapshot().GraphState

	for _, iv := range invVects {
		// Ignore unsupported inventory types.
		if iv.Type != message.InvTypeBlock && iv.Type != message.InvTypeTx {
			continue
		}

		// Add the inventory to the cache of known inventory
		// for the peer.
		imsg.peer.AddKnownInventory(iv)

		// Ignore inventory when we're in headers-first mode.
		if b.headersFirstMode {
			continue
		}

		// Request the inventory if we don't already have it.
		haveInv, err := b.haveInventory(iv)
		if err != nil {
			log.Warn("Unexpected failure when checking for "+
				"existing inventory during inv message "+
				"processing","error", err)
			continue
		}
		if !haveInv {
			if iv.Type == message.InvTypeTx {
				// Skip the transaction if it has already been
				// rejected.
				if _, exists := b.rejectedTxns[iv.Hash]; exists {
					continue
				}
			}

			// Add it to the request queue.
			imsg.peer.RequestQueue = append(imsg.peer.RequestQueue, iv)
			continue
		}

		if iv.Type == message.InvTypeBlock {
			// The block is an orphan block that we already have.
			// When the existing orphan was processed, it requested
			// the missing parent blocks.  When this scenario
			// happens, it means there were more blocks missing
			// than are allowed into a single inventory message.  As
			// a result, once this peer requested the final
			// advertised block, the remote peer noticed and is now
			// resending the orphan block as an available block
			// to signal there are more missing blocks that need to
			// be requested.
			if b.chain.IsKnownOrphan(&iv.Hash) {
				// Request blocks starting at the latest known
				// up to the root of the orphan that just came
				// in.
				locator:= b.chain.GetOrphanParents(&iv.Hash)
				if len(locator)==0 {
					log.Error("Failed to get block locator for the orphan block",
						"error")
					continue
				}
				err = imsg.peer.PushGetBlocksMsg(gs,locator)
				if err != nil {
					log.Error("Failed to push getblocksmsg for orphan chain",
						"error", err)
				}
				continue
			}
		}
	}

	// Request as much as possible at once.  Anything that won't fit into
	// the request will be requested on the next inv message.
	numRequested := 0
	gdmsg := message.NewMsgGetData()
	requestQueue := imsg.peer.RequestQueue
	for len(requestQueue) != 0 {
		iv := requestQueue[0]
		requestQueue[0] = nil
		requestQueue = requestQueue[1:]

		switch iv.Type {
		case message.InvTypeBlock:
			// Request the block if there is not already a pending
			// request.
			if _, exists := b.requestedBlocks[iv.Hash]; !exists {
				b.requestedBlocks[iv.Hash] = struct{}{}
				b.requestedEverBlocks[iv.Hash] = 0
				b.limitMap(b.requestedBlocks, maxRequestedBlocks)
				imsg.peer.RequestedBlocks[iv.Hash] = struct{}{}
				gdmsg.AddInvVect(iv)
				numRequested++
			}

		case message.InvTypeTx:
			// Request the transaction if there is not already a
			// pending request.
			if _, exists := b.requestedTxns[iv.Hash]; !exists {
				b.requestedTxns[iv.Hash] = struct{}{}
				b.requestedEverTxns[iv.Hash] = 0
				b.limitMap(b.requestedTxns, maxRequestedTxns)
				imsg.peer.RequestedTxns[iv.Hash] = struct{}{}
				gdmsg.AddInvVect(iv)
				numRequested++
			}
		}

		if numRequested >= wire.MaxInvPerMsg {
			break
		}
	}
	imsg.peer.RequestQueue = requestQueue
	if len(gdmsg.InvList) > 0 {
		imsg.peer.QueueMessage(gdmsg, nil)
	}else{
		err:= imsg.peer.PushGetBlocksMsg(gs,nil)
		if err != nil {
			log.Error("Failed to push getblocksmsg for all chain",
				"error", err)
		}
	}
}
