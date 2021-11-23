// +build zmq

package zmq

import (
	"fmt"
	"github.com/Qitmeer/qng-core/config"
	"github.com/Qitmeer/qng-core/core/types"
)

type ZMQTxRawPublishNotifier struct {
	*ZMQPublishNotifier
}

func (zp *ZMQTxRawPublishNotifier) Init(cfg *config.Config) error {
	if len(cfg.Zmqpubrawtx) <= 0 {
		return fmt.Errorf("No config")
	}
	if cfg.Zmqpubrawtx == "default" || cfg.Zmqpubrawtx == "*" {
		cfg.Zmqpubrawtx = defaultTxRawEndpoint
	}
	return zp.initialization(cfg.Zmqpubrawtx)
}

func (zp *ZMQTxRawPublishNotifier) NotifyBlock(block *types.SerializedBlock) error {
	return nil
}

func (zp *ZMQTxRawPublishNotifier) NotifyTransaction(txs []*types.Tx) error {
	txsLen := len(txs) - 1
	if txsLen < 0 {
		return nil
	}
	for k, transaction := range txs {
		txBytes, err := transaction.Transaction().Serialize()
		if err != nil {
			return err
		}
		err = zp.sendMessage(txBytes, k < txsLen)
		if err != nil {
			return err
		}
	}
	return nil
}

func (zp *ZMQTxRawPublishNotifier) Shutdown() {
	zp.shutdown()
}
