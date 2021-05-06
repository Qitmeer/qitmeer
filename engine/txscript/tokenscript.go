package txscript

import (
	"github.com/Qitmeer/qitmeer/core/address"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/crypto/ecc"
	"github.com/Qitmeer/qitmeer/params"
)

type TokenNewScript struct {
	pops []ParsedOpcode
}

func (s *TokenNewScript) Name() string {
	return scriptClassToName[TokenPubKeyHashTy]
}

func (s *TokenNewScript) GetClass() ScriptClass {
	return TokenPubKeyHashTy
}

func (s *TokenNewScript) Match(pops []ParsedOpcode) bool {
	return isTokenPubkeyHash(pops)
}

func (s *TokenNewScript) SetOpcode(pops []ParsedOpcode) error {
	s.pops = pops
	return nil
}

func (s *TokenNewScript) GetAddresses() []types.Address {
	var addrs []types.Address
	addr, err := address.NewPubKeyHashAddress(s.pops[8].data, params.ActiveNetParams.Params, ecc.ECDSA_Secp256k1)
	if err == nil {
		addrs = append(addrs, addr)
	} else {
		log.Error(err.Error())
	}
	return addrs
}

func (s *TokenNewScript) RequiredSigs() bool {
	return true
}

func (s *TokenNewScript) GetCoinId() types.CoinID {
	coinId, err := makeScriptNum(s.pops[0].data, true, 5)
	if err != nil {
		log.Error(err.Error())
		return types.CoinID(0)
	}
	return types.CoinID(coinId)
}

func (s *TokenNewScript) GetUpLimit() uint64 {
	limit, err := makeScriptNum(s.pops[1].data, true, 8)
	if err != nil {
		log.Error(err.Error())
		return 0
	}
	return uint64(limit)
}

func (s *TokenNewScript) GetName() string {
	return string(s.pops[2].data)
}
