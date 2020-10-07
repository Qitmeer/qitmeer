package p2p

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
)

func (s *Service) sendGetBlocksRequest(ctx context.Context, id peer.ID) error {
	ctx, cancel := context.WithTimeout(ctx, ReqTimeout)
	defer cancel()

	resp := s.getChainState()
	stream, err := s.Send(ctx, resp, RPCChainState, id)
	if err != nil {
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

	if code != responseCodeSuccess {
		s.Peers().IncrementBadResponses(stream.Conn().RemotePeer())
		return errors.New(errMsg)
	}

	msg := &pb.ChainState{}
	if err := s.Encoding().DecodeWithMaxLength(stream, msg); err != nil {
		return err
	}
	s.Peers().SetChainState(stream.Conn().RemotePeer(), msg)

	ret, err := s.validateChainStateMessage(ctx, msg, id)
	if err != nil {
		s.Peers().IncrementBadResponses(stream.Conn().RemotePeer())
		if ret == retErrInvalidChainState {
			if err := s.sendGoodByeAndDisconnect(ctx, codeInvalidChainState, stream.Conn().RemotePeer()); err != nil {
				return err
			}
		}
	}
	return err
}
