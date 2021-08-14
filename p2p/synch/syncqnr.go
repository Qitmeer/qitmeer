/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"context"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/p2p/common"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	"github.com/Qitmeer/qitmeer/p2p/qnode"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"sync/atomic"
)

func (s *Sync) sendQNRRequest(ctx context.Context, pe *peers.Peer, qnr *pb.SyncQNR) (*pb.SyncQNR, error) {
	ctx, cancel := context.WithTimeout(ctx, ReqTimeout)
	defer cancel()

	stream, err := s.Send(ctx, qnr, RPCSyncQNR, pe.GetID())
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

	if !code.IsSuccess() {
		s.Peers().IncrementBadResponses(stream.Conn().RemotePeer(), "QNR request rsp")
		return nil, errors.New(errMsg)
	}

	msg := &pb.SyncQNR{}
	if err := s.Encoding().DecodeWithMaxLength(stream, msg); err != nil {
		return nil, err
	}

	return msg, err
}

func (s *Sync) QNRHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) *common.Error {
	pe := s.peers.Get(stream.Conn().RemotePeer())
	if pe == nil {
		return ErrPeerUnknown
	}

	ctx, cancel := context.WithTimeout(ctx, HandleTimeout)
	var err error
	defer func() {
		cancel()
	}()

	m, ok := msg.(*pb.SyncQNR)
	if !ok {
		err = fmt.Errorf("message is not type *pb.GraphState")
		return ErrMessage(err)
	}

	if pe.QNR() == nil {
		err = s.peerSync.LookupNode(pe, string(m.Qnr))
		if err != nil {
			return ErrMessage(err)
		}
	}

	if s.p2p.Node() == nil {
		return ErrMessage(fmt.Errorf("Disable Node V5"))
	}

	e := s.EncodeResponseMsg(stream, &pb.SyncQNR{Qnr: []byte(s.p2p.Node().String())})
	if e != nil {
		return e
	}
	return nil
}

func (s *Sync) LookupNode(pe *peers.Peer, peNode *qnode.Node) {
	pnResult := s.p2p.Resolve(peNode)
	if pnResult != nil {
		if pe != nil {
			pe.SetQNR(pnResult.Record())
		}
		log.Debug(fmt.Sprintf("Lookup success: %s", pnResult.ID()))
	} else {
		log.Debug(fmt.Sprintf("Lookup fail: %s", peNode.ID()))
	}
}

func (ps *PeerSync) processQNR(msg *SyncQNRMsg) error {
	if !msg.pe.IsConnected() {
		return fmt.Errorf("peer is not active")
	}
	qnr, err := ps.sy.sendQNRRequest(ps.sy.p2p.Context(), msg.pe, &pb.SyncQNR{Qnr: []byte(msg.qnr)})
	if err != nil {
		log.Error(err.Error())
		return err
	}
	if msg.pe.QNR() == nil {
		return ps.LookupNode(msg.pe, string(qnr.Qnr))
	}
	return nil
}

func (ps *PeerSync) SyncQNR(pe *peers.Peer, qnr string) {
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&ps.shutdown) != 0 {
		return
	}

	ps.msgChan <- &SyncQNRMsg{pe: pe, qnr: qnr}
}

func (ps *PeerSync) LookupNode(pe *peers.Peer, qnr string) error {
	peerNode, err := qnode.Parse(qnode.ValidSchemes, qnr)
	if err != nil {
		return err
	}
	ps.sy.LookupNode(pe, peerNode)
	return nil
}
