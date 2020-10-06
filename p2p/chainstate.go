package p2p

import (
	"context"
	"errors"
	"fmt"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"time"
)

var (
	errGeneric               = errors.New("generic error")
	errInvalidGenesis        = errors.New("invalid genesis")
	errInvalidProtcolVersion = errors.New("invalid protcol version")
)

func (s *Service) chainStateHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) error {
	defer func() {
		closeSteam(stream)
	}()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	SetRPCStreamDeadlines(stream)
	m, ok := msg.(*pb.ChainState)
	if !ok {
		return fmt.Errorf("message is not type *pb.ChainState")
	}

	if err := s.validateChainStateMessage(ctx, m); err != nil {
		log.Debug(fmt.Sprintf("Invalid chain state message from peer:peer=%s  error=%v", stream.Conn().RemotePeer(), err))

		respCode := byte(0)
		switch err {
		case errGeneric:
			respCode = responseCodeServerError
		case errInvalidGenesis, errInvalidProtcolVersion:
			// Respond with our status and disconnect with the peer.
			s.Peers().SetChainState(stream.Conn().RemotePeer(), m)
			if err := s.respondWithStatus(ctx, stream); err != nil {
				return err
			}
			closeSteam(stream)
			if err := s.sendGoodByeAndDisconnect(ctx, codeInvalidChainState, stream.Conn().RemotePeer()); err != nil {
				return err
			}
			return nil
		default:
			respCode = responseCodeInvalidRequest
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
		if err := s.Disconnect(stream.Conn().RemotePeer()); err != nil {
			log.Error("Failed to disconnect from peer:%v", err)
		}
		return originalErr
	}
	s.Peers().SetChainState(stream.Conn().RemotePeer(), m)

	return s.respondWithStatus(ctx, stream)
}

func (s *Service) validateChainStateMessage(ctx context.Context, msg *pb.ChainState) error {
	// TODO validate

	return nil
}
