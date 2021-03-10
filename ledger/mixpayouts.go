package ledger

import (
	. "github.com/Qitmeer/qitmeer/core/types"
)

func initMix() {

	addPayout2("XmspWkqJv6a4sziWrPZbSWQ37WoNEmTD1xm",
		Amount{Value: 10000 * AtomsPerCoin, Id: QITID}, "76a914cec9c6fb443a62b5e8d7a00fc3ad0f3cef1f070588ac")
	addPayout2("XmspWkqJv6a4sziWrPZbSWQ37WoNEmTD1xm",
		Amount{Value: 10000 * AtomsPerCoin, Id: METID}, "76a914cec9c6fb443a62b5e8d7a00fc3ad0f3cef1f070588ac")
	addPayout2("XmspWkqJv6a4sziWrPZbSWQ37WoNEmTD1xm",
		Amount{Value: 10000 * AtomsPerCoin, Id: TERID}, "76a914cec9c6fb443a62b5e8d7a00fc3ad0f3cef1f070588ac")

	addPayout2("XmkgU8m4G2GwRjz6rEVskG9HAabT5uUS8Fy",
		Amount{Value: 20000 * AtomsPerCoin, Id: QITID}, "76a914807b61617f12c3166f4531ddeae484b82187ae2f88ac")
	addPayout2("XmkgU8m4G2GwRjz6rEVskG9HAabT5uUS8Fy",
		Amount{Value: 20000 * AtomsPerCoin, Id: METID}, "76a914807b61617f12c3166f4531ddeae484b82187ae2f88ac")
	addPayout2("XmkgU8m4G2GwRjz6rEVskG9HAabT5uUS8Fy",
		Amount{Value: 20000 * AtomsPerCoin, Id: TERID}, "76a914807b61617f12c3166f4531ddeae484b82187ae2f88ac")

	addPayout2("XmnsdQkQYWHih65kMyZPo5bFzRpEyGc3N9x",
		Amount{Value: 30000 * AtomsPerCoin, Id: QITID}, "76a9149887f352a02c4e60d99bcd2eab33c8b7b0198b0488ac")
	addPayout2("XmnsdQkQYWHih65kMyZPo5bFzRpEyGc3N9x",
		Amount{Value: 30000 * AtomsPerCoin, Id: METID}, "76a9149887f352a02c4e60d99bcd2eab33c8b7b0198b0488ac")
	addPayout2("XmnsdQkQYWHih65kMyZPo5bFzRpEyGc3N9x",
		Amount{Value: 30000 * AtomsPerCoin, Id: TERID}, "76a9149887f352a02c4e60d99bcd2eab33c8b7b0198b0488ac")

}