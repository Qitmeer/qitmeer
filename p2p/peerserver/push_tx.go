package peerserver

import (
	"qitmeer/common/hash"
	"qitmeer/core/message"
	"qitmeer/log"
)

// pushTxMsg sends a tx message for the provided transaction hash to the
// connected peer.  An error is returned if the transaction hash is not known.
func (s *PeerServer) pushTxMsg(sp *serverPeer, hash *hash.Hash, doneChan chan<- struct{}, waitChan <-chan struct{}) error {
	// Attempt to fetch the requested transaction from the pool.  A
	// call could be made to check for existence first, but simply trying
	// to fetch a missing transaction results in the same behavior.
	// Do not allow peers to request transactions already in a block
	// but are unconfirmed, as they may be expensive. Restrict that
	// to the authenticated RPC only.
	tx, _, err := s.TxMemPool.FetchTransaction(hash, false)
	if err != nil {
		log.Trace("Unable to fetch tx %v from transaction pool",
			"tx hash",hash, "error", err)

		if doneChan != nil {
			doneChan <- struct{}{}
		}
		return err
	}

	// Once we have fetched data wait for any previous operation to finish.
	if waitChan != nil {
		<-waitChan
	}

	sp.QueueMessage(&message.MsgTx{Tx:tx.Tx}, doneChan)

	return nil
}
