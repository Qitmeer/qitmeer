package p2p

import (
	"context"
	"fmt"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/peer"
)

func (s *Service) syncDAGHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) error {
	ctx, cancel := context.WithTimeout(ctx, HandleTimeout)
	var err error
	respCode := responseCodeServerError
	defer func() {
		if respCode != responseCodeSuccess {
			resp, err := s.generateErrorResponse(respCode, err.Error())
			if err != nil {
				log.Error(fmt.Sprintf("Failed to generate a response error:%v", err))
			} else {
				if _, err := stream.Write(resp); err != nil {
					log.Debug(fmt.Sprintf("Failed to write to stream:%v", err))
				}
			}
		}
		closeSteam(stream)
		cancel()
	}()

	SetRPCStreamDeadlines(stream)
	/*m, ok := msg.(*pb.SyncDAG)
	if !ok {
		err = fmt.Errorf("message is not type *pb.Hash")
		return err
	}*/

	respCode = responseCodeSuccess
	return nil
}

func (s *Service) syncDAGBlocks(id peer.ID) error {
	return nil
}
