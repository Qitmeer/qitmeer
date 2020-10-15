package p2p

import (
	"context"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/peer"
)

func (s *Service) sendTxRequest(ctx context.Context, id peer.ID, txhash *hash.Hash) (*pb.Transaction, error) {
	ctx, cancel := context.WithTimeout(ctx, ReqTimeout)
	defer cancel()

	stream, err := s.Send(ctx, &pb.Hash{Hash: txhash.Bytes()}, RPCGetBlocks, id)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := stream.Reset(); err != nil {
			log.Error(fmt.Sprintf("Failed to reset stream with protocol %s,%v", stream.Protocol(), err))
		}
	}()

	code, errMsg, err := ReadRspCode(stream, s.Encoding())
	if err != nil {
		return nil, err
	}

	if code != responseCodeSuccess {
		s.Peers().IncrementBadResponses(stream.Conn().RemotePeer())
		return nil, errors.New(errMsg)
	}

	msg := &pb.Transaction{}
	if err := s.Encoding().DecodeWithMaxLength(stream, msg); err != nil {
		return nil, err
	}

	return msg, err
}

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
	m, ok := msg.(*pb.Hash)
	if !ok {
		err = fmt.Errorf("message is not type *pb.Transaction")
		return err
	}
	tx, err := s.TxMemPool.FetchTransaction(changePBHashToHash(m))
	if err != nil {
		log.Error(fmt.Sprintf("Unable to fetch tx from transaction pool tx:%v", err))
		return err
	}
	if err != nil {
		return err
	}

	txbytes, err := tx.Tx.Serialize()
	if err != nil {
		return err
	}

	_, err = stream.Write([]byte{responseCodeSuccess})
	if err != nil {
		return err
	}

	pbtx := &pb.Transaction{TxBytes: txbytes}
	_, err = s.Encoding().EncodeWithMaxLength(stream, pbtx)
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

func (s *Service) getTx(id peer.ID, txHash *hash.Hash) error {
	tx, err := s.sendTxRequest(s.ctx, id, txHash)
	if err != nil {
		return err
	}
	return s.handleTxMsg(tx)
}
