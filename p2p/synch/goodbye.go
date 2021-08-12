/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"context"
	"fmt"
	"github.com/Qitmeer/qitmeer/p2p/common"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/peer"
	"time"
)

// goodbyeRPCHandler reads the incoming goodbye rpc message from the peer.
func (s *Sync) goodbyeRPCHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) *common.Error {
	ctx, cancel := context.WithTimeout(ctx, HandleTimeout)
	defer cancel()

	m, ok := msg.(*uint64)
	if !ok {
		return ErrMessage(fmt.Errorf("wrong message type for goodbye, got %T, wanted *uint64", msg))
	}
	logReason := fmt.Sprintf("Reason:%s", common.ErrorCode(*m).String())
	log.Debug(fmt.Sprintf("Peer has sent a goodbye message:%s (%s)", stream.Conn().RemotePeer(), logReason))
	// closes all streams with the peer
	err := s.p2p.Disconnect(stream.Conn().RemotePeer())
	if err != nil {
		return common.NewError(common.ErrStreamBase, err)
	}
	return nil
}

func (s *Sync) sendGoodByeAndDisconnect(ctx context.Context, code common.ErrorCode, id peer.ID) error {
	return sendGoodByeAndDisconnect(ctx, code, id, s.p2p)
}

func sendGoodByeAndDisconnect(ctx context.Context, code common.ErrorCode, id peer.ID, rpc common.P2PRPC) error {
	if err := sendGoodByeMessage(ctx, code, id, rpc); err != nil {
		log.Debug(fmt.Sprintf("Could not send goodbye message: %v ",err))
	}
	if err := rpc.Disconnect(id); err != nil {
		return err
	}
	return nil
}

func sendGoodByeMessage(ctx context.Context, code common.ErrorCode, id peer.ID, rpc common.P2PRPC) error {
	ctx, cancel := context.WithTimeout(ctx, ReqTimeout)
	defer cancel()

	msg := uint64(code)
	stream, err := Send(ctx, rpc, &msg, RPCGoodByeTopic, id)
	if err != nil {
		return fmt.Errorf("failed send code %v to peer %v : %v ", code, id, err)
	}
	defer func() {
		if err := stream.Reset(); err != nil {
			log.Error(fmt.Sprintf("Failed to reset stream with protocol %s", stream.Protocol()))
		}
	}()
	logReason := fmt.Sprintf("Reason:%s", code.String())
	log.Debug(fmt.Sprintf("Sending Goodbye message to peer:%s (%s)", stream.Conn().RemotePeer(), logReason))
	// Add a short delay to allow the stream to flush before resetting it.
	// There is still a chance that the peer won't receive the message.
	time.Sleep(50 * time.Millisecond)
	return nil
}
