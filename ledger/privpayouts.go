package ledger

import (
	"github.com/Qitmeer/qitmeer/core/types"
)

func initPriv() {
	addrS := "RmBKxMWg4C4EMzYowisDEGSBwmnR6tPgjLs"
	// txscript.PayToAddrScript(addr)
	addPayout2(addrS, types.Amount{Value: 5 * types.AtomsPerCoin, Id: types.QITID}, "76a91437733b37b9f09ce024a5ffbd4570fc1e242c907a88ac")
	addPayout2(addrS, types.Amount{Value: 5 * types.AtomsPerCoin, Id: types.MEERID}, "76a91437733b37b9f09ce024a5ffbd4570fc1e242c907a88ac")

	// lock 2021-04-09  1617897600 txscript.PayToAddrLockScript(addr,1617897600)
	// addPayoutLock(addrS, types.Amount{Value: 500 * types.AtomsPerCoin, Id: types.MEERID}, "76a91437733b37b9f09ce024a5ffbd4570fc1e242c907a88acb10480286f60", 1617897600)

	// lock 2022-04-09 1649433600 txscript.PayToAddrLockScript(addr,1649433600)
	// addPayoutLock(addrS, types.Amount{Value: 500 * types.AtomsPerCoin, Id: types.QITID}, "76a91437733b37b9f09ce024a5ffbd4570fc1e242c907a88acb104005c5062", 1649433600)
}
