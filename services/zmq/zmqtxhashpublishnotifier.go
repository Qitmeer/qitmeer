// +build zmq

package zmq

import (
	"fmt"
	"github.com/Qitmeer/qng-core/config"
	"github.com/Qitmeer/qng-core/core/types"
)

type ZMQTxHashPublishNotifier struct {
	*ZMQPublishNotifier
}

func (zp *ZMQTxHashPublishNotifier) Init(cfg *config.Config) error {
	if len(cfg.Zmqpubhashtx) <= 0 {
		return fmt.Errorf("No config")
	}
	if cfg.Zmqpubhashtx == "default" || cfg.Zmqpubhashtx == "*" {
		cfg.Zmqpubhashtx = defaultTxHashEndpoint
	}
	return zp.initialization(cfg.Zmqpubhashtx)
}

func (zp *ZMQTxHashPublishNotifier) NotifyBlock(block *types.SerializedBlock) error {
	return nil
}

func (zp *ZMQTxHashPublishNotifier) NotifyTransaction(txs []*types.Tx) error {
	txsLen := len(txs) - 1
	if txsLen < 0 {
		return nil
	}
	for k, transaction := range txs {
		err := zp.sendMessage(transaction.Hash().Bytes(), k < txsLen)
		if err != nil {
			return err
		}
	}
	return nil
}

func (zp *ZMQTxHashPublishNotifier) Shutdown() {
	zp.shutdown()
}
