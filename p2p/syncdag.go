package p2p

import (
	"context"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/peer"
)

// MaxBlockLocatorsPerMsg is the maximum number of block locator hashes allowed
// per message.
const MaxBlockLocatorsPerMsg = 500

func (s *Service) sendSyncDAGRequest(ctx context.Context, id peer.ID, sd *pb.SyncDAG) (*pb.SubDAG, error) {
	ctx, cancel := context.WithTimeout(ctx, ReqTimeout)
	defer cancel()

	stream, err := s.Send(ctx, sd, RPCSyncDAG, id)
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

	msg := &pb.SubDAG{}
	if err := s.Encoding().DecodeWithMaxLength(stream, msg); err != nil {
		return nil, err
	}

	return msg, err
}

func (s *Service) syncDAGHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) error {
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
	m, ok := msg.(*pb.SyncDAG)
	if !ok {
		err = fmt.Errorf("message is not type *pb.Hash")
		return err
	}
	s.peers.UpdateGraphState(stream.Conn().RemotePeer(), m.GraphState)
	gs, err := s.peers.GraphState(stream.Conn().RemotePeer())
	if !ok {
		err = fmt.Errorf("Graph State error")
		return err
	}

	blocks, point := s.PeerSync().dagSync.CalcSyncBlocks(gs, changeHashs(m.MainLocator), blockdag.SubDAGMode, MaxBlockLocatorsPerMsg)
	if len(blocks) <= 0 {
		err = fmt.Errorf("No blocks")
		return err
	}
	sd := &pb.SubDAG{SyncPoint: &pb.Hash{Hash: point.Bytes()}, GraphState: s.getGraphState(), Blocks: []*pb.BlockData{}}
	for _, blockHash := range blocks {
		block, err := s.Chain.FetchBlockByHash(blockHash)
		if err != nil {
			return err
		}

		blockBytes, err := block.Bytes()
		if err != nil {
			return err
		}
		sd.Blocks = append(sd.Blocks, &pb.BlockData{BlockBytes: blockBytes})
	}
	_, err = stream.Write([]byte{responseCodeSuccess})
	if err != nil {
		return err
	}

	_, err = s.Encoding().EncodeWithMaxLength(stream, sd)
	if err != nil {
		return err
	}

	respCode = responseCodeSuccess
	return nil
}

func (s *Service) syncDAGBlocks(id peer.ID) error {
	sd := &pb.SyncDAG{GraphState: s.getGraphState()}
	_, err := s.sendSyncDAGRequest(s.ctx, id, sd)
	if err != nil {
		return err
	}

	return nil
}
