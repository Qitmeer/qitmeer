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
	"github.com/Qitmeer/qitmeer/p2p/common"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
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
			log.Trace(fmt.Sprintf("Failed to reset stream with protocol %s,%v", stream.Protocol(), err))
		}
	}()

	code, errMsg, err := ReadRspCode(stream, s.Encoding())
	if err != nil {
		return err
	}
	if !code.IsSuccess() && code != common.ErrDAGConsensus {
		s.Peers().IncrementBadResponses(stream.Conn().RemotePeer(), "chain state request")
		return errors.New(errMsg)
	}

	msg := &pb.ChainState{}
	if err := s.Encoding().DecodeWithMaxLength(stream, msg); err != nil {
		return err
	}

	s.UpdateChainState(pe, msg, code == common.ErrNone)

	if code == common.ErrDAGConsensus {
		if err := s.sendGoodByeAndDisconnect(ctx, common.ErrDAGConsensus, stream.Conn().RemotePeer()); err != nil {
			return err
		}
		return errors.New(errMsg)
	}
	ret, err := s.validateChainStateMessage(ctx, msg, id)
	if err != nil {
		if ret == retErrInvalidChainState {
			if err := s.sendGoodByeAndDisconnect(ctx, common.ErrDAGConsensus, stream.Conn().RemotePeer()); err != nil {
				return err
			}
		} else {
			s.Peers().IncrementBadResponses(stream.Conn().RemotePeer(), "chain state resp")
		}
	}
	return err
}

func (s *Sync) chainStateHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) *common.Error {
	pe := s.peers.Get(stream.Conn().RemotePeer())
	if pe == nil {
		return ErrPeerUnknown
	}
	log.Trace(fmt.Sprintf("chainStateHandler:%s", pe.GetID()))

	ctx, cancel := context.WithTimeout(ctx, HandleTimeout)
	defer cancel()

	m, ok := msg.(*pb.ChainState)
	if !ok {
		return ErrMessage(fmt.Errorf("message is not type *pb.ChainState"))
	}

	if ret, err := s.validateChainStateMessage(ctx, m, stream.Conn().RemotePeer()); err != nil {
		log.Debug(fmt.Sprintf("Invalid chain state message from peer:peer=%s  error=%v", stream.Conn().RemotePeer(), err))
		if ret == retErrInvalidChainState {
			// Respond with our status and disconnect with the peer.
			s.UpdateChainState(pe, m, false)
			if err := s.EncodeResponseMsgPro(stream, s.getChainState(), common.ErrDAGConsensus); err != nil {
				return err
			}
			return nil
		} else if ret != retErrGeneric {
			s.Peers().IncrementBadResponses(stream.Conn().RemotePeer(), "chain state handler")
		}
		return ErrMessage(err)
	}
	if !s.bidirectionalChannelCapacity(pe, stream.Conn()) {
		s.UpdateChainState(pe, m, false)
		if err := s.EncodeResponseMsgPro(stream, s.getChainState(), common.ErrDAGConsensus); err != nil {
			return err
		}
		return nil
	}
	s.UpdateChainState(pe, m, true)
	return s.EncodeResponseMsg(stream, s.getChainState())
}

func (s *Sync) UpdateChainState(pe *peers.Peer, chainState *pb.ChainState, action bool) {
	pe.SetChainState(chainState)
	if !action {
		return
	}
	go s.peerSync.immediatelyConnected(pe)
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
