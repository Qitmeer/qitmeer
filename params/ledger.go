package params

import (
	"github.com/HalalChain/qitmeer-lib/core/types"
)

// TokenPayout is a payout for block 1 which specifies an address and an amount
// to pay to that address in a transaction output.
type TokenPayout struct {
	Address string
	PkScript []byte
	Amount  uint64
}

// BlockOneLedger specifies the list of payouts in the coinbase of
// genesis. Must be a constant fixed in the code.
// If there are no payouts to be given, set this
// to an empty slice.
// TODO revisit the block one ICO design
var BlockOneLedger []*TokenPayout

// BlockOneSubsidy returns the total subsidy of block height 1 for the
// network.
func BlockOneSubsidy() uint64 {
	if len(BlockOneLedger) == 0 {
		return 0
	}

	sum := uint64(0)
	for _, output := range BlockOneLedger {
		sum += output.Amount
	}

	return sum
}

// pay out tokens to a ledger.
func Ledger(tx *types.Transaction) {
	// Block one is a special block that might pay out tokens to a ledger.
	if len(BlockOneLedger) != 0 {
		// Convert the addresses in the ledger into useable format.
		for _, payout := range BlockOneLedger {
			// Make payout to this address.
			tx.AddTxOut(&types.TxOutput{
				Amount:   payout.Amount,
				PkScript: payout.PkScript,
			})
		}
		tx.TxIn[0].AmountIn = BlockOneSubsidy()
	}
}