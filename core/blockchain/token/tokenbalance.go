package token

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/types"
	"strings"
)

// TokenBalance specifies the token balance and the locked meer amount
type TokenBalance struct {
	Balance    int64
	LockedMeer int64
}

type TokenBalancesMap map[types.CoinID]TokenBalance

func (tbs *TokenBalancesMap) UpdateBalance(update *BalanceUpdate) error {
	tokenId := update.TokenAmount.Id
	tb := (*tbs)[tokenId]
	switch update.Typ {
	case types.TxTypeTokenMint:
		tb.Balance += update.TokenAmount.Value
		tb.LockedMeer += update.MeerAmount
	case types.TxTypeTokenUnmint:
		if tb.Balance-update.TokenAmount.Value < 0 {
			return fmt.Errorf("can't unmint token %v more than token balance %v", update.TokenAmount, tb)
		}
		tb.Balance -= update.TokenAmount.Value
		if tb.LockedMeer-update.MeerAmount < 0 {
			return fmt.Errorf("can't unlock %v meer more than locked meer %v", update.MeerAmount, tb)
		}
		tb.LockedMeer -= update.MeerAmount
	default:
		return fmt.Errorf("unknown balance update type %v", update.Typ)
	}
	(*tbs)[tokenId] = tb
	return nil
}

func (tbs *TokenBalancesMap) UpdatesBalance(updates []BalanceUpdate) error {
	for _, update := range updates {
		err := tbs.UpdateBalance(&update)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tb *TokenBalancesMap) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "[")
	for k, v := range *tb {
		b.WriteString(fmt.Sprintf("%v:{balance:%v,locked-meer:%v},", k.Name(), v.Balance, v.LockedMeer))
	}
	fmt.Fprintf(&b, "]")
	return b.String()
}
func (tb *TokenBalancesMap) Copy() *TokenBalancesMap {
	newTb := TokenBalancesMap{}
	for k, v := range *tb {
		newTb[k] = v
	}
	return &newTb
}
