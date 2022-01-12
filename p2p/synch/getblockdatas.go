/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"context"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/bloom"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/p2p/common"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/peer"
	"sync/atomic"
	"time"
)

const BLOCKDATA_SSZ_HEAD_SIZE = 4

func (s *Sync) sendGetBlockDataRequest(ctx context.Context, id peer.ID, locator *pb.GetBlockDatas) (*pb.BlockDatas, error) {
	ctx, cancel := context.WithTimeout(ctx, ReqTimeout)
	defer cancel()

	stream, err := s.Send(ctx, locator, RPCGetBlockDatas, id)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := stream.Reset()
		if err != nil {
			log.Error(fmt.Sprintf("Failed to close stream with protocol %s,%v", stream.Protocol(), err))
		}
	}()

	code, errMsg, err := ReadRspCode(stream, s.Encoding())
	if err != nil {
		return nil, err
	}

	if !code.IsSuccess() {
		s.Peers().IncrementBadResponses(stream.Conn().RemotePeer(), "get block date request rsp")
		return nil, errors.New(errMsg)
	}

	msg := &pb.BlockDatas{}
	if err := s.Encoding().DecodeWithMaxLength(stream, msg); err != nil {
		return nil, err
	}
	return msg, err
}

func (s *Sync) sendGetMerkleBlockDataRequest(ctx context.Context, id peer.ID, req *pb.MerkleBlockRequest) (*pb.MerkleBlockResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, ReqTimeout)
	defer cancel()

	stream, err := s.Send(ctx, req, RPCGetMerkleBlocks, id)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := stream.Reset()
		if err != nil {
			log.Error(fmt.Sprintf("Failed to close stream with protocol %s,%v", stream.Protocol(), err))
		}
	}()

	code, errMsg, err := ReadRspCode(stream, s.Encoding())
	if err != nil {
		return nil, err
	}

	if !code.IsSuccess() {
		s.Peers().IncrementBadResponses(stream.Conn().RemotePeer(), "get merkle bock date request rsp")
		return nil, errors.New(errMsg)
	}

	msg := &pb.MerkleBlockResponse{}
	if err := s.Encoding().DecodeWithMaxLength(stream, msg); err != nil {
		return nil, err
	}
	return msg, err
}

func (s *Sync) getBlockDataHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) *common.Error {
	ctx, cancel := context.WithTimeout(ctx, HandleTimeout)
	var err error
	defer func() {
		cancel()
	}()

	m, ok := msg.(*pb.GetBlockDatas)
	if !ok {
		err = fmt.Errorf("message is not type *pb.Hash")
		return ErrMessage(err)
	}
	bds := []*pb.BlockData{}
	bd := &pb.BlockDatas{Locator: bds}
	for _, bdh := range m.Locator {
		blockHash, err := hash.NewHash(bdh.Hash)
		if err != nil {
			err = fmt.Errorf("invalid block hash")
			return ErrMessage(err)
		}
		block, err := s.p2p.BlockChain().FetchBlockByHash(blockHash)
		if err != nil {
			return ErrMessage(err)
		}

		blocks, err := block.Bytes()
		if err != nil {
			return ErrMessage(err)
		}
		pbbd := pb.BlockData{BlockBytes: blocks}
		if uint64(bd.SizeSSZ()+pbbd.SizeSSZ()+BLOCKDATA_SSZ_HEAD_SIZE) >= s.p2p.Encoding().GetMaxChunkSize() {
			break
		}
		bd.Locator = append(bd.Locator, &pbbd)
	}
	e := s.EncodeResponseMsg(stream, bd)
	if e != nil {
		err = e.Error
		return e
	}
	return nil
}

func (s *Sync) getMerkleBlockDataHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) *common.Error {
	ctx, cancel := context.WithTimeout(ctx, HandleTimeout)
	var err error
	defer func() {
		cancel()
	}()
	m, ok := msg.(*pb.MerkleBlockRequest)
	if !ok {
		err = fmt.Errorf("message is not type *pb.Hash")
		return ErrMessage(err)
	}
	pe := s.peers.Get(stream.Conn().RemotePeer())
	if pe == nil {
		return ErrPeerUnknown
	}
	filter := pe.Filter()
	// Do not send a response if the peer doesn't have a filter loaded.
	if !filter.IsLoaded() {
		log.Warn("filter not loaded!")
		return nil
	}
	bds := []*pb.MerkleBlock{}
	bd := &pb.MerkleBlockResponse{Data: bds}
	for _, bdh := range m.Hashes {
		blockHash, err := hash.NewHash(bdh.Hash)
		if err != nil {
			err = fmt.Errorf("invalid block hash")
			return ErrMessage(err)
		}
		block, err := s.p2p.BlockChain().FetchBlockByHash(blockHash)
		if err != nil {
			return ErrMessage(err)
		}
		// Generate a merkle block by filtering the requested block according
		// to the filter for the peer.
		merkle, _ := bloom.NewMerkleBlock(block, filter)
		// Finally, send any matched transactions.
		pbbd := pb.MerkleBlock{Header: merkle.Header.BlockData(),
			Transactions: uint64(merkle.Transactions),
			Hashes:       changeHashsToPBHashs(merkle.Hashes),
			Flags:        merkle.Flags,
		}
		bd.Data = append(bd.Data, &pbbd)
	}
	e := s.EncodeResponseMsg(stream, bd)
	if e != nil {
		err = e.Error
		return e
	}
	return nil
}

func (ps *PeerSync) processGetBlockDatas(pe *peers.Peer, blocks []*hash.Hash) error {
	if !ps.isSyncPeer(pe) || !pe.IsConnected() {
		err := fmt.Errorf("no sync peer")
		log.Trace(err.Error())
		return err
	}
	blocksReady := []*hash.Hash{}
	blockDatas := []*BlockData{}
	blockDataM := map[hash.Hash]*BlockData{}

	for _, b := range blocks {
		if ps.sy.p2p.BlockChain().HaveBlock(b) {
			continue
		}
		blkd:=&BlockData{Hash:b}
		blockDataM[*blkd.Hash]=blkd
		blockDatas = append(blockDatas,blkd)
		if ps.sy.p2p.BlockChain().HasBlockInDB(b) {
			sb,err:=ps.sy.p2p.BlockChain().FetchBlockByHash(b)
			if err == nil {
				blkd.Block=sb
				continue
			}
		}
		blocksReady = append(blocksReady, b)
	}
	if len(blockDatas) <= 0 {
		ps.continueSync(false)
		return nil
	}
	if !ps.longSyncMod {
		bs := ps.sy.p2p.BlockChain().BestSnapshot()
		if pe.GraphState().GetTotal() >= bs.GraphState.GetTotal()+MaxBlockLocatorsPerMsg {
			ps.longSyncMod = true
		}
	}
	if len(blocksReady) > 0 {
		log.Trace(fmt.Sprintf("processGetBlockDatas sendGetBlockDataRequest peer=%v, blocks=%v ", pe.GetID(), blocksReady))
		bd, err := ps.sy.sendGetBlockDataRequest(ps.sy.p2p.Context(), pe.GetID(), &pb.GetBlockDatas{Locator: changeHashsToPBHashs(blocksReady)})
		if err != nil {
			log.Warn(fmt.Sprintf("getBlocks send:%v", err))
			ps.updateSyncPeer(true)
			return err
		}
		log.Trace(fmt.Sprintf("Received:Locator=%d",len(bd.Locator)))
		for _, b := range bd.Locator {
			block, err := types.NewBlockFromBytes(b.BlockBytes)
			if err != nil {
				log.Warn(fmt.Sprintf("getBlocks from:%v", err))
				break
			}
			bd,ok:=blockDataM[*block.Hash()]
			if ok {
				bd.Block = block
			}
		}
	}

	behaviorFlags := blockchain.BFP2PAdd
	add := 0
	hasOrphan := false

	lastSync := ps.lastSync

	for _, b := range blockDatas {
		if atomic.LoadInt32(&ps.shutdown) != 0 {
			break
		}
		block:=b.Block
		if block == nil {
			log.Trace(fmt.Sprintf("No block bytes:%s",b.Hash.String()))
			continue
		}
		isOrphan, err := ps.sy.p2p.BlockChain().ProcessBlock(block, behaviorFlags)
		if err != nil {
			log.Error("Failed to process block", "hash", block.Hash(), "error", err)
			break
		}
		if isOrphan {
			hasOrphan = true
			break
		}
		add++
		ps.lastSync = time.Now()
	}
	log.Debug(fmt.Sprintf("getBlockDatas:%d/%d  spend:%s", add, len(blockDatas), time.Since(lastSync).Truncate(time.Second).String()))

	var err error
	if add > 0 {
		ps.sy.p2p.TxMemPool().PruneExpiredTx()

		if ps.longSyncMod {
			if ps.IsCompleteForSyncPeer() {
				log.Info("Your synchronization has been completed.")
				ps.longSyncMod = false
			}

			if ps.IsCurrent() {
				log.Info("You're up to date now.")
				ps.longSyncMod = false
			}
		}
	} else {
		err = fmt.Errorf("no get blocks")
	}
	ps.continueSync(hasOrphan)
	return err
}

func (ps *PeerSync) processGetMerkleBlockDatas(pe *peers.Peer, blocks []*hash.Hash) error {
	if !ps.isSyncPeer(pe) || !pe.IsConnected() {
		err := fmt.Errorf("no sync peer")
		log.Trace(err.Error())
		return err
	}
	filter := pe.Filter()
	// Do not send a response if the peer doesn't have a filter loaded.
	if !filter.IsLoaded() {
		err := fmt.Errorf("filter not loaded")
		log.Trace(err.Error())
		return nil
	}

	blocksReady := []*hash.Hash{}

	for _, b := range blocks {
		if ps.sy.p2p.BlockChain().HaveBlock(b) {
			continue
		}
		blocksReady = append(blocksReady, b)
	}
	if len(blocksReady) <= 0 {
		return nil
	}
	if !ps.longSyncMod {
		bs := ps.sy.p2p.BlockChain().BestSnapshot()
		if pe.GraphState().GetTotal() >= bs.GraphState.GetTotal()+MaxBlockLocatorsPerMsg {
			ps.longSyncMod = true
		}
	}

	bd, err := ps.sy.sendGetMerkleBlockDataRequest(ps.sy.p2p.Context(), pe.GetID(), &pb.MerkleBlockRequest{Hashes: changeHashsToPBHashs(blocksReady)})
	if err != nil {
		log.Warn(fmt.Sprintf("sendGetMerkleBlockDataRequest send:%v", err))
		return err
	}
	log.Debug(fmt.Sprintf("sendGetMerkleBlockDataRequest:%d", len(bd.Data)))
	return nil
}

func (ps *PeerSync) GetBlockDatas(pe *peers.Peer, blocks []*hash.Hash) {
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&ps.shutdown) != 0 {
		return
	}

	ps.msgChan <- &GetBlockDatasMsg{pe: pe, blocks: blocks}
}

// handleGetData is invoked when a peer receives a getdata qitmeer message and
// is used to deliver block and transaction information.
func (ps *PeerSync) OnGetData(sp *peers.Peer, invList []*pb.InvVect) error {
	txs := make([]*pb.Hash, 0)
	blocks := make([]*pb.Hash, 0)
	merkleBlocks := make([]*pb.Hash, 0)
	for _, iv := range invList {
		log.Trace(fmt.Sprintf("OnGetData:%s (%s)", InvType(iv.Type).String(), changePBHashToHash(iv.Hash)))
		switch InvType(iv.Type) {
		case InvTypeTx:
			txs = append(txs, iv.Hash)
		case InvTypeBlock:
			blocks = append(blocks, iv.Hash)
		case InvTypeFilteredBlock:
			merkleBlocks = append(merkleBlocks, iv.Hash)
		default:
			log.Warn(fmt.Sprintf("Unknown type in inventory request %d",
				iv.Type))
			continue
		}
	}
	if len(txs) > 0 {
		err := ps.processGetTxs(sp, changePBHashsToHashs(txs))
		if err != nil {
			log.Info("processGetTxs Error", "err", err.Error())
			return err
		}
	}
	if len(blocks) > 0 {
		err := ps.processGetBlockDatas(sp, changePBHashsToHashs(blocks))
		if err != nil {
			log.Info("processGetBlockDatas Error", "err", err.Error())
			return err
		}
	}
	if len(merkleBlocks) > 0 {
		err := ps.processGetMerkleBlockDatas(sp, changePBHashsToHashs(merkleBlocks))
		if err != nil {
			log.Info("processGetBlockDatas Error", "err", err.Error())
			return err
		}
	}
	return nil
}
