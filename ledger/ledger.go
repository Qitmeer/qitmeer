package ledger

import (
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/core/types"
)

type LedgerParams struct {
	GenesisAmountUnit int64 // the unit amount of equally divided.
	MaxLockHeight     int64 // the max lock height
}

const PercentBase = 10000

// TokenPayout is a payout for block 1 which specifies an address and an amount
// to pay to that address in a transaction output.
type TokenPayout struct {
	Address  string
	PkScript []byte
	Amount   types.Amount
}

type TokenPayoutReGen struct {
	Payout    TokenPayout
	GenAmount types.Amount
}

type PayoutList []TokenPayoutReGen

func (p PayoutList) Len() int { return len(p) }
func (p PayoutList) Less(i, j int) bool {
	x, _ := (&types.Amount{Value: 0, Id: types.MEERID}).Add(&p[i].GenAmount, &p[i].Payout.Amount)
	y, _ := (&types.Amount{Value: 0, Id: types.MEERID}).Add(&p[j].GenAmount, &p[j].Payout.Amount)
	return x.Value < y.Value
}
func (p PayoutList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

type PayoutList2 []TokenPayoutReGen

func (p PayoutList2) Len() int { return len(p) }
func (p PayoutList2) Less(i, j int) bool {
	return p[i].Payout.Address < p[j].Payout.Address
}
func (p PayoutList2) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// GenesisLedger specifies the list of payouts in the coinbase of
// genesis. Must be a constant fixed in the code.
// If there are no payouts to be given, set this
// to an empty slice.
var GenesisLedger []*TokenPayout

// BlockOneSubsidy returns the total subsidy of block height 1 for the
// network.
func GenesisLedgerSubsidy() types.Amount {
	zero := &types.Amount{Value: 0, Id: types.MEERID}
	if len(GenesisLedger) == 0 {
		return *zero
	}
	sum := zero
	for _, output := range GenesisLedger {
		sum.Add(sum, &output.Amount)
	}
	return *sum
}

func addPayout(addr string, amount uint64, pksStr string) {
	pks, err := hex.DecodeString(pksStr)
	if err != nil {
		fmt.Printf("Error %v - address:%s  amount:%d\n", err, addr, amount)
		return
	}
	var amt *types.Amount
	amt, err = types.NewMeer(amount)
	if err != nil {
		fmt.Printf("Error %v - address:%s  amount:%d\n", err, addr, amount)
		return
	}
	GenesisLedger = append(GenesisLedger, &TokenPayout{addr, pks, *amt})
	//fmt.Printf("Add payout (%d) - address:%s  amount:%d\n",len(GenesisLedger),addr,amount)
}

func addPayout2(addr string, amount types.Amount, pksStr string) {
	pks, err := hex.DecodeString(pksStr)
	if err != nil {
		fmt.Printf("Error %v - address:%s  amount:%d\n", err, addr, amount)
		return
	}
	//TODO input check for amout
	GenesisLedger = append(GenesisLedger, &TokenPayout{addr, pks, amount})
}

// pay out tokens to a ledger.
func Ledger(tx *types.Transaction, netType protocol.Network) {
	if len(GenesisLedger) != 0 {
		GenesisLedger = GenesisLedger[:0]
	}
	switch netType {
	case protocol.MainNet:
		initMain()
	case protocol.TestNet:
		initTest()
	case protocol.MixNet:
		initMix()
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

	if len(tx.TxOut) <= 0 {
		tx.AddTxOut(&types.TxOutput{
			Amount: types.Amount{Value: 0, Id: types.MEERID},
		})
	}
}
