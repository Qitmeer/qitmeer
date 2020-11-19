/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"context"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"time"
)

const (
	retSuccess = iota
	retErrGeneric
	retErrInvalidChainState
)

func (s *Sync) sendChainStateRequest(ctx context.Context, id peer.ID) error {
	pe := s.peers.Get(id)
	if pe == nil {
		return peers.ErrPeerUnknown
	}
	log.Trace(fmt.Sprintf("sendChainStateRequest:%s", id))
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

	if code != ResponseCodeSuccess {
		s.Peers().IncrementBadResponses(stream.Conn().RemotePeer())
		return errors.New(errMsg)
	}

	msg := &pb.ChainState{}
	if err := s.Encoding().DecodeWithMaxLength(stream, msg); err != nil {
		return err
	}

	s.UpdateChainState(pe, msg, true)

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

func (s *Sync) chainStateHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) error {
	defer func() {
		closeSteam(stream)
	}()

	pe := s.peers.Get(stream.Conn().RemotePeer())
	if pe == nil {
		return peers.ErrPeerUnknown
	}
	log.Trace(fmt.Sprintf("chainStateHandler:%s", pe.GetID()))

	ctx, cancel := context.WithTimeout(ctx, HandleTimeout)
	defer cancel()

	SetRPCStreamDeadlines(stream)
	m, ok := msg.(*pb.ChainState)
	if !ok {
		return fmt.Errorf("message is not type *pb.ChainState")
	}

	if ret, err := s.validateChainStateMessage(ctx, m, stream.Conn().RemotePeer()); err != nil {
		log.Debug(fmt.Sprintf("Invalid chain state message from peer:peer=%s  error=%v", stream.Conn().RemotePeer(), err))
		respCode := byte(0)
		switch ret {
		case retErrGeneric:
			respCode = ResponseCodeServerError
		case retErrInvalidChainState:
			// Respond with our status and disconnect with the peer.
			s.UpdateChainState(pe, m, false)
			if err := s.respondWithChainState(ctx, stream); err != nil {
				return err
			}
			closeSteam(stream)
			if err := s.sendGoodByeAndDisconnect(ctx, codeInvalidChainState, stream.Conn().RemotePeer()); err != nil {
				return err
			}
			return nil
		default:
			respCode = ResponseCodeInvalidRequest
			s.Peers().IncrementBadResponses(stream.Conn().RemotePeer())
		}

		originalErr := err
		resp, err := s.generateErrorResponse(respCode, err.Error())
		if err != nil {
			log.Error(fmt.Sprintf("Failed to generate a response error:%v", err))
		} else {
			if _, err := stream.Write(resp); err != nil {
				// The peer may already be ignoring us, as we disagree on fork version, so log this as debug only.
				log.Debug(fmt.Sprintf("Failed to write to stream:%v", err))
			}
		}
		closeSteam(stream)
		// Add a short delay to allow the stream to flush before closing the connection.
		// There is still a chance that the peer won't receive the message.
		time.Sleep(50 * time.Millisecond)
		if err := s.p2p.Disconnect(stream.Conn().RemotePeer()); err != nil {
			log.Error("Failed to disconnect from peer:%v", err)
		}
		return originalErr
	}
	s.UpdateChainState(pe, m, true)

	return s.respondWithChainState(ctx, stream)
}

func (s *Sync) UpdateChainState(pe *peers.Peer, chainState *pb.ChainState, action bool) {
	pe.SetChainState(chainState)
	if !action {
		return
	}
	if pe.ConnectionState().IsConnecting() {
		go s.peerSync.immediatelyConnected(pe)
		return
	}
	go s.peerSync.PeerUpdate(pe, false)
}

func (s *Sync) validateChainStateMessage(ctx context.Context, msg *pb.ChainState, id peer.ID) (int, error) {
	if msg == nil {
		return retErrGeneric, fmt.Errorf("msg is nil")
	}
	if protocol.HasServices(protocol.ServiceFlag(msg.Services), protocol.Relay) {
		return retSuccess, nil
	}
	if protocol.HasServices(protocol.ServiceFlag(msg.Services), protocol.Observer) {
		return retSuccess, nil
	}
	pe := s.peers.Get(id)
	if msg == nil {
		return retErrGeneric, fmt.Errorf("peer is Unkonw:%s", id)
	}
	genesisHash := s.p2p.GetGenesisHash()
	msgGenesisHash, err := hash.NewHash(msg.GenesisHash.Hash)
	if err != nil {
		return retErrGeneric, fmt.Errorf("invalid genesis")
	}
	if !msgGenesisHash.IsEqual(genesisHash) {
		return retErrInvalidChainState, fmt.Errorf("invalid genesis")
	}
	// Notify and disconnect clients that have a protocol version that is
	// too old.
	if msg.ProtocolVersion < uint32(protocol.InitialProcotolVersion) {
		return retErrInvalidChainState, fmt.Errorf("protocol version must be %d or greater",
			protocol.InitialProcotolVersion)
	}
	if msg.GraphState.Total <= 0 {
		return retErrInvalidChainState, fmt.Errorf("invalid graph state")
	}

	if pe.Direction() == network.DirInbound {
		// Reject outbound peers that are not full nodes.
		wantServices := protocol.Full
		if !protocol.HasServices(protocol.ServiceFlag(msg.Services), wantServices) {
			// missingServices := wantServices & ^msg.Services
			missingServices := protocol.MissingServices(protocol.ServiceFlag(msg.Services), wantServices)
			return retErrInvalidChainState, fmt.Errorf("Rejecting peer %s with services %v "+
				"due to not providing desired services %v\n", id.String(), msg.Services, missingServices)
		}
	}

	return retSuccess, nil
}

func (s *Sync) respondWithChainState(ctx context.Context, stream network.Stream) error {
	resp := s.getChainState()
	if resp == nil {
		return fmt.Errorf("no chain state")
	}

	if _, err := stream.Write([]byte{ResponseCodeSuccess}); err != nil {
		log.Error(fmt.Sprintf("Failed to write to stream:%v", err))
	}
	_, err := s.Encoding().EncodeWithMaxLength(stream, resp)
	return err
}

func (s *Sync) getChainState() *pb.ChainState {
	genesisHash := s.p2p.GetGenesisHash()

	cs := &pb.ChainState{
		GenesisHash:     &pb.Hash{Hash: genesisHash.Bytes()},
		ProtocolVersion: s.p2p.Config().ProtocolVersion,
		Timestamp:       uint64(roughtime.Now().Unix()),
		Services:        uint64(s.p2p.Config().Services),
		GraphState:      s.getGraphState(),
		UserAgent:       []byte(s.p2p.Config().UserAgent),
		DisableRelayTx:  s.p2p.Config().DisableRelayTx,
	}

	return cs
}

func (s *Sync) getGraphState() *pb.GraphState {
	bs := s.p2p.BlockChain().BestSnapshot()

	gs := &pb.GraphState{
		Total:      uint32(bs.GraphState.GetTotal()),
		Layer:      uint32(bs.GraphState.GetLayer()),
		MainHeight: uint32(bs.GraphState.GetMainHeight()),
		MainOrder:  uint32(bs.GraphState.GetMainOrder()),
		Tips:       []*pb.Hash{},
	}
	for tip := range bs.GraphState.GetTips().GetMap() {
		gs.Tips = append(gs.Tips, &pb.Hash{Hash: tip.Bytes()})
	}

	return gs
}
