/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:metadata.go
 * Date:7/17/20 11:33 AM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package p2p

import (
	"context"
	"fmt"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	"time"
)

// metaDataHandler reads the incoming metadata rpc request from the peer.
func (s *Service) metaDataHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) error {
	defer func() {
		closeSteam(stream)
	}()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	SetRPCStreamDeadlines(stream)

	if _, err := stream.Write([]byte{responseCodeSuccess}); err != nil {
		return err
	}
	_, err := s.Encoding().EncodeWithMaxLength(stream, s.Metadata())
	return err
}

func (s *Service) sendMetaDataRequest(ctx context.Context, id peer.ID) (*pb.MetaData, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	stream, err := s.Send(ctx, new(interface{}), RPCMetaDataTopic, id)
	if err != nil {
		return nil, err
	}
	// we close the stream outside of `send` because
	// metadata requests send no payload, so closing the
	// stream early leads it to a reset.
	defer func() {
		if err := stream.Reset(); err != nil {
			log.Error(fmt.Sprintf("Failed to reset stream for protocol %s  %v", stream.Protocol(), err))
		}
	}()
	code, errMsg, err := ReadStatusCode(stream, s.Encoding())
	if err != nil {
		return nil, err
	}
	if code != 0 {
		s.Peers().IncrementBadResponses(stream.Conn().RemotePeer())
		return nil, errors.New(errMsg)
	}
	msg := new(pb.MetaData)
	if err := s.Encoding().DecodeWithMaxLength(stream, msg); err != nil {
		return nil, err
	}
	return msg, nil
}
