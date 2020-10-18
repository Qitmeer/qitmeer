/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"context"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	libp2pcore "github.com/libp2p/go-libp2p-core"
)

func (s *Sync) sendGraphStateRequest(ctx context.Context, pe *peers.Peer, gs *pb.GraphState) (*pb.GraphState, error) {
	ctx, cancel := context.WithTimeout(ctx, ReqTimeout)
	defer cancel()

	stream, err := s.Send(ctx, gs, RPCGraphState, pe.GetID())
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

	msg := &pb.GraphState{}
	if err := s.Encoding().DecodeWithMaxLength(stream, msg); err != nil {
		return nil, err
	}

	return msg, err
}

func (s *Sync) graphStateHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) error {
	pe := s.peers.Get(stream.Conn().RemotePeer())
	if pe == nil {
		return peers.ErrPeerUnknown
	}

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
	m, ok := msg.(*pb.GraphState)
	if !ok {
		err = fmt.Errorf("message is not type *pb.GraphState")
		return err
	}
	pe.UpdateGraphState(m)

	_, err = stream.Write([]byte{responseCodeSuccess})
	if err != nil {
		return err
	}
	_, err = s.Encoding().EncodeWithMaxLength(stream, s.getGraphState())
	if err != nil {
		return err
	}
	respCode = responseCodeSuccess
	return nil
}

func (s *Sync) updateGraphState(pe *peers.Peer) error {
	gs, err := s.sendGraphStateRequest(s.p2p.Context(), pe, s.getGraphState())
	if err != nil {
		log.Error(err.Error())
		return err
	}
	pe.UpdateGraphState(gs)
	return nil
}
