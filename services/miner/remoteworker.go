package miner

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qng-core/core/types/pow"
	"sync"
	"sync/atomic"
)

type RemoteWorker struct {
	started  int32
	shutdown int32

	miner *Miner
	sync.Mutex
}

func (w *RemoteWorker) GetType() string {
	return RemoteWorkerType
}

func (w *RemoteWorker) Start() error {
	err := w.miner.initCoinbase()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	// Already started?
	if atomic.AddInt32(&w.started, 1) != 1 {
		return nil
	}

	log.Info("Start Remote Worker...")
	w.miner.updateBlockTemplate(false)
	return nil
}

func (w *RemoteWorker) Stop() {
	if atomic.AddInt32(&w.shutdown, 1) != 1 {
		log.Warn(fmt.Sprintf("Remote Worker is already in the process of shutting down"))
		return
	}
	log.Info("Stop Remote Worker...")

}

func (w *RemoteWorker) IsRunning() bool {
	return atomic.LoadInt32(&w.started) != 0
}

func (w *RemoteWorker) Update() {
	if atomic.LoadInt32(&w.shutdown) != 0 {
		return
	}
}

func (w *RemoteWorker) GetRequest(powType pow.PowType, reply chan *gbtResponse) {
	if atomic.LoadInt32(&w.shutdown) != 0 {
		close(reply)
		return
	}

	w.Lock()
	defer w.Unlock()

	if w.miner.powType != powType {
		w.miner.powType = powType
		if err := w.miner.updateBlockTemplate(true); err != nil {
			reply <- &gbtResponse{nil, err}
			return
		}
	}
	var headerBuf bytes.Buffer
	err := w.miner.template.Block.Header.Serialize(&headerBuf)
	if err != nil {
		reply <- &gbtResponse{nil, err}
		return
	}
	hexBlockHeader := hex.EncodeToString(headerBuf.Bytes())
	reply <- &gbtResponse{hexBlockHeader, err}
}

func NewRemoteWorker(miner *Miner) *RemoteWorker {
	w := RemoteWorker{
		miner: miner,
	}
	return &w
}
