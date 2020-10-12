package p2p

import (
	"context"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/types"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/peer"
	"sort"
)

func (s *Service) sendGetBlocksRequest(ctx context.Context, id peer.ID, blockhash *hash.Hash) (*pb.BlockData, error) {
	ctx, cancel := context.WithTimeout(ctx, ReqTimeout)
	defer cancel()

	stream, err := s.Send(ctx, &pb.Hash{Hash: blockhash.Bytes()}, RPCGetBlocks, id)
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

	msg := &pb.BlockData{}
	if err := s.Encoding().DecodeWithMaxLength(stream, msg); err != nil {
		return nil, err
	}

	return msg, err
}

func (s *Service) getBlocksHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) error {
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
	ib := s.Chain.BlockDAG().GetBlock(blockHash)
	if ib == nil {
		err = fmt.Errorf("invalid block hash")
		return err
	}
	block, err := s.Chain.FetchBlockByHash(blockHash)
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
	bd := &pb.BlockData{DagID: uint32(ib.GetID()), BlockBytes: blocks}
	_, err = s.Encoding().EncodeWithMaxLength(stream, bd)
	if err != nil {
		return err
	}
	respCode = responseCodeSuccess
	return nil
}

func (s *Service) getBlocks(id peer.ID, blocks []*hash.Hash) error {
	blockdatas := BlockDataSlice{}
	for _, b := range blocks {
		bd, err := s.sendGetBlocksRequest(s.ctx, id, b)
		if err != nil {
			log.Warn(fmt.Sprintf("getBlocks send:%v", err))
			continue
		}
		blockdatas = append(blockdatas, bd)
	}
	log.Trace(fmt.Sprintf("getBlocks:%d", len(blockdatas)))
	if len(blockdatas) <= 0 {
		return fmt.Errorf("no blocks return")
	}
	if len(blockdatas) >= 2 {
		sort.Sort(blockdatas)
	}
	behaviorFlags := blockchain.BFP2PAdd

	add := 0
	for _, bd := range blockdatas {
		block, err := types.NewBlockFromBytes(bd.BlockBytes)
		if err != nil {
			log.Warn(fmt.Sprintf("getBlocks from:%v", err))
			continue
		}
		isOrphan, err := s.Chain.ProcessBlock(block, behaviorFlags)
		if err != nil {
			log.Error("Failed to process block", "hash", block.Hash(), "error", err)
			continue
		}
		if isOrphan {
			continue
		}
		add++
	}
	log.Trace(fmt.Sprintf("getBlocks:%d/%d", add, len(blockdatas)))
	if add > 0 {
		s.TxMemPool.PruneExpiredTx()

		isCurrent := s.peerSync.IsCurrent()
		if isCurrent {
			log.Info("Your synchronization has been completed. ")
		}
	} else {
		return fmt.Errorf("no get blocks")
	}
	return nil
}
