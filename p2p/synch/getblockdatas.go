/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"context"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/p2p/common"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	"github.com/btcsuite/btcd/wire"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/peer"
	"sync/atomic"
)

const BLOCKDATA_SSZ_HEAD_SIZE = 4

func (s *Sync) sendGetBlockDataRequest(ctx context.Context, id peer.ID, locator *pb.GetBlockDatas) (*pb.BlockDatas, error) {
	ctx, cancel := context.WithTimeout(ctx, ReqTimeout)
	defer cancel()

	stream, err := s.Send(ctx, locator, RPCGetBlockDatas, id)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := stream.Reset()
		if err != nil {
			log.Error(fmt.Sprintf("Failed to close stream with protocol %s,%v", stream.Protocol(), err))
		}
	}()

	code, errMsg, err := ReadRspCode(stream, s.Encoding())
	if err != nil {
		return nil, err
	}

	if !code.IsSuccess() {
		s.Peers().IncrementBadResponses(stream.Conn().RemotePeer())
		return nil, errors.New(errMsg)
	}

	msg := &pb.BlockDatas{}
	if err := s.Encoding().DecodeWithMaxLength(stream, msg); err != nil {
		return nil, err
	}
	return msg, err
}

func (s *Sync) getBlockDataHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) *common.Error {
	ctx, cancel := context.WithTimeout(ctx, HandleTimeout)
	var err error
	defer func() {
		cancel()
	}()

	m, ok := msg.(*pb.GetBlockDatas)
	if !ok {
		err = fmt.Errorf("message is not type *pb.Hash")
		return ErrMessage(err)
	}
	bds := []*pb.BlockData{}
	bd := &pb.BlockDatas{Locator: bds}
	for _, bdh := range m.Locator {
		blockHash, err := hash.NewHash(bdh.Hash)
		if err != nil {
			err = fmt.Errorf("invalid block hash")
			return ErrMessage(err)
		}
		block, err := s.p2p.BlockChain().FetchBlockByHash(blockHash)
		if err != nil {
			return ErrMessage(err)
		}

		blocks, err := block.Bytes()
		if err != nil {
			return ErrMessage(err)
		}
		pbbd := pb.BlockData{BlockBytes: blocks}
		if uint64(bd.SizeSSZ()+pbbd.SizeSSZ()+BLOCKDATA_SSZ_HEAD_SIZE) >= s.p2p.Encoding().GetMaxChunkSize() {
			break
		}
		bd.Locator = append(bd.Locator, &pbbd)
	}
	e := s.EncodeResponseMsg(stream, bd)
	if e != nil {
		err = e.Error
		return e
	}
	return nil
}

func (ps *PeerSync) processGetBlockDatas(pe *peers.Peer, blocks []*hash.Hash) error {
	if !ps.isSyncPeer(pe) || !pe.IsActive() {
		err := fmt.Errorf("no sync peer")
		log.Trace(err.Error())
		return err
	}
	blocksReady := []*hash.Hash{}

	for _, b := range blocks {
		if ps.sy.p2p.BlockChain().HaveBlock(b) {
			continue
		}
		blocksReady = append(blocksReady, b)
	}
	if len(blocksReady) <= 0 {
		return nil
	}
	if !ps.longSyncMod {
		bs := ps.sy.p2p.BlockChain().BestSnapshot()
		if pe.GraphState().GetTotal() >= bs.GraphState.GetTotal()+MaxBlockLocatorsPerMsg {
			ps.longSyncMod = true
		}
	}

	bd, err := ps.sy.sendGetBlockDataRequest(ps.sy.p2p.Context(), pe.GetID(), &pb.GetBlockDatas{Locator: changeHashsToPBHashs(blocksReady)})
	if err != nil {
		log.Warn(fmt.Sprintf("getBlocks send:%v", err))
		return err
	}
	behaviorFlags := blockchain.BFP2PAdd
	add := 0
	hasOrphan := false

	for _, b := range bd.Locator {
		if atomic.LoadInt32(&ps.shutdown) != 0 {
			break
		}
		block, err := types.NewBlockFromBytes(b.BlockBytes)
		if err != nil {
			log.Warn(fmt.Sprintf("getBlocks from:%v", err))
			break
		}
		isOrphan, err := ps.sy.p2p.BlockChain().ProcessBlock(block, behaviorFlags)
		if err != nil {
			log.Error("Failed to process block", "hash", block.Hash(), "error", err)
			break
		}
		if isOrphan {
			hasOrphan = true
			break
		}
		add++
	}
	log.Debug(fmt.Sprintf("getBlockDatas:%d/%d", add, len(bd.Locator)))

	if add > 0 {
		ps.sy.p2p.TxMemPool().PruneExpiredTx()

		if ps.longSyncMod {
			if ps.IsCompleteForSyncPeer() {
				log.Info("Your synchronization has been completed.")
				ps.longSyncMod = false
			}

			if ps.IsCurrent() {
				log.Info("You're up to date now.")
				ps.longSyncMod = false
			}
		}

		if !hasOrphan {
			go ps.UpdateGraphState(pe)
		}
	} else {
		err = fmt.Errorf("no get blocks")
	}
	if add < len(bd.Locator) {
		go ps.PeerUpdate(pe, hasOrphan)
	}
	return err
}

func (ps *PeerSync) GetBlockDatas(pe *peers.Peer, blocks []*hash.Hash) {
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&ps.shutdown) != 0 {
		return
	}

	ps.msgChan <- &GetBlockDatasMsg{pe: pe, blocks: blocks}
}

// handleGetData is invoked when a peer receives a getdata bitcoin message and
// is used to deliver block and transaction information.
func (ps *PeerSync) OnGetData(sp *peers.Peer, msg *pb.Inventory) {
	numAdded := 0
	notFound := wire.NewMsgNotFound()

	length := len(msg.Invs)

	// We wait on this wait channel periodically to prevent queuing
	// far more data than we can send in a reasonable time, wasting memory.
	// The waiting occurs after the database fetch for the next one to
	// provide a little pipelining.
	var waitChan chan struct{}
	doneChan := make(chan struct{}, 1)

	for i, iv := range msg.Invs {
		var c chan struct{}
		// If this will be the last message we send.
		if i == length-1 && len(notFound.InvList) == 0 {
			c = doneChan
		} else if (i+1)%3 == 0 {
			// Buffered so as to not make the send goroutine block.
			c = make(chan struct{}, 1)
		}
		var err error
		switch InvType(iv.Type) {
		case InvTypeTx:
			err = sp.pushTxMsg(sp, &iv.Hash, c, waitChan, types.BaseEncoding)
		case InvTypeBlock:
			err = sp.pushBlockMsg(sp, &iv.Hash, c, waitChan, types.BaseEncoding)
		case InvTypeFilteredBlock:
			err = sp.pushMerkleBlockMsg(sp, &iv.Hash, c, waitChan, types.BaseEncoding)
		default:
			log.Warn(fmt.Sprintf("Unknown type in inventory request %d",
				iv.Type))
			continue
		}
		if err != nil {
			notFound.AddInvVect(iv)

			// When there is a failure fetching the final entry
			// and the done channel was sent in due to there
			// being no outstanding not found inventory, consume
			// it here because there is now not found inventory
			// that will use the channel momentarily.
			if i == len(msg.Invs)-1 && c != nil {
				<-c
			}
		}
		numAdded++
		waitChan = c
	}
	if len(notFound.InvList) != 0 {
		sp.QueueMessage(notFound, doneChan)
	}

	// Wait for messages to be sent. We can send quite a lot of data at this
	// point and this will keep the peer busy for a decent amount of time.
	// We don't process anything else by them in this time so that we
	// have an idea of when we should hear back from them - else the idle
	// timeout could fire when we were only half done sending the blocks.
	if numAdded > 0 {
		<-doneChan
	}
}
