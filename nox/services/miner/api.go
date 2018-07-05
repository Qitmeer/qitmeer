package miner

import (
	"github.com/noxproject/nox/rpc"
	"fmt"
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/common/util"
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

func (api *PublicMinerAPI) Generate(count uint) ([]string, error) {
	if count > 1000 {
		return nil, fmt.Errorf("error, more than 1000")
	}
	result := []string{}
	for i:=uint(0); i<count; i++ {
		h := hash.HashH(util.ReadSizedRand(nil,32))
		result = append(result,h.String())
	}
	return result,nil
}

