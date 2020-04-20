// +build !zmq

package zmq

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/core/types"
)

type ZMQBlockHashPublishNotifier struct {
	*ZMQPublishNotifier
}

func (zp *ZMQBlockHashPublishNotifier) Init(cfg *config.Config) error {
	if len(cfg.Zmqpubhashblock) <= 0 {
		return fmt.Errorf("No config")
	}

	if cfg.Zmqpubhashblock == "default" || cfg.Zmqpubhashblock == "*" {
		cfg.Zmqpubhashblock = defaultBlockHashEndpoint
	}
	return zp.initialization(cfg.Zmqpubhashblock)
}

func (zp *ZMQBlockHashPublishNotifier) NotifyBlock(block *types.SerializedBlock) error {
	return zp.sendMessage(block.Hash().Bytes(), false)
}

func (zp *ZMQBlockHashPublishNotifier) NotifyTransaction(txs []*types.Tx) error {
	return nil
}

func (zp *ZMQBlockHashPublishNotifier) Shutdown() {
	zp.shutdown()
}
