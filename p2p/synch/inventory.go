package synch

import (
	"context"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/p2p/common"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	libp2pcore "github.com/libp2p/go-libp2p-core"
)

func (s *Sync) sendInventoryRequest(ctx context.Context, pe *peers.Peer, inv *pb.Inventory) error {
	ctx, cancel := context.WithTimeout(ctx, ReqTimeout)
	defer cancel()

	stream, err := s.Send(ctx, inv, RPCInventory, pe.GetID())
	if err != nil {
		log.Trace(fmt.Sprintf("Failed to send inventory request to peer=%v, err=%v", pe.GetID(), err.Error()))
		return err
	}
	defer func() {
		if err := stream.Reset(); err != nil {
			log.Error(fmt.Sprintf("Failed to reset stream with protocol %s,%v", stream.Protocol(), err))
		}
	}()

	code, errMsg, err := ReadRspCode(stream, s.Encoding())
	if err != nil {
		return err
	}

	if !code.IsSuccess() {
		s.Peers().IncrementBadResponses(stream.Conn().RemotePeer(), "inventory request rsp")
		return errors.New(errMsg)
	}
	return err
}

func (s *Sync) inventoryHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) *common.Error {
	pe := s.peers.Get(stream.Conn().RemotePeer())
	if pe == nil {
		return ErrPeerUnknown
	}

	ctx, cancel := context.WithTimeout(ctx, HandleTimeout)
	var err error
	defer func() {
		cancel()
	}()

	m, ok := msg.(*pb.Inventory)
	if !ok {
		err = fmt.Errorf("message is not type *pb.Inventory")
		return ErrMessage(err)
	}
	err = s.handleInventory(m, pe)
	if err != nil {
		return ErrMessage(err)
	}
	e := s.EncodeResponseMsg(stream, nil)
	if e != nil {
		return e
	}
	return nil
}

func (s *Sync) handleInventory(msg *pb.Inventory, pe *peers.Peer) error {
	if len(msg.Invs) <= 0 {
		return nil
	}
	txs := []*hash.Hash{}
	hasBlocks := false
	for _, inv := range msg.Invs {
		h := changePBHashToHash(inv.Hash)
		if InvType(inv.Type) == InvTypeBlock {
			hasBlocks = true
		} else if InvType(inv.Type) == InvTypeTx {
			if s.p2p.Config().DisableRelayTx {
				continue
			}
			if s.haveInventory(inv) {
				continue
			}
			txs = append(txs, h)
		}
	}
	if hasBlocks {
		//s.peerSync.GetBlocks(pe, blocks)
		s.peerSync.UpdateGraphState(pe)
	}
	if len(txs) > 0 {
		go s.peerSync.getTxs(pe, txs)
	}
	return nil
}

// haveInventory returns whether or not the inventory represented by the passed
// inventory vector is known.  This includes checking all of the various places
// inventory can be when it is in different states such as blocks that are part
// of the main chain, on a side chain, in the orphan pool, and transactions that
// are in the memory pool (either the main pool or orphan pool).
func (s *Sync) haveInventory(invVect *pb.InvVect) bool {
	h := changePBHashToHash(invVect.Hash)
	switch InvType(invVect.Type) {
	case InvTypeBlock:
		// Ask chain if the block is known to it in any form (main
		// chain, side chain, or orphan).
		return s.p2p.BlockChain().HaveBlock(h)

	case InvTypeTx:
		// Ask the transaction memory pool if the transaction is known
		// to it in any form (main pool or orphan).

		if s.p2p.TxMemPool().HaveTransaction(h) {
			return true
		}

		prevOut := types.TxOutPoint{Hash: *h}
		for i := uint32(0); i < 2; i++ {
			prevOut.OutIndex = i
			entry, err := s.p2p.BlockChain().FetchUtxoEntry(prevOut)
			if err != nil {
				return false
			}
			if entry != nil && !entry.IsSpent() {
				return true
			}
		}
		return false
	}

	// The requested inventory is is an unsupported type, so just claim
	// it is known to avoid requesting it.
	return true
}
