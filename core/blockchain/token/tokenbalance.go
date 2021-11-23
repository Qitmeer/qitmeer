package token

import (
	"fmt"
	"github.com/Qitmeer/qng-core/core/types"
	"strings"
)

// TokenBalance specifies the token balance and the locked meer amount
type TokenBalance struct {
	Balance    int64
	LockedMeer int64
}

type TokenBalancesMap map[types.CoinID]TokenBalance

func (tbs *TokenBalancesMap) Update(update *BalanceUpdate) error {
	tokenId := update.TokenAmount.Id
	tb := TokenBalance{}

	srcTB, ok := (*tbs)[tokenId]
	if ok {
		tb = srcTB
	}
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

func (tbs *TokenBalancesMap) Updates(updates []BalanceUpdate) error {
	for _, update := range updates {
		err := tbs.Update(&update)
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

func CheckUnMintUpdate(update *BalanceUpdate) error {
	if update.Typ != types.TxTypeTokenUnmint {
		return fmt.Errorf("checkUnMintUpdate : wrong update type %v", update.Typ)
	}
	if err := checkUpdateCommon(update); err != nil {
		return err
	}
	return nil
}

func CheckMintUpdate(update *BalanceUpdate) error {
	if update.Typ != types.TxTypeTokenMint {
		return fmt.Errorf("checkUnMintUpdate : wrong update type %v", update.Typ)
	}
	if err := checkUpdateCommon(update); err != nil {
		return err
	}
	return nil
}

func checkUpdateCommon(update *BalanceUpdate) error {
	if !types.IsKnownCoinID(update.TokenAmount.Id) {
		return fmt.Errorf("checkUpdateCommon : unknown token id %v", update.TokenAmount.Id.Name())
	}
	if update.TokenAmount.Value <= 0 {
		return fmt.Errorf("checkUpdateCommon : wrong token amount : %v", update.TokenAmount.Value)
	}
	if update.MeerAmount <= 0 {
		return fmt.Errorf("checkUpdateCommon : wrong meer amount : %v", update.MeerAmount)
	}
	return nil
}
