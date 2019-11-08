package peerserver

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/message"
	"github.com/Qitmeer/qitmeer/log"
)

// pushBlockMsg sends a block message for the provided block hash to the
// connected peer.  An error is returned if the block hash is not known.
func (s *PeerServer) pushBlockMsg(sp *serverPeer, hash *hash.Hash, doneChan chan<- struct{}, waitChan <-chan struct{}) error {

	block, err := sp.server.BlockManager.GetChain().FetchBlockByHash(hash)
	if err != nil {
		log.Trace("Unable to fetch requested block hash", "hash", hash,
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
	sp.QueueMessage(&message.MsgBlock{Block: block.Block()}, doneChan)

	return nil
}
