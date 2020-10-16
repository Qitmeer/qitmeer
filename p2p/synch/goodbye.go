/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"context"
	"fmt"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/peer"
	"time"
)

const (
	codeGenericError uint64 = iota
	codeInvalidChainState
)

var goodByes = map[uint64]string{
	codeGenericError:      "generic error",
	codeInvalidChainState: "invalid chain state",
}

// goodbyeRPCHandler reads the incoming goodbye rpc message from the peer.
func (s *Sync) goodbyeRPCHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) error {
	defer func() {
		if err := stream.Close(); err != nil {
			log.Error("Failed to close stream")
		}
	}()
	ctx, cancel := context.WithTimeout(ctx, HandleTimeout)
	defer cancel()
	SetRPCStreamDeadlines(stream)

	m, ok := msg.(*uint64)
	if !ok {
		return fmt.Errorf("wrong message type for goodbye, got %T, wanted *uint64", msg)
	}
	logReason := fmt.Sprintf("Reason:%s", goodbyeMessage(*m))
	log.Debug(fmt.Sprintf("Peer has sent a goodbye message:%s (%s)", stream.Conn().RemotePeer(), logReason))
	// closes all streams with the peer
	return s.Disconnect(stream.Conn().RemotePeer())
}

func (s *Sync) sendGoodByeMessage(ctx context.Context, code uint64, id peer.ID) error {
	ctx, cancel := context.WithTimeout(ctx, ReqTimeout)
	defer cancel()

	stream, err := s.Send(ctx, &code, RPCGoodByeTopic, id)
	if err != nil {
		return err
	}
	defer func() {
		if err := stream.Reset(); err != nil {
			log.Error(fmt.Sprintf("Failed to reset stream with protocol %s", stream.Protocol()))
		}
	}()
	logReason := fmt.Sprintf("Reason:%s", goodbyeMessage(code))
	log.Debug(fmt.Sprintf("Sending Goodbye message to peer:%s (%s)", stream.Conn().RemotePeer(), logReason))
	// Add a short delay to allow the stream to flush before resetting it.
	// There is still a chance that the peer won't receive the message.
	time.Sleep(50 * time.Millisecond)
	return nil
}

func (s *Sync) sendGoodByeAndDisconnect(ctx context.Context, code uint64, id peer.ID) error {
	if err := s.sendGoodByeMessage(ctx, code, id); err != nil {
		log.Debug(fmt.Sprintf("Could not send goodbye message to peer, error:%v , peer:%s", err, id))
	}
	if err := s.Disconnect(id); err != nil {
		return err
	}
	return nil
}

// sends a goodbye message for a generic error
func (s *Sync) sendGenericGoodbyeMessage(ctx context.Context, id peer.ID) error {
	return s.sendGoodByeMessage(ctx, codeGenericError, id)
}

func goodbyeMessage(num uint64) string {
	reason, ok := goodByes[num]
	if ok {
		return reason
	}
	return fmt.Sprintf("unknown goodbye value of %d Received", num)
}
