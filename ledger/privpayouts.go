package ledger

import (
	. "github.com/Qitmeer/qitmeer/core/types"
)

func initPriv() {
	addPayout2("RmBKxMWg4C4EMzYowisDEGSBwmnR6tPgjLs", Amount{Value: 5000 * AtomsPerCoin, Id: MEERID}, "76a91437733b37b9f09ce024a5ffbd4570fc1e242c907a88ac")
	addPayout2("RmBKxMWg4C4EMzYowisDEGSBwmnR6tPgjLs", Amount{Value: 500 * AtomsPerCoin, Id: QITID}, "76a91437733b37b9f09ce024a5ffbd4570fc1e242c907a88ac")
}
