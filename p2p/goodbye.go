package p2p

import (
	"context"
	"fmt"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"time"
)

const (
	codeClientShutdown uint64 = iota
	codeWrongNetwork
	codeGenericError
)

var goodByes = map[uint64]string{
	codeClientShutdown: "client shutdown",
	codeWrongNetwork:   "irrelevant network",
	codeGenericError:   "fault/error",
}

// goodbyeRPCHandler reads the incoming goodbye rpc message from the peer.
func (s *Service) goodbyeRPCHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) error {
	defer func() {
		if err := stream.Close(); err != nil {
			log.Error("Failed to close stream")
		}
	}()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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

func goodbyeMessage(num uint64) string {
	reason, ok := goodByes[num]
	if ok {
		return reason
	}
	return fmt.Sprintf("unknown goodbye value of %d Received", num)
}
