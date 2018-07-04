// Copyright (c) 2017-2018 The nox developers

package txscript

import (
	"github.com/noxproject/nox/log"
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/engine/txscript"
	btcparams "github.com/noxproject/nox/params/btc"
	btcaddr "github.com/noxproject/nox/params/btc/addr"
)

type PKHashScript struct {
	ops    []txscript.ParsedOpcode
}

func (s *PKHashScript) Match(pops []txscript.ParsedOpcode) bool{
	return len(pops) == 5 &&
		pops[0].GetOpcode().GetValue() == txscript.OP_DUP &&
		pops[1].GetOpcode().GetValue() == txscript.OP_HASH160 &&
		pops[2].GetOpcode().GetValue() == txscript.OP_DATA_20 &&
		pops[3].GetOpcode().GetValue() == txscript.OP_EQUALVERIFY &&
		pops[4].GetOpcode().GetValue() == txscript.OP_CHECKSIG
}
func (s *PKHashScript) SetOpcode(pops []txscript.ParsedOpcode) error {
	s.ops = pops
	return nil
}
var PubKeyHashTy txscript.ScriptClass = txscript.PubKeyHashTy

func (s *PKHashScript) GetClass() txscript.ScriptClass{
	return PubKeyHashTy
}
// requiredSigs = 1
func (s *PKHashScript) GetAddresses() []types.Address {
	// A pay-to-pubkey-hash script is of the form:
	//  OP_DUP OP_HASH160 <hash> OP_EQUALVERIFY OP_CHECKSIG
	// Therefore the pubkey hash is the 3rd item on the stack.
	// Skip the pubkey hash if it's invalid for some reason.
	var addrs []types.Address
	addr, err := btcaddr.NewAddressPubKeyHash(s.ops[2].GetData(),
		&btcparams.MainNetParams)
	if err == nil {
		addrs = append(addrs, addr)
	}
	return addrs
}
func (s *PKHashScript) RequiredSigs() bool {
	return true
}

var _ txscript.Script = (*PKHashScript)(nil)

func init() {
	sin := &PKHashScript{}
	txscript.RegisterScript(sin)
	txscript.UseLogger(log.New(log.Ctx{"module": "txscript engine"}))
}


