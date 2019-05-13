package peerserver

import (
	"qitmeer/common/hash"
	"qitmeer/core/message"
	"qitmeer/log"
)

// pushBlockMsg sends a block message for the provided block hash to the
// connected peer.  An error is returned if the block hash is not known.
func (s *PeerServer) pushBlockMsg(sp *serverPeer, hash *hash.Hash, doneChan chan<- struct{}, waitChan <-chan struct{}) error {

	block, err := sp.server.BlockManager.GetChain().BlockByHash(hash)
	if err != nil {
		log.Trace("Unable to fetch requested block hash", "hash",hash,
			"error", err)

		if doneChan != nil {
			doneChan <- struct{}{}
		}
		return err
	}

	// Once we have fetched data wait for any previous operation to finish.
	if waitChan != nil {
		<-waitChan
	}

	// We only send the channel for this message if we aren't sending
	// an inv straight after.
	var dc chan<- struct{}
	continueHash := sp.continueHash
	sendInv := continueHash != nil && continueHash.IsEqual(hash)
	if !sendInv {
		dc = doneChan
	}
	sp.QueueMessage(&message.MsgBlock{Block:block.Block()}, dc)

	// When the peer requests the final block that was advertised in
	// response to a getblocks message which requested more blocks than
	// would fit into a single message, send it a new inventory message
	// to trigger it to issue another getblocks message for the next
	// batch of inventory.
	if sendInv {
		best := sp.server.BlockManager.GetChain().BestSnapshot()
		invMsg := message.NewMsgInvSizeHint(1)
		iv := message.NewInvVect(message.InvTypeBlock, &best.Hash)
		invMsg.AddInvVect(iv)
		sp.QueueMessage(invMsg, doneChan)
		sp.continueHash = nil
	}
	return nil
}
