// +build zmq

package zmq

import (
	"fmt"
	"github.com/Qitmeer/qng-core/config"
	"github.com/Qitmeer/qng-core/core/types"
)

type ZMQBlockRawPublishNotifier struct {
	*ZMQPublishNotifier
}

func (zp *ZMQBlockRawPublishNotifier) Init(cfg *config.Config) error {
	if len(cfg.Zmqpubrawblock) <= 0 {
		return fmt.Errorf("No config")
	}
	if cfg.Zmqpubrawblock == "default" || cfg.Zmqpubrawblock == "*" {
		cfg.Zmqpubrawblock = defaultBlockRawEndpoint
	}
	return zp.initialization(cfg.Zmqpubrawblock)
}

func (zp *ZMQBlockRawPublishNotifier) NotifyBlock(block *types.SerializedBlock) error {
	blockBytes, err := block.Bytes()
	if err != nil {
		log.Error("block bytes:%v", err)
		return err
	}
	return zp.sendMessage(blockBytes, false)
}

func (zp *ZMQBlockRawPublishNotifier) NotifyTransaction(txs []*types.Tx) error {
	return nil
}

func (zp *ZMQBlockRawPublishNotifier) Shutdown() {
	zp.shutdown()
}
