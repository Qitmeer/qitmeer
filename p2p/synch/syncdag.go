/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"context"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/p2p/common"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/peer"
	"strings"
	"sync/atomic"
)

// MaxBlockLocatorsPerMsg is the maximum number of block locator hashes allowed
// per message.
const MaxBlockLocatorsPerMsg = 2000

func (s *Sync) sendSyncDAGRequest(ctx context.Context, id peer.ID, sd *pb.SyncDAG) (*pb.SubDAG, error) {
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

	if !code.IsSuccess() {
		s.Peers().IncrementBadResponses(stream.Conn().RemotePeer(), "sync DAG request rsp")
		return nil, errors.New(errMsg)
	}
	msg := &pb.SubDAG{}

	if err := s.Encoding().DecodeWithMaxLength(stream, msg); err != nil {
		return nil, err
	}

	return msg, err
}

func (s *Sync) syncDAGHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) *common.Error {
	pe := s.peers.Get(stream.Conn().RemotePeer())
	if pe == nil {
		return ErrPeerUnknown
	}

	ctx, cancel := context.WithTimeout(ctx, HandleTimeout)
	var err error
	defer func() {
		cancel()
	}()

	m, ok := msg.(*pb.SyncDAG)
	if !ok {
		err = fmt.Errorf("message is not type *pb.Hash")
		return ErrMessage(err)
	}
	pe.UpdateGraphState(m.GraphState)

	gs := pe.GraphState()
	blocks, point := s.PeerSync().dagSync.CalcSyncBlocks(gs, changePBHashsToHashs(m.MainLocator), blockdag.SubDAGMode, MaxBlockLocatorsPerMsg)
	pe.UpdateSyncPoint(point)
	/*	if len(blocks) <= 0 {
		err = fmt.Errorf("No blocks")
		return err
	}*/
	sd := &pb.SubDAG{SyncPoint: &pb.Hash{Hash: point.Bytes()}, GraphState: s.getGraphState(), Blocks: changeHashsToPBHashs(blocks)}

	e := s.EncodeResponseMsg(stream, sd)
	if e != nil {
		return e
	}
	return nil
}

func debugSyncDAG(m *pb.SyncDAG) string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("SyncDAG: graphstate=(%v,%v,%v), ",
		m.GraphState.MainOrder, m.GraphState.MainHeight, m.GraphState.Layer,
	))
	sb.WriteString("locator=[")
	size := len(m.MainLocator)
	for i, h := range m.MainLocator {
		sb.WriteString(changePBHashToHash(h).String())
		if i+1 < size {
			sb.WriteString(",")
		}
	}
	sb.WriteString("]")
	sb.WriteString(fmt.Sprintf(", size=%d ", size))
	return sb.String()
}

func (ps *PeerSync) processSyncDAGBlocks(pe *peers.Peer) error {
	log.Trace(fmt.Sprintf("processSyncDAGBlocks peer=%v ", pe.GetID()))
	if !ps.isSyncPeer(pe) || !pe.IsConnected() {
		return fmt.Errorf("no sync peer")
	}

	point := pe.SyncPoint()
	mainLocator := ps.dagSync.GetMainLocator(point)
	sd := &pb.SyncDAG{MainLocator: changeHashsToPBHashs(mainLocator), GraphState: ps.sy.getGraphState()}
	log.Trace(fmt.Sprintf("processSyncDAGBlocks sendSyncDAG point=%v, sd=%v", point.String(), debugSyncDAG(sd)))
	subd, err := ps.sy.sendSyncDAGRequest(ps.sy.p2p.Context(), pe.GetID(), sd)
	if err != nil {
		log.Trace(fmt.Sprintf("processSyncDAGBlocks err=%v ", err.Error()))
		ps.updateSyncPeer(true)
		return err
	}
	log.Trace(fmt.Sprintf("processSyncDAGBlocks result graphstate=(%v,%v,%v), blocks=%v ",
		subd.GraphState.MainOrder, subd.GraphState.MainHeight, subd.GraphState.Layer,
		len(subd.Blocks)))
	pe.UpdateSyncPoint(changePBHashToHash(subd.SyncPoint))
	pe.UpdateGraphState(subd.GraphState)

	if len(subd.Blocks) <= 0 {
		ps.updateSyncPeer(true)
		return fmt.Errorf("No sync dag blocks")
	}
	log.Trace(fmt.Sprintf("processSyncDAGBlocks do GetBlockDatas blocks=%v ", len(subd.Blocks)))
	go ps.GetBlockDatas(pe, changePBHashsToHashs(subd.Blocks))

	return nil
}

func (ps *PeerSync) syncDAGBlocks(pe *peers.Peer) {
	if pe == nil {
		return
	}
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&ps.shutdown) != 0 {
		return
	}

	ps.msgChan <- &syncDAGBlocksMsg{pe: pe}
}
