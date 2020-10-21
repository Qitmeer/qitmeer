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
	"github.com/Qitmeer/qitmeer/p2p/peers"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/peer"
	"sync/atomic"
	"time"
)

func (s *Sync) sendGetBlockDataRequest(ctx context.Context, id peer.ID, blockhash *hash.Hash) (*pb.BlockData, error) {
	ctx, cancel := context.WithTimeout(ctx, ReqTimeout)
	defer cancel()

	stream, err := s.Send(ctx, &pb.Hash{Hash: blockhash.Bytes()}, RPCGetBlockDatas, id)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := stream.Reset()
		//err := stream.Close()
		if err != nil {
			log.Error(fmt.Sprintf("Failed to close stream with protocol %s,%v", stream.Protocol(), err))
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

	msg := &pb.BlockData{}
	if err := s.Encoding().DecodeWithMaxLength(stream, msg); err != nil {
		return nil, err
	}

	return msg, err
}

func (s *Sync) getBlockDataHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) error {
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
	m, ok := msg.(*pb.Hash)
	if !ok {
		err = fmt.Errorf("message is not type *pb.Hash")
		return err
	}
	blockHash, err := hash.NewHash(m.Hash)
	if err != nil {
		err = fmt.Errorf("invalid block hash")
		return err
	}
	block, err := s.p2p.BlockChain().FetchBlockByHash(blockHash)
	if err != nil {
		return err
	}

	blocks, err := block.Bytes()
	if err != nil {
		return err
	}
	_, err = stream.Write([]byte{responseCodeSuccess})
	if err != nil {
		return err
	}
	bd := &pb.BlockData{BlockBytes: blocks}
	_, err = s.Encoding().EncodeWithMaxLength(stream, bd)
	if err != nil {
		return err
	}
	respCode = responseCodeSuccess
	return nil
}

func (ps *PeerSync) processGetBlockDatas(pe *peers.Peer, blocks []*hash.Hash) error {
	behaviorFlags := blockchain.BFP2PAdd
	add := 0
	hasOrphan := false
	for _, b := range blocks {
		if !pe.IsActive() {
			break
		}
		time.Sleep(time.Second)
		bd, err := ps.sy.sendGetBlockDataRequest(ps.sy.p2p.Context(), pe.GetID(), b)
		if err != nil {
			log.Warn(fmt.Sprintf("getBlocks send:%v", err))
			continue
		}
		block, err := types.NewBlockFromBytes(bd.BlockBytes)
		if err != nil {
			log.Warn(fmt.Sprintf("getBlocks from:%v", err))
			continue
		}
		isOrphan, err := ps.sy.p2p.BlockChain().ProcessBlock(block, behaviorFlags)
		if err != nil {
			log.Error("Failed to process block", "hash", block.Hash(), "error", err)
			continue
		}
		if isOrphan {
			hasOrphan = true
			break
		}
		add++
	}
	log.Debug(fmt.Sprintf("getBlockDatas:%d/%d", add, len(blocks)))

	var err error
	if add > 0 {
		ps.sy.p2p.TxMemPool().PruneExpiredTx()

		isCurrent := ps.IsCurrent()
		if isCurrent {
			log.Info("Your synchronization has been completed. ")
		}

		if !hasOrphan {
			go ps.UpdateGraphState(pe)
		}
	} else {
		err = fmt.Errorf("no get blocks")
	}
	if add < len(blocks) {
		ps.IntellectSyncBlocks(hasOrphan)
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
