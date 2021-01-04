package blockchain

import (
	"bytes"
	"encoding/hex"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/params"
	"testing"
)

func Test_CheckTransactionSanity(t *testing.T) {
	txStr := "0100000001a31099d3efbe1e76f5576cc8815f1841edc44ac317c5c89d061fadd277d14205ffffffffffffffff010000007841cb020000001976a914c1777151516afe2b9f59bbd1479231aa2f250d2888ac0000000000000000e914955e0114510854244712659feec6092f7169746d6565722f"
	if len(txStr)%2 != 0 {
		txStr = "0" + txStr
	}
	serializedTx, err := hex.DecodeString(txStr)
	if err != nil {
		t.Fatal(err)
	}
	var tx types.Transaction
	err = tx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		t.Fatal(err)
	}
	err = checkTransactionSanityForAllNet(&tx)
	if err != nil {
		t.Fatal(err)
	}

	// We create an attack transacton data
	attackerPkScript, err := hex.DecodeString("76a914c0f0b73c320e1fe38eb1166a57b953e509c8f93e88ac")
	if err != nil {
		panic(err)
	}
	tx.AddTxOut(&types.TxOutput{
		Amount:   types.Amount{Value: 999999999999, Id: types.MEERID},
		PkScript: attackerPkScript,
	})

	err = checkTransactionSanityForAllNet(&tx)
	if err == nil {
		t.Fatal("Successful attack")
	}
}

func checkTransactionSanityForAllNet(tx *types.Transaction) error {
	err := CheckTransactionSanity(tx, &params.TestNetParams)
	if err != nil {
		return err
	}

	err = CheckTransactionSanity(tx, &params.PrivNetParams)
	if err != nil {
		return err
	}

	err = CheckTransactionSanity(tx, &params.MixNetParams)
	if err != nil {
		return err
	}

	err = CheckTransactionSanity(tx, &params.MainNetParams)
	if err != nil {
		return err
	}
	return nil
}
