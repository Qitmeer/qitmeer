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

	if code != responseCodeSuccess {
		s.Peers().IncrementBadResponses(stream.Conn().RemotePeer())
		return nil, errors.New(errMsg)
	}

	msg := &pb.DagBlocks{}
	if err := s.Encoding().DecodeWithMaxLength(stream, msg); err != nil {
		return nil, err
	}

	return msg, err
}

func (s *Sync) getBlocksHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) error {
	ctx, cancel := context.WithTimeout(ctx, HandleTimeout)
	var err error
	respCode := responseCodeServerError
	defer func() {
		if respCode != responseCodeSuccess {
			resp, err := s.generateErrorResponse(respCode, err.Error())
			if err != nil {
				log.Error(fmt.Sprintf("Failed to generate a response error:%v", err))
			} else {
				if _, err := stream.Write(resp); err != nil {
					log.Debug(fmt.Sprintf("Failed to write to stream:%v", err))
				}
			}
		}
		closeSteam(stream)
		cancel()
	}()

	SetRPCStreamDeadlines(stream)
	m, ok := msg.(*pb.GetBlocks)
	if !ok {
		err = fmt.Errorf("message is not type *pb.Hash")
		return err
	}
	blocks, _ := s.PeerSync().dagSync.CalcSyncBlocks(nil, changePBHashsToHashs(m.Locator), blockdag.DirectMode, MaxBlockLocatorsPerMsg)

	_, err = stream.Write([]byte{responseCodeSuccess})
	if err != nil {
		return err
	}
	bd := &pb.DagBlocks{Blocks: changeHashsToPBHashs(blocks)}
	_, err = s.Encoding().EncodeWithMaxLength(stream, bd)
	if err != nil {
		return err
	}
	respCode = responseCodeSuccess
	return nil
}

func (ps *PeerSync) processGetBlocks(pe *peers.Peer, blocks []*hash.Hash) error {
	if len(blocks) <= 0 {
		return fmt.Errorf("no blocks")
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
