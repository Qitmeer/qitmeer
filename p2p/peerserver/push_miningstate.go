package peerserver

import (
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer-lib/core/message"
)

// pushMiningStateMsg pushes a mining state message to the queue for a
// requesting peer.
func (sp *serverPeer) pushMiningStateMsg(height uint32, blockHashes []hash.Hash) error {
	// Nothing to send, abort.
	if len(blockHashes) == 0 {
		return nil
	}

	// Construct the mining state request and queue it to be sent.
	msg := message.NewMsgMiningState()
	msg.Height = height
	for i := range blockHashes {
		err := msg.AddBlockHash(&blockHashes[i])
		if err != nil {
			return err
		}
	}

	sp.QueueMessage(msg, nil)

	return nil
}
