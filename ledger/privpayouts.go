package ledger

import (
	. "github.com/Qitmeer/qitmeer/core/types"
)

func initPriv() {
	addPayout2("RmBKxMWg4C4EMzYowisDEGSBwmnR6tPgjLs", Amount{5000 * AtomsPerCoin, MEERID}, "76a91437733b37b9f09ce024a5ffbd4570fc1e242c907a88ac")
	addPayout2("RmBKxMWg4C4EMzYowisDEGSBwmnR6tPgjLs", Amount{500 * AtomsPerCoin, QITID}, "76a91437733b37b9f09ce024a5ffbd4570fc1e242c907a88ac")
}
