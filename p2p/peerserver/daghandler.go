package peerserver

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/message"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/p2p/connmgr"
	"github.com/Qitmeer/qitmeer/p2p/peer"
)

// OnGraphState
func (sp *serverPeer) OnGraphState(p *peer.Peer, msg *message.MsgGraphState) {
	p.UpdateLastGS(msg.GS)
}

// OnSyncResult
func (sp *serverPeer) OnSyncResult(p *peer.Peer, msg *message.MsgSyncResult) {
	p.UpdateLastGS(msg.GS)

	if msg.Mode == blockdag.SubDAGMode {
		chain := sp.server.BlockManager.GetChain()
		gs := chain.BestSnapshot().GraphState
		mainLocator := sp.server.BlockManager.DAGSync().GetMainLocator(p.PrevGet.Point)

		sp.PushSyncDAGMsg(gs, mainLocator)
	}
}

func (sp *serverPeer) OnSyncDAG(p *peer.Peer, msg *message.MsgSyncDAG) {
	if msg.GS.IsGenesis() && !msg.GS.GetTips().HasOnly(sp.server.chainParams.GenesisHash) {
		sp.addBanScore(0, connmgr.SeriousScore, "onsyncdag")
		log.Warn(fmt.Sprintf("Wrong genesis(%s) from peer(%s),your genesis is %s",
			msg.GS.GetTips().List()[0].String(), p.String(), sp.server.chainParams.GenesisHash.String()))
		return
	}

	p.UpdateLastGS(msg.GS)

	chain := sp.server.BlockManager.GetChain()
	dagSync := sp.server.BlockManager.DAGSync()
	gs := chain.BestSnapshot().GraphState
	blocks, point := dagSync.CalcSyncBlocks(msg.GS, msg.MainLocator, blockdag.SubDAGMode, message.MaxBlockLocatorsPerMsg)

	if point != nil {
		spmsg := message.NewMsgSyncPoint(gs.Clone(), point)
		p.QueueMessage(spmsg, nil)
	}
	hsLen := len(blocks)
	if hsLen == 0 {
		log.Trace(fmt.Sprintf("Sorry, there are not these blocks for %s", p.String()))
		return
	}

	invMsg := message.NewMsgInv()
	invMsg.GS = gs
	for i := 0; i < hsLen; i++ {
		iv := message.NewInvVect(message.InvTypeBlock, blocks[i])
		invMsg.AddInvVect(iv)
	}
	if len(invMsg.InvList) > 0 {
		p.QueueMessage(invMsg, nil)
	}
}

func (sp *serverPeer) OnSyncPoint(p *peer.Peer, msg *message.MsgSyncPoint) {
	p.UpdateLastGS(msg.GS)

	if sp.server.BlockManager.GetChain().BlockDAG().HasBlock(msg.SyncPoint) {
		p.PrevGet.UpdatePoint(msg.SyncPoint)
	}
}
