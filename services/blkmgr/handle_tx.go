package blkmgr

import (
	"fmt"
	"github.com/Qitmeer/qitmeer-lib/params/dcr/types"
	"github.com/Qitmeer/qitmeer/services/mempool"
)

const (

	// maxRejectedTxns is the maximum number of rejected transactions
	// hashes to store in memory.
	maxRejectedTxns = 1000
)

// handleTxMsg handles transaction messages from all peers.
func (b *BlockManager) handleTxMsg(tmsg *txMsg) {
	sp, exists := b.peers[tmsg.peer.Peer]
	if !exists {
		log.Warn(fmt.Sprintf("Received tx message from unknown peer %s", sp))
		return
	}
	// NOTE:  BitcoinJ, and possibly other wallets, don't follow the spec of
	// sending an inventory message and allowing the remote peer to decide
	// whether or not they want to request the transaction via a getdata
	// message.  Unfortunately, the reference implementation permits
	// unrequested data, so it has allowed wallets that don't follow the
	// spec to proliferate.  While this is not ideal, there is no check here
	// to disconnect peers for sending unsolicited transactions to provide
	// interoperability.
	txHash := tmsg.tx.Hash()

	// Ignore transactions that we have already rejected.  Do not
	// send a reject message here because if the transaction was already
	// rejected, the transaction was unsolicited.
	if _, exists := b.rejectedTxns[*txHash]; exists {
		log.Debug(fmt.Sprintf("Ignoring unsolicited previously rejected "+
			"transaction %v from %s", txHash, tmsg.peer))
		return
	}

	// Process the transaction to include validation, insertion in the
	// memory pool, orphan handling, etc.
	allowOrphans := b.config.MaxOrphanTxs > 0
	acceptedTxs, err := b.chain.GetTxManager().MemPool().ProcessTransaction(tmsg.tx,
		allowOrphans, true, true)

	// Remove transaction from request maps. Either the mempool/chain
	// already knows about it and as such we shouldn't have any more
	// instances of trying to fetch it, or we failed to insert and thus
	// we'll retry next time we get an inv.
	delete(tmsg.peer.RequestedTxns, *txHash)
	delete(b.requestedTxns, *txHash)

	if err != nil {
		// Do not request this transaction again until a new block
		// has been processed.
		b.rejectedTxns[*txHash] = struct{}{}
		b.limitMap(b.rejectedTxns, maxRejectedTxns)

		// When the error is a rule error, it means the transaction was
		// simply rejected as opposed to something actually going wrong,
		// so log it as such.  Otherwise, something really did go wrong,
		// so log it as an actual error.
		if _, ok := err.(mempool.RuleError); ok {
			log.Debug(fmt.Sprintf("Rejected transaction %v from %s: %v",
				txHash, tmsg.peer, err))
		} else {
			log.Error(fmt.Sprintf("Failed to process transaction %v: %v",
				txHash, err))
		}

		// Convert the error into an appropriate reject message and
		// send it.
		code, reason := mempool.ErrToRejectErr(err)
		tmsg.peer.PushRejectMsg(wire.CmdTx, code, reason, txHash,
			false)
		return
	}

	b.notify.AnnounceNewTransactions(acceptedTxs)
}
