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

	if code != ResponseCodeSuccess {
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
	respCode := ResponseCodeServerError
	defer func() {
		if respCode != ResponseCodeSuccess {
			resp, err := s.generateErrorResponse(respCode, err.Error())
			if err != nil {
				log.Error(fmt.Sprintf("Failed to generate a response error:%v", err))
			} else {
				if _, err := stream.Write(resp); err != nil {
					log.Debug(fmt.Sprintf("Failed to write to stream:%v", err))
				}
			}
		}
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
	respCode = ResponseCodeSuccess
	return nil
}

func (ps *PeerSync) processGetBlockDatas(pe *peers.Peer, blocks []*hash.Hash) error {
	if !ps.isSyncPeer(pe) || !pe.IsActive() {
		return fmt.Errorf("no sync peer")
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

		if ps.IsCompleteForSyncPeer() {
			log.Info("Your synchronization has been completed.")
		}

		if ps.IsCurrent() {
			log.Info("You're up to date now.")
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
