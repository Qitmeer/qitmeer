package p2p

import (
	"context"
	"fmt"
	"github.com/Qitmeer/qitmeer/core/types"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	libp2pcore "github.com/libp2p/go-libp2p-core"
)

func (s *Service) txHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) error {
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
	m, ok := msg.(*pb.Transaction)
	if !ok {
		err = fmt.Errorf("message is not type *pb.Transaction")
		return err
	}
	err = s.handleTxMsg(m)
	if err != nil {
		return err
	}
	_, err = stream.Write([]byte{responseCodeSuccess})
	if err != nil {
		return err
	}
	respCode = responseCodeSuccess
	return nil
}

func (s *Service) handleTxMsg(msg *pb.Transaction) error {
	tx := changePBTxToTx(msg)
	if tx == nil {
		return fmt.Errorf("message is not type *pb.Transaction")
	}
	// Process the transaction to include validation, insertion in the
	// memory pool, orphan handling, etc.
	allowOrphans := s.cfg.MaxOrphanTxs > 0
	acceptedTxs, err := s.TxMemPool.ProcessTransaction(types.NewTx(tx), allowOrphans, true, true)
	if err != nil {
		return fmt.Errorf("Failed to process transaction %v: %v\n", tx.TxHash().String(), err.Error())
	}
	s.Notify.AnnounceNewTransactions(acceptedTxs)

	return nil
}
