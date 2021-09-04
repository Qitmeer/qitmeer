package p2p

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	"github.com/Qitmeer/qitmeer/params"
	"sync"
	"sync/atomic"
	"time"
)

type broadcastInventoryAdd relayMsg

type broadcastInventoryDel *hash.Hash

type relayMsg struct {
	hash *hash.Hash
	data interface{}
}

type Rebroadcast struct {
	started  int32
	shutdown int32

	wg   sync.WaitGroup
	quit chan struct{}

	modifyRebroadcastInv chan interface{}

	s *Service

	regainMP      bool
	regainMPLimit int
}

func (r *Rebroadcast) Start() {
	// Already started?
	if atomic.AddInt32(&r.started, 1) != 1 {
		return
	}

	log.Info("Starting Rebroadcast")

	r.wg.Add(1)
	go r.handler()
}

func (r *Rebroadcast) Stop() error {
	// Make sure this only happens once.
	if atomic.AddInt32(&r.shutdown, 1) != 1 {
		log.Info("Rebroadcast is already in the process of shutting down")
		return nil
	}

	log.Info("Rebroadcast shutting down")

	close(r.quit)

	r.wg.Wait()
	return nil

}

func (r *Rebroadcast) handler() {
	timer := time.NewTimer(params.ActiveNetParams.TargetTimePerBlock)
	pendingInvs := make(map[hash.Hash]interface{})

out:
	for {
		select {
		case riv := <-r.modifyRebroadcastInv:
			switch msg := riv.(type) {
			case broadcastInventoryAdd:
				pendingInvs[*msg.hash] = msg.data
			case broadcastInventoryDel:
				delete(pendingInvs, *msg)
			}

		case <-timer.C:
			for h, data := range pendingInvs {
				dh := h
				if _, ok := data.(*types.TxDesc); ok {
					if !r.s.TxMemPool().HaveTransaction(&dh) {
						r.RemoveInventory(&dh)
						continue
					}
				}

				r.s.RelayInventory(data, nil)
			}

			rt := int64(len(pendingInvs)/50) * int64(params.ActiveNetParams.TargetTimePerBlock)
			if rt < int64(params.ActiveNetParams.TargetTimePerBlock) {
				rt = int64(params.ActiveNetParams.TargetTimePerBlock)
			}
			timer.Reset(time.Duration(rt))

			r.s.sy.Peers().UpdateBroadcasts()

			r.onRegainMempool()

		case <-r.quit:
			break out
		}
	}
	timer.Stop()

cleanup:
	for {
		select {
		case <-r.modifyRebroadcastInv:
		default:
			break cleanup
		}
	}
	r.wg.Done()
}

func (r *Rebroadcast) AddInventory(h *hash.Hash, data interface{}) {
	// Ignore if shutting down.
	if atomic.LoadInt32(&r.shutdown) != 0 {
		return
	}

	r.modifyRebroadcastInv <- broadcastInventoryAdd{hash: h, data: data}
}

func (r *Rebroadcast) RemoveInventory(h *hash.Hash) {
	// Ignore if shutting down.
	if atomic.LoadInt32(&r.shutdown) != 0 {
		return
	}

	r.modifyRebroadcastInv <- broadcastInventoryDel(h)
}

func (r *Rebroadcast) RegainMempool() {
	if r.regainMP || r.regainMPLimit <= 0 {
		return
	}
	r.regainMP = true
}

func (r *Rebroadcast) onRegainMempool() {
	if !r.regainMP || r.regainMPLimit <= 0 {
		return
	}
	if !r.s.PeerSync().IsCurrent() {
		return
	}
	r.regainMP = false
	r.regainMPLimit--

	r.s.sy.Peers().ForPeers(peers.PeerConnected, func(pe *peers.Peer) {
		r.s.sy.SendMempoolRequest(r.s.Context(), pe)
	})
}

func NewRebroadcast(s *Service) *Rebroadcast {
	r := Rebroadcast{
		s:                    s,
		quit:                 make(chan struct{}),
		modifyRebroadcastInv: make(chan interface{}),
		regainMP:             false,
		regainMPLimit:        1,
	}

	return &r
}
