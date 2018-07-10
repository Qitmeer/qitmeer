// Copyright (c) 2017-2018 The nox developers

package miner

import (
	"github.com/noxproject/nox/rpc"
	"fmt"
	"github.com/noxproject/nox/services/common/error"
)

func (c *CPUMiner)	APIs() []rpc.API {
	return []rpc.API{
		{
			NameSpace: rpc.DefaultServiceNameSpace,
			Service:   NewPublicMinerAPI(c),
		},
	}
}


type PublicMinerAPI struct {
	miner *CPUMiner
}

func NewPublicMinerAPI(c *CPUMiner) *PublicMinerAPI {
	return &PublicMinerAPI{c}
}

func (api *PublicMinerAPI) Generate(numBlocks uint32) ([]string, error) {
	// Respond with an error if there are no addresses to pay the
	// created blocks to.
	if len(api.miner.config.GetMinningAddrs()) == 0 {
		return nil, er.RpcInternalError("No payment addresses specified "+
			"via --miningaddr", "Configuration")
	}

	// Respond with an error if the client is requesting 0 blocks to be generated.
	if numBlocks == 0 {
		return nil, er.RpcInternalError("Invalid number of blocks",
			"Configuration")
	}
	if numBlocks > 1000 {
		return nil, fmt.Errorf("error, more than 1000")
	}
	blockHashes, err := api.miner.GenerateNBlocks(numBlocks)
	if err != nil {
		return nil, er.RpcInternalError("Could not generate blocks",
			"Configuration")
	}
	// Create a reply
	reply := make([]string, numBlocks)

	// Mine the correct number of blocks, assigning the hex representation of the
	// hash of each one to its place in the reply.
	for i, hash := range blockHashes {
		reply[i] = hash.String()
	}

	return reply, nil
}

