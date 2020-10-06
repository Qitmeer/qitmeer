package p2p

import (
	"context"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/core/protocol"
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

func (s *Service) sendChainStateRequest(ctx context.Context, id peer.ID) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
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

	if code != responseCodeSuccess {
		s.Peers().IncrementBadResponses(stream.Conn().RemotePeer())
		return errors.New(errMsg)
	}

	msg := &pb.ChainState{}
	if err := s.Encoding().DecodeWithMaxLength(stream, msg); err != nil {
		return err
	}
	s.Peers().SetChainState(stream.Conn().RemotePeer(), msg)

	ret, err := s.validateChainStateMessage(ctx, msg)
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

	if ret, err := s.validateChainStateMessage(ctx, m); err != nil {
		log.Debug(fmt.Sprintf("Invalid chain state message from peer:peer=%s  error=%v", stream.Conn().RemotePeer(), err))
		respCode := byte(0)
		switch ret {
		case retErrGeneric:
			respCode = responseCodeServerError
		case retErrInvalidChainState:
			// Respond with our status and disconnect with the peer.
			s.Peers().SetChainState(stream.Conn().RemotePeer(), m)
			if err := s.respondWithChainState(ctx, stream); err != nil {
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

	return s.respondWithChainState(ctx, stream)
}

func (s *Service) validateChainStateMessage(ctx context.Context, msg *pb.ChainState) (int, error) {
	if msg == nil {
		return retErrGeneric, fmt.Errorf("msg is nil")
	}
	genesisHash := s.BlockManager.GetChain().BlockDAG().GetGenesisHash()
	msgGenesisHash, err := hash.NewHash(msg.GenesisHash)
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
	return retSuccess, nil
}

func (s *Service) respondWithChainState(ctx context.Context, stream network.Stream) error {
	resp := s.getChainState()
	if resp == nil {
		return fmt.Errorf("no chain state")
	}

	if _, err := stream.Write([]byte{responseCodeSuccess}); err != nil {
		log.Error(fmt.Sprintf("Failed to write to stream:%v", err))
	}
	_, err := s.Encoding().EncodeWithMaxLength(stream, resp)
	return err
}

func (s *Service) getChainState() *pb.ChainState {
	genesisHash := s.BlockManager.GetChain().BlockDAG().GetGenesisHash()
	cs := &pb.ChainState{
		GenesisHash:     genesisHash.Bytes(),
		ProtocolVersion: s.cfg.ProtocolVersion,
		Timestamp:       uint64(roughtime.Now().Unix()),
		Services:        uint64(s.cfg.Services),
		GraphState:      1,
		UserAgent:       []byte(s.cfg.UserAgent),
	}

	return cs
}
