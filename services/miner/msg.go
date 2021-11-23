package miner

import (
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/core/json"
	"github.com/Qitmeer/qng-core/core/types/pow"
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
