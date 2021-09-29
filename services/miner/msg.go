package miner

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/core/types/pow"
)

type StartCPUMiningMsg struct {
}

type CPUMiningGenerateMsg struct {
	discreteNum int
	block       chan *hash.Hash
	powType     pow.PowType
}

type BlockChainChangeMsg struct {
}

type MempoolChangeMsg struct {
}

type gbtResponse struct {
	result interface{}
	err    error
}

type GBTMiningMsg struct {
	request *json.TemplateRequest
	reply   chan *gbtResponse
}

type RemoteMiningMsg struct {
	powType pow.PowType
	reply   chan *gbtResponse
}
