// Copyright (c) 2017-2018 The qitmeer developers

package address

import (
	"github.com/Qitmeer/qitmeer/common/encode/base58"
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/rpc"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
	"sync"
)

type AddressApi struct {
	sync.Mutex
	params *params.Params
	config *config.Config
}

type PublicAddressAPI struct {
	addressApi *AddressApi
}

func NewAddressApi(cfg *config.Config, par *params.Params) *AddressApi {
	return &AddressApi{
		config: cfg,
		params: par,
	}
}

func NewPublicAddressAPI(ai *AddressApi) *PublicAddressAPI {
	pmAPI := &PublicAddressAPI{addressApi: ai}
	return pmAPI
}

func (c *AddressApi) APIs() []rpc.API {
	return []rpc.API{
		{
			NameSpace: cmds.DefaultServiceNameSpace,
			Service:   NewPublicAddressAPI(c),
			Public:    true,
		},
	}
}

func (api *PublicAddressAPI) CheckAddress(address string, network string) (interface{}, error) {
	_, ver, err := base58.QitmeerCheckDecode(address)
	if err != nil {
		return false, rpc.RpcInvalidError("Invalid address :" + err.Error())
	}
	var p *params.Params
	switch network {
	case "privnet":
		p = &params.PrivNetParams
	case "testnet":
		p = &params.TestNetParams
	case "mainnet":
		p = &params.MainNetParams
	case "mixnet":
		p = &params.MixNetParams
	default:
		return false, rpc.RpcInvalidError("Invalid network : privnet | testnet | mainnet | mixnet")
	}
	if p.PubKeyHashAddrID != ver {
		return false, rpc.RpcRuleError("address prefix error , need %s , actual: %s,network not match,please check it",
			p.NetworkAddressPrefix, address[0:1])
	}
	return true, nil
}
