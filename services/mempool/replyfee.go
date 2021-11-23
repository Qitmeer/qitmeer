// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2017-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package mempool

import "github.com/Qitmeer/qng-core/core/types"

// calcMinRequiredTxRelayFee returns the minimum transaction fee required for a
// transaction with the passed serialized size to be accepted into the memory
// pool and relayed.
func calcMinRequiredTxRelayFee(serializedSize int64, minRelayTxFee types.Amount) int64 {
	// Calculate the minimum fee for a transaction to be allowed into the
	// mempool and relayed by scaling the base fee (which is the minimum
	// free transaction relay fee).  minTxRelayFee is in Atom/KB, so
	// multiply by serializedSize (which is in bytes) and divide by 1000 to
	// get minimum Atoms.
	// TODO: may add additional layer to handle the fee amount types other
	// than MEER. here by default all coin remain the same fee-rate and does
	// not care about the coin type.
	minFee := (serializedSize * int64(minRelayTxFee.Value)) / 1000

	if minFee == 0 && minRelayTxFee.Value > 0 {
		minFee = int64(minRelayTxFee.Value)
	}

	// Set the minimum fee to the maximum possible value if the calculated
	// fee is not in the valid range for monetary amounts.
	if minFee < 0 || minFee > types.MaxAmount {
		minFee = types.MaxAmount
	}

	return minFee
}
