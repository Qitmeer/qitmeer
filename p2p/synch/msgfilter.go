package synch

import (
	"context"
	"fmt"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/p2p/common"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	libp2pcore "github.com/libp2p/go-libp2p-core"
)

func (s *Sync) HandlerFilterMsgAdd(ctx context.Context, msg interface{}, stream libp2pcore.Stream) *common.Error {
	ctx, cancel := context.WithTimeout(ctx, HandleTimeout)
	var err error
	defer func() {
		cancel()
	}()
	pe := s.peers.Get(stream.Conn().RemotePeer())
	if pe == nil {
		return ErrPeerUnknown
	}
	m, ok := msg.(*pb.FilterAddRequest)
	if !ok {
		err = fmt.Errorf("message is not type *MsgFilterAdd")
		return ErrMessage(err)
	}
	s.peerSync.msgChan <- &OnFilterAddMsg{pe: pe, data: &types.MsgFilterAdd{
		Data: m.Data,
	}}
	return nil
}

func (s *Sync) HandlerFilterMsgClear(ctx context.Context, msg interface{}, stream libp2pcore.Stream) *common.Error {
	ctx, cancel := context.WithTimeout(ctx, HandleTimeout)
	var err error
	defer func() {
		cancel()
	}()
	pe := s.peers.Get(stream.Conn().RemotePeer())
	if pe == nil {
		return ErrPeerUnknown
	}
	_, ok := msg.(*pb.FilterClearRequest)
	if !ok {
		err = fmt.Errorf("message is not type *MsgFilterClear")
		return ErrMessage(err)
	}
	s.peerSync.msgChan <- &OnFilterClearMsg{pe: pe, data: &types.MsgFilterClear{}}
	return nil
}

func (s *Sync) HandlerFilterMsgLoad(ctx context.Context, msg interface{}, stream libp2pcore.Stream) *common.Error {
	ctx, cancel := context.WithTimeout(ctx, HandleTimeout)
	var err error
	defer func() {
		cancel()
	}()
	pe := s.peers.Get(stream.Conn().RemotePeer())
	if pe == nil {
		return ErrPeerUnknown
	}
	m, ok := msg.(*pb.FilterLoadRequest)
	if !ok {
		err = fmt.Errorf("message is not type *MsgFilterLoad")
		return ErrMessage(err)
	}
	s.peerSync.msgChan <- &OnFilterLoadMsg{pe: pe, data: &types.MsgFilterLoad{
		Filter:    m.Filter,
		HashFuncs: uint32(m.HashFuncs),
		Tweak:     uint32(m.Tweak),
		Flags:     types.BloomUpdateType(m.Flags),
	}}
	return nil
}
