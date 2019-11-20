package peerserver

import (
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/message"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/log"
)

// handleRelayInvMsg deals with relaying inventory to peers that are not already
// known to have it.  It is invoked from the peerHandler goroutine.
func (s *PeerServer) handleRelayInvMsg(state *peerState, msg relayMsg) {
	log.Trace("handleRelayInvMsg", "msg", msg)
	var gs *blockdag.GraphState
	state.forAllPeers(func(sp *serverPeer) {
		if !sp.Connected() {
			return
		}
		// If the inventory is a block and the peer prefers headers,
		// generate and send a headers message instead of an inventory
		// message.
		if msg.invVect.Type == message.InvTypeBlock && sp.WantsHeaders() {
			blockHeader, ok := msg.data.(types.BlockHeader)
			if !ok {
				log.Warn("Underlying data for headers" +
					" is not a block header")
				return
			}
			msgHeaders := message.NewMsgHeaders(s.BlockManager.GetChain().BestSnapshot().GraphState)
			if err := msgHeaders.AddBlockHeader(&blockHeader); err != nil {
				log.Error("Failed to add block header", "error", err)
				return
			}
			sp.QueueMessage(msgHeaders, nil)
			return
		}

		if msg.invVect.Type == message.InvTypeTx {
			// Don't relay the transaction to the peer when it has
			// transaction relaying disabled.
			if sp.relayTxDisabled() {
				return
			}
		} else if msg.invVect.Type == message.InvTypeBlock {
			gs = s.BlockManager.GetChain().BestSnapshot().GraphState
			sp.QueueInventoryImmediate(msg.invVect, gs)
			return
		}

		// Either queue the inventory to be relayed immediately or with
		// the next batch depending on the immediate flag.
		//
		// It will be ignored in either case if the peer is already
		// known to have the inventory.
		if msg.immediate {
			sp.QueueInventoryImmediate(msg.invVect, gs)
		} else {
			sp.QueueInventory(msg.invVect)
		}
	})
	log.Trace("handleRelayInvMsg done")
}
