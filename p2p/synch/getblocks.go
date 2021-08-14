/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"context"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/p2p/common"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/peer"
	"sync/atomic"
)

func (s *Sync) sendGetBlocksRequest(ctx context.Context, id peer.ID, blocks *pb.GetBlocks) (*pb.DagBlocks, error) {
	ctx, cancel := context.WithTimeout(ctx, ReqTimeout)
	defer cancel()

	stream, err := s.Send(ctx, blocks, RPCGetBlocks, id)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := stream.Reset(); err != nil {
			log.Error(fmt.Sprintf("Failed to reset stream with protocol %s,%v", stream.Protocol(), err))
		}
	}()

	code, errMsg, err := ReadRspCode(stream, s.Encoding())
	if err != nil {
		return nil, err
	}

	if !code.IsSuccess() {
		s.Peers().IncrementBadResponses(stream.Conn().RemotePeer(), "get blocks request rsp")
		return nil, errors.New(errMsg)
	}

	msg := &pb.DagBlocks{}
	if err := s.Encoding().DecodeWithMaxLength(stream, msg); err != nil {
		return nil, err
	}

	return msg, err
}

func (s *Sync) getBlocksHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) *common.Error {
	ctx, cancel := context.WithTimeout(ctx, HandleTimeout)
	var err error
	defer func() {
		cancel()
	}()

	m, ok := msg.(*pb.GetBlocks)
	if !ok {
		err = fmt.Errorf("message is not type *pb.Hash")
		return ErrMessage(err)
	}
	blocks, _ := s.PeerSync().dagSync.CalcSyncBlocks(nil, changePBHashsToHashs(m.Locator), blockdag.DirectMode, MaxBlockLocatorsPerMsg)
	bd := &pb.DagBlocks{Blocks: changeHashsToPBHashs(blocks)}
	e := s.EncodeResponseMsg(stream, bd)
	if e != nil {
		return e
	}
	return nil
}

func (ps *PeerSync) processGetBlocks(pe *peers.Peer, blocks []*hash.Hash) error {
	if len(blocks) <= 0 {
		return fmt.Errorf("no blocks")
	}
	if !ps.isSyncPeer(pe) || !pe.IsConnected() {
		return fmt.Errorf("no sync peer")
	}

	db, err := ps.sy.sendGetBlocksRequest(ps.sy.p2p.Context(), pe.GetID(), &pb.GetBlocks{Locator: changeHashsToPBHashs(blocks)})
	if err != nil {
		return err
	}
	if len(db.Blocks) <= 0 {
		log.Warn("no block need to get")
		return nil
	}
	go ps.GetBlockDatas(pe, changePBHashsToHashs(db.Blocks))
	return err
}

func (ps *PeerSync) GetBlocks(pe *peers.Peer, blocks []*hash.Hash) {
	if pe == nil {
		return
	}
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&ps.shutdown) != 0 {
		return
	}
	if len(blocks) == 1 {
		ps.GetBlockDatas(pe, blocks)
		return
	}
	ps.msgChan <- &GetBlocksMsg{pe: pe, blocks: blocks}
}

func (s *Sync) GetDataHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) *common.Error {
	ctx, cancel := context.WithTimeout(ctx, HandleTimeout)
	var err error
	defer func() {
		cancel()
	}()
	pe := s.peers.Get(stream.Conn().RemotePeer())
	if pe == nil {
		return ErrPeerUnknown
	}
	m, ok := msg.(*pb.Inventory)
	if !ok {
		err = fmt.Errorf("message is not type *MsgFilterLoad")
		return ErrMessage(err)
	}
	s.peerSync.msgChan <- &GetDatasMsg{pe: pe, data: m}
	return nil
}
