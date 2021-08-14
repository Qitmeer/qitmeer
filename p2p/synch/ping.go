/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"context"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/p2p/common"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/peer"
)

// pingHandler reads the incoming ping rpc message from the peer.
func (s *Sync) pingHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) *common.Error {
	pe := s.peers.Get(stream.Conn().RemotePeer())
	if pe == nil {
		return ErrPeerUnknown
	}

	log.Trace(fmt.Sprintf("pingHandler:%s", pe.GetID()))

	m, ok := msg.(*uint64)
	if !ok {
		return ErrMessage(fmt.Errorf("wrong message type for ping, got %T, wanted *uint64", msg))
	}
	valid, err := s.validateSequenceNum(*m, pe)
	if err != nil {
		return common.NewError(common.ErrDAGConsensus, err)
	}
	e := s.EncodeResponseMsg(stream, s.p2p.MetadataSeq())
	if e != nil {
		return e
	}
	if valid {
		return nil
	}

	// The sequence number was not valid.  Start our own ping back to the peer.
	go func(id peer.ID) {
		// New context so the calling function doesn't cancel on us.
		ctx, cancel := context.WithTimeout(s.p2p.Context(), TtfbTimeout)
		defer cancel()
		md, err := s.sendMetaDataRequest(ctx, id)
		if err != nil {
			log.Debug(fmt.Sprintf("Failed to send metadata request:peer=%s  error=%v", id, err))
			return
		}
		// update metadata if there is no error
		pe.SetMetadata(md)
	}(stream.Conn().RemotePeer())

	return nil
}

func (s *Sync) SendPingRequest(ctx context.Context, id peer.ID) error {
	pe := s.peers.Get(id)
	if pe == nil {
		return peers.ErrPeerUnknown
	}
	log.Trace(fmt.Sprintf("SendPingRequest:%s", pe.GetID()))
	ctx, cancel := context.WithTimeout(ctx, ReqTimeout)
	defer cancel()

	metadataSeq := s.p2p.MetadataSeq()
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
	s.p2p.Host().Peerstore().RecordLatency(id, roughtime.Now().Sub(currentTime))

	if code != 0 {
		s.Peers().IncrementBadResponses(stream.Conn().RemotePeer(), "ping request rsp")
		return errors.New(errMsg)
	}
	msg := new(uint64)
	if err := s.Encoding().DecodeWithMaxLength(stream, msg); err != nil {
		return err
	}
	valid, err := s.validateSequenceNum(*msg, pe)
	if err != nil {
		s.Peers().IncrementBadResponses(stream.Conn().RemotePeer(), "ping request rsp validate seq num")
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

	pe.SetMetadata(md)
	return nil
}

// validates the peer's sequence number.
func (s *Sync) validateSequenceNum(seq uint64, pe *peers.Peer) (bool, error) {
	md := pe.Metadata()
	if md == nil {
		return false, nil
	}
	if md.SeqNumber != seq {
		return false, nil
	}
	return true, nil
}
