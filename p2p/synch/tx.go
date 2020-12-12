/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"context"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/p2p/common"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/peer"
	"sync/atomic"
)

func (s *Sync) sendTxRequest(ctx context.Context, id peer.ID, txhash *hash.Hash) (*pb.Transaction, error) {
	ctx, cancel := context.WithTimeout(ctx, ReqTimeout)
	defer cancel()

	stream, err := s.Send(ctx, &pb.Hash{Hash: txhash.Bytes()}, RPCTransaction, id)
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

	if code != ResponseCodeSuccess {
		s.Peers().IncrementBadResponses(stream.Conn().RemotePeer())
		return nil, errors.New(errMsg)
	}

	msg := &pb.Transaction{}
	if err := s.Encoding().DecodeWithMaxLength(stream, msg); err != nil {
		return nil, err
	}

	return msg, err
}

func (s *Sync) txHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) *common.P2PError {
	ctx, cancel := context.WithTimeout(ctx, HandleTimeout)
	var err error
	respCode := ResponseCodeServerError
	defer func() {
		if respCode != ResponseCodeSuccess {
			resp, err := s.generateErrorResponse(respCode, err.Error())
			if err != nil {
				log.Error(fmt.Sprintf("Failed to generate a response error:%v", err))
			} else {
				if _, err := stream.Write(resp); err != nil {
					log.Debug(fmt.Sprintf("Failed to write to stream:%v", err))
				}
			}
		}
		cancel()
	}()

	m, ok := msg.(*pb.Hash)
	if !ok {
		err = fmt.Errorf("message is not type *pb.Transaction")
		return common.NewP2PError(common.ErrMessage, err)
	}
	tx, err := s.p2p.TxMemPool().FetchTransaction(changePBHashToHash(m))
	if err != nil {
		log.Error(fmt.Sprintf("Unable to fetch tx from transaction pool tx:%v", err))
		return common.NewP2PError(common.ErrMessage, err)
	}

	txbytes, err := tx.Tx.Serialize()
	if err != nil {
		return common.NewP2PError(common.ErrMessage, err)
	}

	pbtx := &pb.Transaction{TxBytes: txbytes}
	e := s.EncodeResponseMsg(stream, pbtx)
	if e != nil {
		return e
	}
	respCode = ResponseCodeSuccess
	return nil
}

func (s *Sync) handleTxMsg(msg *pb.Transaction) error {
	tx := changePBTxToTx(msg)
	if tx == nil {
		return fmt.Errorf("message is not type *pb.Transaction")
	}
	// Process the transaction to include validation, insertion in the
	// memory pool, orphan handling, etc.
	allowOrphans := s.p2p.Config().MaxOrphanTxs > 0
	acceptedTxs, err := s.p2p.TxMemPool().ProcessTransaction(types.NewTx(tx), allowOrphans, true, true)
	if err != nil {
		return fmt.Errorf("Failed to process transaction %v: %v\n", tx.TxHash().String(), err.Error())
	}
	s.p2p.Notify().AnnounceNewTransactions(acceptedTxs)

	return nil
}

func (ps *PeerSync) processGetTxs(pe *peers.Peer, txs []*hash.Hash) error {
	for _, txh := range txs {
		tx, err := ps.sy.sendTxRequest(ps.sy.p2p.Context(), pe.GetID(), txh)
		if err != nil {
			return err
		}
		err = ps.sy.handleTxMsg(tx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ps *PeerSync) getTxs(pe *peers.Peer, txs []*hash.Hash) {
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&ps.shutdown) != 0 {
		return
	}

	ps.msgChan <- &getTxsMsg{pe: pe, txs: txs}
}
