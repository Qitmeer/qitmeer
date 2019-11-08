package ledger

import (
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/core/types"
)

// TokenPayout is a payout for block 1 which specifies an address and an amount
// to pay to that address in a transaction output.
type TokenPayout struct {
	Address  string
	PkScript []byte
	Amount   uint64
}

// GenesisLedger specifies the list of payouts in the coinbase of
// genesis. Must be a constant fixed in the code.
// If there are no payouts to be given, set this
// to an empty slice.
var GenesisLedger []*TokenPayout

// BlockOneSubsidy returns the total subsidy of block height 1 for the
// network.
func GenesisLedgerSubsidy() uint64 {
	if len(GenesisLedger) == 0 {
		return 0
	}

	sum := uint64(0)
	for _, output := range GenesisLedger {
		sum += output.Amount
	}

	return sum
}

func addPayout(addr string, amount uint64, pksStr string) {
	pks, err := hex.DecodeString(pksStr)
	if err != nil {
		fmt.Printf("Error %v - address:%s  amount:%d\n", err, addr, amount)
		return
	}
	GenesisLedger = append(GenesisLedger, &TokenPayout{addr, pks, amount})
	//fmt.Printf("Add payout (%d) - address:%s  amount:%d\n",len(GenesisLedger),addr,amount)
}

// pay out tokens to a ledger.
func Ledger(tx *types.Transaction, netType protocol.Network) {
	switch netType {
	case protocol.MainNet:
		initMain()
	case protocol.TestNet:
		initTest()
	case protocol.PrivNet:
		initPriv()
	}

	// Block one is a special block that might pay out tokens to a ledger.
	if len(GenesisLedger) != 0 {
		// Convert the addresses in the ledger into useable format.
		for _, payout := range GenesisLedger {
			// Make payout to this address.
			tx.AddTxOut(&types.TxOutput{
				Amount:   payout.Amount,
				PkScript: payout.PkScript,
			})
		}
	}
}
