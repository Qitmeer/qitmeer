package synch

import (
	"context"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
)

func (s *Sync) sendInventoryRequest(ctx context.Context, pe *peers.Peer, inv *pb.Inventory) error {
	ctx, cancel := context.WithTimeout(ctx, ReqTimeout)
	defer cancel()

	stream, err := s.Send(ctx, inv, RPCInventory, pe.GetID())
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
	return err
}
