// +build !zmq

package zmq

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/core/types"
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

func (zp *ZMQTxHashPublishNotifier) NotifyTransaction(block []*types.Tx) error {
	return nil
}

func (zp *ZMQTxHashPublishNotifier) Shutdown() {
	zp.shutdown()
}
