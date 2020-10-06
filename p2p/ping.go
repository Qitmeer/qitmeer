/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:ping.go
 * Date:7/17/20 10:51 AM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package p2p

import (
	"context"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/peer"
)

// pingHandler reads the incoming ping rpc message from the peer.
func (s *Service) pingHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) error {
	SetRPCStreamDeadlines(stream)

	m, ok := msg.(*uint64)
	if !ok {
		closeSteam(stream)
		return fmt.Errorf("wrong message type for ping, got %T, wanted *uint64", msg)
	}
	valid, err := s.validateSequenceNum(*m, stream.Conn().RemotePeer())
	if err != nil {
		closeSteam(stream)
		return err
	}
	if _, err := stream.Write([]byte{responseCodeSuccess}); err != nil {
		closeSteam(stream)
		return err
	}
	if _, err := s.Encoding().EncodeWithMaxLength(stream, s.MetadataSeq()); err != nil {
		closeSteam(stream)
		return err
	}

	if valid {
		closeSteam(stream)
		return nil
	}

	// The sequence number was not valid.  Start our own ping back to the peer.
	go func() {
		defer func() {
			closeSteam(stream)
		}()
		// New context so the calling function doesn't cancel on us.
		ctx, cancel := context.WithTimeout(context.Background(), ttfbTimeout)
		defer cancel()
		md, err := s.sendMetaDataRequest(ctx, stream.Conn().RemotePeer())
		if err != nil {
			log.Debug(fmt.Sprintf("Failed to send metadata request:peer=%s  error=%v", stream.Conn().RemotePeer(), err))
			return
		}
		// update metadata if there is no error
		s.Peers().SetMetadata(stream.Conn().RemotePeer(), md)
	}()

	return nil
}

func (s *Service) sendPingRequest(ctx context.Context, id peer.ID) error {
	ctx, cancel := context.WithTimeout(ctx, ReqTimeout)
	defer cancel()

	metadataSeq := s.MetadataSeq()
	stream, err := s.Send(ctx, &metadataSeq, RPCPingTopic, id)
	if err != nil {
		return err
	}
	currentTime := roughtime.Now()
	defer func() {
		if err := stream.Reset(); err != nil {
			log.Error(fmt.Sprintf("Failed to reset stream with protocol %s", stream.Protocol()))
		}
	}()

	code, errMsg, err := ReadRspCode(stream, s.Encoding())
	if err != nil {
		return err
	}
	// Records the latency of the ping request for that peer.
	s.Host().Peerstore().RecordLatency(id, roughtime.Now().Sub(currentTime))

	if code != 0 {
		s.Peers().IncrementBadResponses(stream.Conn().RemotePeer())
		return errors.New(errMsg)
	}
	msg := new(uint64)
	if err := s.Encoding().DecodeWithMaxLength(stream, msg); err != nil {
		return err
	}
	valid, err := s.validateSequenceNum(*msg, stream.Conn().RemotePeer())
	if err != nil {
		s.Peers().IncrementBadResponses(stream.Conn().RemotePeer())
		return err
	}
	if valid {
		return nil
	}
	md, err := s.sendMetaDataRequest(ctx, stream.Conn().RemotePeer())
	if err != nil {
		// do not increment bad responses, as its
		// already done in the request method.
		return err
	}
	s.Peers().SetMetadata(stream.Conn().RemotePeer(), md)
	return nil
}

// validates the peer's sequence number.
func (s *Service) validateSequenceNum(seq uint64, id peer.ID) (bool, error) {
	md, err := s.Peers().Metadata(id)
	if err != nil {
		return false, err
	}
	if md == nil {
		return false, nil
	}
	if md.SeqNumber != seq {
		return false, nil
	}
	return true, nil
}
